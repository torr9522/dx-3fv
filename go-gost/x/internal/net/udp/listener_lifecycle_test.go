package udp

import (
	"errors"
	"net"
	"sync"
	"testing"
	"time"
)

func TestClosedConnNeverAcceptsQueuedDatagram(t *testing.T) {
	packetConn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer packetConn.Close()

	for i := 0; i < 1024; i++ {
		conn := newConn(
			packetConn,
			packetConn.LocalAddr(),
			&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 53000 + i%1000},
			1,
			true,
		)
		if err := conn.Close(); err != nil {
			t.Fatal(err)
		}
		if err := conn.WriteQueue([]byte("late datagram")); !errors.Is(err, net.ErrClosed) {
			t.Fatalf("closed connection accepted datagram at iteration %d: %v", i, err)
		}
	}
}

func TestConcurrentFirstDatagramsShareOneSession(t *testing.T) {
	packetConn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer packetConn.Close()

	listener := &listener{
		conn:   packetConn,
		cqueue: make(chan net.Conn, 128),
		config: &ListenConfig{ReadQueueSize: 128, Keepalive: true},
	}
	listener.connPool = newConnPool(time.Hour)
	defer listener.connPool.Close()

	peer := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 53000}
	start := make(chan struct{})
	results := make(chan *conn, 128)
	var workers sync.WaitGroup
	for i := 0; i < cap(results); i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			results <- listener.getConn(peer)
		}()
	}
	close(start)
	workers.Wait()
	close(results)

	var first *conn
	for result := range results {
		if result == nil {
			t.Fatal("concurrent session creation returned nil")
		}
		if first == nil {
			first = result
		} else if result != first {
			t.Fatal("same peer received more than one live UDP session")
		}
	}
	if queued := len(listener.cqueue); queued != 1 {
		t.Fatalf("accepted session count = %d, want 1", queued)
	}
}

func TestExpiredSessionCannotDeleteReplacement(t *testing.T) {
	packetConn, err := net.ListenPacket("udp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer packetConn.Close()

	pool := newConnPool(time.Hour)
	defer pool.Close()
	peer := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 53000}
	old := newConn(packetConn, packetConn.LocalAddr(), peer, 1, true)
	replacement := newConn(packetConn, packetConn.LocalAddr(), peer, 1, true)
	old.SetIdle(true)
	pool.Set("peer", old)
	if !old.expireIfIdle() {
		t.Fatal("idle session did not expire")
	}

	pool.Set("peer", replacement)
	if pool.CompareAndDelete("peer", old) {
		t.Fatal("expiration deleted a replacement session")
	}
	actual, ok := pool.Get("peer")
	if !ok || actual != replacement {
		t.Fatal("replacement session was not retained")
	}
}
