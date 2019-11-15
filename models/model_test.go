package models

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

var (
	testDBHost    = utils.GetEnv("DB_HOST", "127.0.0.1")
	testDBPort    = utils.GetEnv("DB_PORT", "3000")
	testNamespace = "test"
)

func TestMain(m *testing.M) {
	c := m.Run()
	tearDown()
	os.Exit(c)
}

func tearDown() {
	ac, close := GetTestAerospikeClient()
	defer close()

	if err := ac.TruncateNamespace(testNamespace); err != nil {
		panic(err)
	}
}

func GetTestAerospikeClient() (*db.AerospikeClient, func()) {
	p, _ := strconv.Atoi(testDBPort)
	ac := db.NewAerospikeClient(testDBHost, testNamespace, p)

	return ac, func() { ac.Close() }
}

func TestNewModelModelExists(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	// Test object creation
	m, err := NewModel("collaborative", "", []string{"articleId"}, ac)
	defer ac.TruncateSet("collaborative")
	if err != nil {
		assert.Equal(t, "model with name 'collaborative' exists already", err.Error())
	} else {
		assert.NotNil(t, m)
	}
}

func TestModelPublish(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	m, err := NewModel("test1", "_", []string{"a", "b"}, ac)
	defer ac.TruncateSet("test1")
	if err != nil {
		t.Fail()
	}

	if err := m.PublishModel(ac); err != nil {
		t.Fail()
	}

	assert.Equal(t, true, m.IsPublished())
	assert.Equal(t, false, m.IsStaged())
}

func TestModelStaged(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	m, err := NewModel("test2", "_", []string{"a", "b"}, ac)
	defer ac.TruncateSet("test2")
	if err != nil {
		t.Fail()
	}

	if err := m.StageModel(ac); err != nil {
		t.Fail()
	}

	assert.Equal(t, true, m.IsStaged())
	assert.Equal(t, false, m.IsPublished())
}

func TestModelDelete(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	m, err := NewModel("test3", "_", []string{"a", "b"}, ac)
	defer ac.TruncateSet("test3")
	if err != nil {
		t.Fail()
	}
	if err := m.DeleteModel(ac); err != nil {
		t.Fail()
	}

	assert.Equal(t, true, m.IsDeleted())
}

func TestUpdateSignalOrder(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	m, err := NewModel("test4", "_", []string{"a", "b"}, ac)
	defer ac.TruncateSet("test4")
	if err != nil {
		t.Fail()
	}

	so := []string{"c", "d"}
	if err := m.UpdateSignalOrder(so, ac); err != nil {
		t.Fail()
	}

	assert.Equal(t, 2, len(m.SignalOrder))
	assert.EqualValues(t, so, m.SignalOrder)
}

func TestRequireSignalFormat(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	m, err := NewModel("test5", "_", []string{"a", "b", "c"}, ac)
	defer ac.TruncateSet("test5")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, true, m.RequireSignalFormat())
}

func TestGetAllModels(t *testing.T) {
	ac, close := GetTestAerospikeClient()
	defer close()

	// cleanup the database first
	if err := ac.TruncateSet(setNameAllModels); err != nil {
		t.Fail()
	}
	time.Sleep(time.Second * 2)

	_, err := NewModel("test6", "_", []string{"a", "b"}, ac)
	defer ac.TruncateSet("test6")
	if err != nil {
		t.Fail()
	}

	_, err = NewModel("test7", "_", []string{"a", "b"}, ac)
	defer ac.TruncateSet("test7")
	if err != nil {
		t.Fail()
	}

	models, err := GetAllModels(ac)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 2, len(models))
}

func TestCorrectSignalFormat(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	m, err := NewModel("test8", "_", []string{"a", "b"}, ac)
	defer ac.TruncateSet("test8")
	if err != nil {
		t.Fail()
	}

	tests := map[string]struct {
		input    string
		expected bool
	}{
		"correct": {
			input:    "11_22",
			expected: true,
		},
		"not correct 1": {
			input:    "11_33_33_33",
			expected: false,
		},
		"not correct 2": {
			input:    "11",
			expected: false,
		},
		"not correct 3": {
			input:    "11_",
			expected: false,
		},
		"not correct 4": {
			input:    "_11_",
			expected: false,
		},
		"not correct 5": {
			input:    "_11",
			expected: false,
		},
		"not correct 6": {
			input:    "_",
			expected: false,
		},
		"not correct 7": {
			input:    "11____",
			expected: false,
		},
		"not correct 8": {
			input:    "____11",
			expected: false,
		},
		"not correct 9": {
			input:    "",
			expected: false,
		},
	}
	for testName, test := range tests {
		t.Logf("Running test case %s", testName)
		o := m.CorrectSignalFormat(test.input)
		assert.Equal(t, test.expected, o)
	}
}
