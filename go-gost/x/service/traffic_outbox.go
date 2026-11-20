package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-gost/x/config"
)

type trafficOutbox struct {
	ReportID    string              `json:"reportId"`
	Items       []TrafficReportItem `json:"items"`
	CreatedTime int64               `json:"createdTime"`
}

type trafficOutboxQueue struct {
	Batches []trafficOutbox `json:"batches"`
}

func trafficOutboxPath() string {
	configPath := strings.TrimSpace(config.PersistPath())
	if configPath == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(configPath), "gost-traffic-outbox.json")
}

func newTrafficReportID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func newTrafficOutbox(items []TrafficReportItem) (*trafficOutbox, error) {
	path := trafficOutboxPath()
	if path == "" {
		return nil, nil
	}
	id, err := newTrafficReportID()
	if err != nil {
		return nil, err
	}
	outbox := &trafficOutbox{
		ReportID:    id,
		Items:       append([]TrafficReportItem(nil), items...),
		CreatedTime: time.Now().UnixMilli(),
	}
	queue, err := loadTrafficOutboxQueue()
	if err != nil {
		return nil, err
	}
	queue.Batches = append(queue.Batches, *outbox)
	if err := saveTrafficOutboxQueue(queue); err != nil {
		return nil, err
	}
	return outbox, nil
}

func loadTrafficOutboxQueue() (*trafficOutboxQueue, error) {
	path := trafficOutboxPath()
	if path == "" {
		return &trafficOutboxQueue{}, nil
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &trafficOutboxQueue{}, nil
	}
	if err != nil {
		return nil, err
	}
	var queue trafficOutboxQueue
	if err := json.Unmarshal(data, &queue); err != nil {
		return nil, err
	}
	for _, outbox := range queue.Batches {
		if strings.TrimSpace(outbox.ReportID) == "" || len(outbox.Items) == 0 {
			return nil, errors.New("invalid traffic outbox")
		}
	}
	return &queue, nil
}

func saveTrafficOutboxQueue(queue *trafficOutboxQueue) error {
	path := trafficOutboxPath()
	if path == "" || queue == nil {
		return nil
	}
	if len(queue.Batches) == 0 {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	data, err := json.Marshal(queue)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".traffic-outbox-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func loadTrafficOutbox() (*trafficOutbox, error) {
	queue, err := loadTrafficOutboxQueue()
	if err != nil || len(queue.Batches) == 0 {
		return nil, err
	}
	outbox := queue.Batches[0]
	return &outbox, nil
}

func removeTrafficOutbox(reportID string) error {
	path := trafficOutboxPath()
	if path == "" {
		return nil
	}
	queue, err := loadTrafficOutboxQueue()
	if err != nil {
		return err
	}
	if len(queue.Batches) == 0 {
		return nil
	}
	if queue.Batches[0].ReportID != strings.TrimSpace(reportID) {
		return errors.New("traffic outbox report ID changed")
	}
	queue.Batches = queue.Batches[1:]
	return saveTrafficOutboxQueue(queue)
}
