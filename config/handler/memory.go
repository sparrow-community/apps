package handler

import (
	"errors"
	"fmt"
	"go-micro.dev/v4/config/source"
	"go-micro.dev/v4/config/source/file"
	"strings"
	"sync"
	"time"
)

type Memory struct {
	sync.RWMutex
	exit    chan bool
	sources map[string]*fileSource
}

type fileSource struct {
	source source.Source
	set    *source.ChangeSet
}

func (m *Memory) watch(path string, fs *fileSource) {
	watch := func(path string, w source.Watcher) error {
		for {
			cs, err := w.Next()
			if err != nil {
				return err
			}
			m.Lock()
			m.sources[path].set = cs
			m.Unlock()
		}
	}

	for {
		w, err := fs.source.Watch()
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		done := make(chan bool)
		go func() {
			select {
			case <-done:
			case <-m.exit:
			}
			_ = w.Stop()
		}()

		if err := watch(path, w); err != nil {
			time.Sleep(time.Second)
		}

		close(done)
	}
}

func (m *Memory) Watch(paths ...string) error {
	var errs []string

	for _, path := range paths {
		s := file.NewSource(
			file.WithPath(path),
			source.WithEncoder(&BytesEncoder{}),
		)
		if _, ok := m.sources[path]; ok {
			continue
		}
		set, err := s.Read()
		if err != nil {
			errs = append(errs, fmt.Sprintf("error loading s %s: %v", s, err))
			continue
		}
		fs := &fileSource{source: s, set: set}
		m.Lock()
		m.sources[path] = fs
		m.Unlock()
		go m.watch(path, fs)
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func (m *Memory) Get(path string) (*source.ChangeSet, error) {
	if fs, ok := m.sources[path]; ok {
		return fs.set, nil
	}
	return nil, errors.New(fmt.Sprintf("not wartch %s", path))
}

// BytesEncoder .
type BytesEncoder struct{}

func (b BytesEncoder) Encode(_ interface{}) ([]byte, error) {
	panic("not support for encoding")
}

func (b BytesEncoder) Decode(_ []byte, _ interface{}) error {
	panic("not support for decoding")
}

func (b BytesEncoder) String() string {
	return "bytes"
}

func NewMemory() *Memory {
	return &Memory{
		sources: map[string]*fileSource{},
	}
}
