package scheduler

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/juandiii/jetson-monitor/logging"
)

type NotifiedEntry struct {
	URL         string `json:"url"`
	LastUpdated int64  `json:"lastUpdated"`
}

type StateStore interface {
	IsCoolingDown(key string, cooldown time.Duration) bool
	Touch(key string) error
	Clear(key string) error
	All() []NotifiedEntry
}

var globalLocks sync.Map

func lockFor(path string) *sync.Mutex {
	v, _ := globalLocks.LoadOrStore(path, &sync.Mutex{})
	return v.(*sync.Mutex)
}

type FileStateStore struct {
	path string
	mu   *sync.Mutex
	data map[string]int64
	log  logging.Logger
}

func NewFileStateStore(path string, log logging.Logger) *FileStateStore {
	clean := filepath.Clean(path)
	s := &FileStateStore{
		path: clean,
		mu:   lockFor(clean),
		data: make(map[string]int64),
		log:  log,
	}
	s.load()
	return s
}

func (s *FileStateStore) load() {
	f, err := os.Open(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		if s.log != nil {
			s.log.Errorf("open state file: %v", err)
		}
		return
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	var arr []NotifiedEntry
	if err := dec.Decode(&arr); err == nil && arr != nil {
		for _, r := range arr {
			if r.URL == "" || r.LastUpdated <= 0 {
				continue
			}
			s.data[r.URL] = r.LastUpdated
		}
		return
	}
}

func (s *FileStateStore) persistLocked() {
	dir := filepath.Dir(s.path)
	base := filepath.Base(s.path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		if s.log != nil {
			s.log.Errorf("mkdir state dir: %v", err)
		}
		return
	}

	records := make([]NotifiedEntry, 0, len(s.data))
	for u, ts := range s.data {
		records = append(records, NotifiedEntry{URL: u, LastUpdated: ts})
	}
	sort.Slice(records, func(i, j int) bool { return records[i].URL < records[j].URL })

	f, err := os.CreateTemp(dir, base+".*.tmp")
	if err != nil {
		if s.log != nil {
			s.log.Errorf("create temp state: %v", err)
		}
		return
	}
	tmp := f.Name()

	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(records); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		if s.log != nil {
			s.log.Errorf("encode state: %v", err)
		}
		return
	}

	_ = f.Sync()
	_ = f.Close()

	_ = os.Remove(s.path)
	if err := os.Rename(tmp, s.path); err != nil {
		_ = os.Remove(tmp)
		if s.log != nil {
			s.log.Errorf("rename state file: %v", err)
		}
		return
	}

	if df, err := os.Open(dir); err == nil {
		_ = df.Sync()
		_ = df.Close()
	}
}

func (s *FileStateStore) Touch(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[url] = time.Now().Unix()
	s.persistLocked()
	return nil
}

func (s *FileStateStore) Clear(url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, url)
	s.persistLocked()
	return nil
}

func (s *FileStateStore) IsCoolingDown(url string, cooldown time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	ts := s.data[url]
	if ts == 0 {
		return false
	}
	return time.Since(time.Unix(ts, 0)) < cooldown
}

func (s *FileStateStore) All() []NotifiedEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]NotifiedEntry, 0, len(s.data))
	for u, ts := range s.data {
		out = append(out, NotifiedEntry{URL: u, LastUpdated: ts})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URL < out[j].URL })
	return out
}
