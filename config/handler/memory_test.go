package handler

import (
	"encoding/json"
	"github.com/ghodss/yaml"
	"strings"
	"testing"
)

func TestMemory_Watch(t *testing.T) {
	paths, err := getPaths("../conf")
	if err != nil {
		t.Fatal(err)
	}
	m := NewMemory()
	err = m.Watch(paths...)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		if _, err := m.Get(path); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMemory_Get(t *testing.T) {
	paths, err := getPaths("../conf")
	if err != nil {
		t.Fatal(err)
	}
	newMemory := NewMemory()
	err = newMemory.Watch(paths...)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		cs, err := newMemory.Get(path)
		if err != nil {
			t.Fatal(err)
		}
		s := strings.Split(path, ".")
		suffix := s[len(s)-1]
		switch suffix {
		case "json":
			m := make(map[string]interface{})
			if err := json.Unmarshal(cs.Data, &m); err != nil {
				t.Error(err)
			}
			t.Logf("got %+#v", m)
		case "yaml":
			m := make(map[string]interface{})
			if err := yaml.Unmarshal(cs.Data, &m); err != nil {
				t.Error(err)
			}
			t.Logf("got %+#v", m)
		}
		if _, err := newMemory.Get(path); err != nil {
			t.Fatal(err)
		}
	}
}
