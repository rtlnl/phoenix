package db

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testDBHost    = "127.0.0.1"
	testDBPort    = 3000
	testNamespace = "test"
)

func TestNewAerospikeClient(t *testing.T) {
	s := NewAerospikeClient(testDBHost, testNamespace, testDBPort)
	assert.NotNil(t, s)
}

func TestPing(t *testing.T) {
	s := NewAerospikeClient(testDBHost, testNamespace, testDBPort)

	err := s.Health()
	if err != nil {
		t.Errorf("TestPing(%v) got unexpected error", err)
	}
}

func TestClose(t *testing.T) {
	s := NewAerospikeClient(testDBHost, testNamespace, testDBPort)

	err := s.Close()
	if err != nil {
		t.Errorf("TestClose(%v) got unexpected error", err)
	}
}

func TestSetOne(t *testing.T) {
	s := NewAerospikeClient(testDBHost, testNamespace, testDBPort)

	err := s.AddOne(testNamespace, "key", "value")
	if err != nil {
		t.Errorf("TestSetOne(%v) got unexpected error", err)
	}
}

func TestGetOne(t *testing.T) {
	s := NewAerospikeClient(testDBHost, testNamespace, testDBPort)

	// key --> value
	err := s.AddOne(testNamespace, "key", "value")
	if err != nil {
		t.Errorf("TestGetOne(%v) got unexpected error", err)
	}

	rec, err := s.GetOne(testNamespace, "key")
	if err != nil {
		t.Errorf("TestGetOne(%v) got unexpected error", err)
	}

	for _, value := range rec.Bins {
		if value.(string) != "value" {
			t.Errorf("TestGetOne() expeted greater than 0 got error instead: %v", err)
		}
	}
}

func TestAddMultipleRecords(t *testing.T) {
	// write into Aerospike first
	ac := NewAerospikeClient(testDBHost, testNamespace, testDBPort)

	// clean up previous tests
	if err := ac.TruncateSet("model_1"); err != nil {
		t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
	}

	if err := ac.TruncateSet("model_2"); err != nil {
		t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
	}

	testSetName := "model_1"
	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("key_%d", i)
		v := fmt.Sprintf("value_%d", i)

		err := ac.AddOne(testSetName, k, v)
		if err != nil {
			t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
		}
	}

	// read all
	records, err := ac.GetAllRecords(testSetName)
	if err != nil {
		t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
	}

	if err := ac.AddMultipleRecords("model_2", records); err != nil {
		t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
	}

	storedRecords, err := ac.GetAllRecords("model_2")
	if err != nil {
		t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
	}

	counter := 0
	for range storedRecords.Results() {
		counter++
	}

	if counter != 10 {
		t.Errorf("AddMultipleRecords number doesn't match")
	}
}
