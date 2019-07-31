package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDBHost       = "127.0.0.1"
	testDBPort       = 3000
	defaultNamespace = "personalization"
)

func TestNewAerospikeClient(t *testing.T) {
	s := NewAerospikeClient(testDBHost, defaultNamespace, testDBPort)
	assert.NotNil(t, s)
}

func TestPing(t *testing.T) {
	s := NewAerospikeClient(testDBHost, defaultNamespace, testDBPort)

	err := s.Health()
	if err != nil {
		t.Errorf("TestPing(%v) got unexpected error", err)
	}
}

func TestClose(t *testing.T) {
	s := NewAerospikeClient(testDBHost, defaultNamespace, testDBPort)

	err := s.Close()
	if err != nil {
		t.Errorf("TestClose(%v) got unexpected error", err)
	}
}

func TestSetOne(t *testing.T) {
	s := NewAerospikeClient(testDBHost, defaultNamespace, testDBPort)

	values := map[string]interface{}{
		"extra": 1,
	}

	// key --> user_id:item_id
	err := s.AddOne(defaultNamespace, "123:1", values)
	if err != nil {
		t.Errorf("TestSetOne(%v) got unexpected error", err)
	}
}

func TestGetOne(t *testing.T) {
	s := NewAerospikeClient(testDBHost, defaultNamespace, testDBPort)

	values := map[string]interface{}{
		"extra": 1,
	}

	// key --> user_id:item_id
	err := s.AddOne(defaultNamespace, "123:1", values)
	if err != nil {
		t.Errorf("TestGetOne(%v) got unexpected error", err)
	}

	v, err := s.GetOne(defaultNamespace, "123:1")
	if err != nil {
		t.Errorf("TestGetOne(%v) got unexpected error", err)
	}

	rec := v.(*Record)
	for _, cap := range rec.Bins {
		if cap.(int) <= 0 {
			t.Errorf("TestGetOne() expeted greater than 0 got error instead: %v", err)
		}
	}
}
