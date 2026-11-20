package udp

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-gost/core/common/bufpool"
	"github.com/go-gost/core/logger"
)

type ListenConfig struct {
	Addr           net.Addr
	Backlog        int
	ReadQueueSize  int
	ReadBufferSize int
	TTL            time.Duration
	Keepalive      bool
	Logger         logger.Logger
}
type listener struct {
	conn     net.PacketConn
	cqueue   chan net.Conn
	connPool *connPool
	closed   chan struct{}
	errChan  chan error
	config   *ListenConfig
}

func NewListener(conn net.PacketConn, cfg *ListenConfig) net.Listener {
	if cfg == nil {
		cfg = &ListenConfig{}
	}

	ln := &listener{
		conn:    conn,
		cqueue:  make(chan net.Conn, cfg.Backlog),
		closed:  make(chan struct{}),
		errChan: make(chan error, 1),
		config:  cfg,
	}
	ln.connPool = newConnPool(cfg.TTL).WithLogger(cfg.Logger)
	go ln.listenLoop()

	return ln
}

func (ln *listener) Accept() (conn net.Conn, err error) {
	select {
	case conn = <-ln.cqueue:
		return
	case <-ln.closed:
		return nil, net.ErrClosed
	case err = <-ln.errChan:
		if err == nil {
			err = net.ErrClosed
		}
		return
	}
}

func (ln *listener) listenLoop() {
	for {
		select {
		case <-ln.closed:
			return
		default:
		}

		b := bufpool.Get(ln.config.ReadBufferSize)

		n, raddr, err := ln.conn.ReadFrom(b)
		if err != nil {
			bufpool.Put(b)
			ln.errChan <- err
			close(ln.errChan)
			return
		}

		var enqueueErr error
		for attempt := 0; attempt < 2; attempt++ {
			c := ln.getConn(raddr)
			if c == nil {
				enqueueErr = errors.New("connection queue is full")
				break
			}
			enqueueErr = c.WriteQueue(b[:n])
			if !errors.Is(enqueueErr, net.ErrClosed) {
				break
			}
		}
		if enqueueErr != nil {
			bufpool.Put(b)
			ln.config.Logger.Warn("data discarded: ", enqueueErr)
		}
	}
}

func (ln *listener) Addr() net.Addr {
	if ln.config.Addr != nil {
		return ln.config.Addr
	}
	return ln.conn.LocalAddr()
}

func (ln *listener) Close() error {
	select {
	case <-ln.closed:
	default:
		close(ln.closed)
		ln.conn.Close()
		ln.connPool.Close()
	}

	return nil
}

func (ln *listener) getConn(raddr net.Addr) *conn {
	laddr := ln.Addr()
	peer := raddr
	var packet *packetAddr
	if received, ok := raddr.(*packetAddr); ok {
		packet = received
		laddr, peer = received.local, received.Addr
	}
	key := laddr.String() + "|" + peer.String()
	if packet != nil {
		key += "|" + strconv.Itoa(packet.ifIndex)
	}
	for {
		previous, loaded := ln.connPool.Get(key)
		if loaded && !previous.isClosed() {
			return previous
		}

		candidate := newConn(ln.conn, laddr, peer, ln.config.ReadQueueSize, ln.config.Keepalive)
		candidate.packet = packet
		if loaded {
			if !ln.connPool.CompareAndSwap(key, previous, candidate) {
				candidate.Close()
				continue
			}
		} else if actual, exists := ln.connPool.LoadOrStore(key, candidate); exists {
			candidate.Close()
			if actual != nil && !actual.isClosed() {
				return actual
			}
			continue
		}

		select {
		case ln.cqueue <- candidate:
			return candidate
		default:
			ln.connPool.CompareAndDelete(key, candidate)
			candidate.Close()
			ln.config.Logger.Warnf("connection queue is full, client %s discarded", raddr)
			return nil
		}
	}
}

// conn is a server side connection for UDP client peer, it implements net.Conn and net.PacketConn.
type conn struct {
	net.PacketConn
	localAddr  net.Addr
	remoteAddr net.Addr
	packet     *packetAddr
	rc         chan []byte // data receive queue
	idle       int32       // indicate the connection is idle
	closed     chan struct{}
	closeMutex sync.Mutex
	keepalive  bool
}

func newConn(c net.PacketConn, laddr, remoteAddr net.Addr, queueSize int, keepalive bool) *conn {
	return &conn{
		PacketConn: c,
		localAddr:  laddr,
		remoteAddr: remoteAddr,
		rc:         make(chan []byte, queueSize),
		closed:     make(chan struct{}),
		keepalive:  keepalive,
	}
}

func (c *conn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	select {
	case bb := <-c.rc:
		n = copy(b, bb)
		c.SetIdle(false)
		bufpool.Put(bb)

	case <-c.closed:
		err = net.ErrClosed
		return
	}

	addr = c.remoteAddr

	return
}

func (c *conn) Read(b []byte) (n int, err error) {
	n, _, err = c.ReadFrom(b)
	return
}

func (c *conn) WriteTo(b []byte, addr net.Addr) (n int, err error) {
	if !c.keepalive {
		defer c.Close()
	}
	if c.packet != nil {
		addr = &packetAddr{Addr: addr, local: c.packet.local, control: c.packet.control, ifIndex: c.packet.ifIndex}
	}
	return c.PacketConn.WriteTo(b, addr)
}

func (c *conn) Write(b []byte) (n int, err error) {
	return c.WriteTo(b, c.remoteAddr)
}

func (c *conn) Close() error {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
	return nil
}

func (c *conn) isClosed() bool {
	select {
	case <-c.closed:
		return true
	default:
		return false
	}
}

func (c *conn) LocalAddr() net.Addr {
	return c.localAddr
}

func (c *conn) RemoteAddr() net.Addr {
	return c.remoteAddr
}

func (c *conn) IsIdle() bool {
	return atomic.LoadInt32(&c.idle) > 0
}

func (c *conn) SetIdle(idle bool) {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()
	c.setIdle(idle)
}

func (c *conn) setIdle(idle bool) {
	v := int32(0)
	if idle {
		v = 1
	}
	atomic.StoreInt32(&c.idle, v)
}

func (c *conn) expireIfIdle() bool {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	select {
	case <-c.closed:
		return true
	default:
	}
	if c.IsIdle() {
		close(c.closed)
		return true
	}
	c.setIdle(true)
	return false
}

func (c *conn) WriteQueue(b []byte) error {
	c.closeMutex.Lock()
	defer c.closeMutex.Unlock()

	select {
	case <-c.closed:
		return net.ErrClosed
	default:
	}
	select {
	case c.rc <- b:
		c.setIdle(false)
		return nil
	default:
		return errors.New("recv queue is full")
	}
}
