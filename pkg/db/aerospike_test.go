package db

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/rtlnl/phoenix/utils"
	"github.com/stretchr/testify/assert"
)

var (
	testDBHost    = utils.GetEnv("DB_HOST", "127.0.0.1")
	testDBPort    = utils.GetEnv("DB_PORT", "3000")
	testNamespace = "test"
	testSetName   = "testSetName"
)

func createAerospikeClient() *AerospikeClient {
	p, _ := strconv.Atoi(testDBPort)
	ac := NewAerospikeClient(testDBHost, testNamespace, p)
	ac.TruncateSet(testSetName)

	return ac
}

func TestNewAerospikeClient(t *testing.T) {
	ac := createAerospikeClient()
	assert.NotNil(t, ac)
}

func TestPing(t *testing.T) {
	ac := createAerospikeClient()

	err := ac.Health()
	if err != nil {
		t.Errorf("TestPing(%v) got unexpected error", err)
	}
}

func TestClose(t *testing.T) {
	ac := createAerospikeClient()

	err := ac.Close()
	if err != nil {
		t.Errorf("TestClose(%v) got unexpected error", err)
	}
}

func TestSetOne(t *testing.T) {
	ac := createAerospikeClient()

	err := ac.PutOne(testSetName, "key", "bin_key", "bin_value")
	if err != nil {
		t.Errorf("TestSetOne(%v) got unexpected error", err)
	}

	err = ac.DeleteOne("key", "bin_key")
	if err != nil {
		t.Errorf("TestSetOne(%v) got unexpected error", err)
	}
}

func TestGetOne(t *testing.T) {
	ac := createAerospikeClient()

	// key --> bin_key:bin_value
	err := ac.PutOne(testSetName, "key", "bin_key", "bin_value")
	if err != nil {
		t.Errorf("TestGetOne(%v) got unexpected error", err)
	}

	rec, err := ac.GetOne(testSetName, "key")
	if err != nil {
		t.Errorf("TestGetOne(%v) got unexpected error", err)
	}

	for bk, bv := range rec.Bins {
		if bk != "bin_key" {
			t.Errorf("TestGetOne() expeted %s got error instead: %v", "bin_key", err)
		}

		if bv.(string) != "bin_value" {
			t.Errorf("TestGetOne() expeted %s got error instead: %v", "bin_value", err)
		}
	}
}

func TestAddMultipleRecords(t *testing.T) {
	// write into Aerospike first
	ac := createAerospikeClient()

	// clean up previous tests
	if err := ac.TruncateSet("model_1"); err != nil {
		t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
	}

	if err := ac.TruncateSet("model_2"); err != nil {
		t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
	}

	tsn := "model_1"
	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("key_%d", i)

		bk := fmt.Sprintf("bin_key_%d", i)
		bv := fmt.Sprintf("bin_value_%d", i)

		err := ac.PutOne(tsn, k, bk, bv)
		if err != nil {
			t.Errorf("AddMultipleRecords(%v) got unexpected error", err)
		}
	}

	// read all
	records, err := ac.GetAllRecords(tsn)
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
