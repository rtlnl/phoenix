package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAerospikeClient(t *testing.T) {
	s := NewAerospikeClient("127.0.0.1", "personalization", 3000)
	assert.NotNil(t, s)
}

func TestPing(t *testing.T) {
	s := NewAerospikeClient("127.0.0.1", "personalization", 3000)

	err := s.Health()
	if err != nil {
		t.Errorf("TestPing(%v) got unexpected error", err)
	}
}

func TestClose(t *testing.T) {
	s := NewAerospikeClient("127.0.0.1", "personalization", 3000)

	err := s.Close()
	if err != nil {
		t.Errorf("TestClose(%v) got unexpected error", err)
	}
}

func TestSetOne(t *testing.T) {
	s := NewAerospikeClient("127.0.0.1", "personalization", 3000)

	values := map[string]interface{}{
		"hourly":  1,
		"daily":   1,
		"weekly":  1,
		"monthly": 1,
	}

	// key --> user_id:item_id
	err := s.AddOne("personalization", "123:1", values)
	if err != nil {
		t.Errorf("TestSetOne(%v) got unexpected error", err)
	}
}

func TestGetOne(t *testing.T) {
	s := NewAerospikeClient("127.0.0.1", "personalization", 3000)

	values := map[string]interface{}{
		"hourly":  1,
		"daily":   1,
		"weekly":  1,
		"monthly": 1,
	}

	// key --> user_id:item_id
	err := s.AddOne("personalization", "123:1", values)
	if err != nil {
		t.Errorf("TestSetOne(%v) got unexpected error", err)
	}

	v, err := s.GetOne("personalization", "123:1")
	if err != nil {
		t.Errorf("TestSetOne(%v) got unexpected error", err)
	}

	rec := v.(*Record)
	for _, cap := range rec.Bins {
		if cap.(int) <= 0 {
			t.Errorf("TestSetOne() expeted greater than 0 got error instead: %v", err)
		}
	}
}
