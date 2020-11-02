package models

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/rtlnl/phoenix/utils"
)

var (
	testDBHost     = utils.GetEnv("DB_HOST", "127.0.0.1:6379")
	testDBPassword = utils.GetEnv("DB_PASSWORD", "")
)

func TestMain(m *testing.M) {
	tearUp()
	c := m.Run()
	tearDown()
	os.Exit(c)
}

func tearUp() {
	dbc, err := db.NewRedisClient(testDBHost, db.Password(testDBPassword))
	if err != nil {
		panic(err)
	}
	if err := dbc.Client.FlushAll().Err(); err != nil {
		panic(err)
	}
}

func tearDown() {
	dbc, err := db.NewRedisClient(testDBHost, db.Password(testDBPassword))
	if err != nil {
		panic(err)
	}
	if err := dbc.Client.FlushAll().Err(); err != nil {
		panic(err)
	}
}

func GetTestRedisClient() (db.DB, func()) {
	dbc, err := db.NewRedisClient(testDBHost, db.Password(testDBPassword))
	if err != nil {
		panic(err)
	}
	return dbc, func() {
		err := dbc.Close()
		if err != nil {
			panic(err)
		}
	}
}

func TestNewModel(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	m, err := NewModel("collaborative", "", []string{"articleId"}, dbc)
	if err != nil {
		t.FailNow()
	}

	assert.NotNil(t, m)
	assert.Equal(t, "collaborative", m.Name)
	assert.Equal(t, "", m.Concatenator)
	assert.Equal(t, 1, len(m.SignalOrder))
}

func TestNewModelReservedWorld(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	// Test object creation
	_, err := NewModel("models", "", []string{"articleId"}, dbc)
	if err == nil {
		t.FailNow()
	}

	assert.Equal(t, "cannot use models as name. this name is reserved", err.Error())
}

func TestGetModel(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	model, err := GetModel("collaborative", dbc)
	if err != nil {
		t.FailNow()
	}

	assert.NotNil(t, model)
	assert.Equal(t, "collaborative", model.Name)
	assert.Equal(t, "", model.Concatenator)
	assert.Equal(t, 1, len(model.SignalOrder))
}

func TestModelDelete(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	m, err := NewModel("test3", "_", []string{"a", "b"}, dbc)
	if err != nil {
		t.FailNow()
	}
	if err := m.DeleteModel(dbc); err != nil {
		t.FailNow()
	}
}

func TestModelDeleteFromContainer(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	m, err := NewModel("test4", "_", []string{"a", "b"}, dbc)
	if err != nil {
		t.FailNow()
	}
	_, err = NewContainer("pp", "cmp", []string{"test4"}, dbc)
	if err != nil {
		t.FailNow()
	}

	if err := m.DeleteModel(dbc); err != nil {
		t.FailNow()
	}

	cont, err := GetContainer("pp", "cmp", dbc)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 0, len(cont.Models))
}

func TestUpdateSignalOrder(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	m, err := NewModel("test4", "_", []string{"a", "b"}, dbc)
	if err != nil {
		t.FailNow()
	}

	so := []string{"c", "d"}
	if err := m.UpdateSignalOrder(so, dbc); err != nil {
		t.FailNow()
	}

	assert.Equal(t, 2, len(m.SignalOrder))
	assert.EqualValues(t, so, m.SignalOrder)
}

func TestRequireSignalFormat(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	m, err := NewModel("test5", "_", []string{"a", "b", "c"}, dbc)
	if err != nil {
		t.FailNow()
	}
	assert.Equal(t, true, m.RequireSignalFormat())
}

func TestGetAllModels(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	// drop all tables first
	err := dbc.DropTable(tableModels)
	if err != nil {
		t.FailNow()
	}

	_, err = NewModel("test6", "_", []string{"a", "b"}, dbc)
	if err != nil {
		t.FailNow()
	}

	_, err = NewModel("test7", "_", []string{"a", "b"}, dbc)
	if err != nil {
		t.FailNow()
	}

	models, count, err := GetAllModels(dbc)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, 2, len(models))
	assert.Equal(t, 2, count)
}

func TestCorrectSignalFormat(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	m, err := NewModel("test8", "_", []string{"a", "b"}, dbc)
	if err != nil {
		t.FailNow()
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

func TestDeserializeModel(t *testing.T) {
	ser := `{"name":"test","signalOrder":["article","signal"],"concatenator":"_"}`
	m, err := DeserializeModel(ser)
	if err != nil {
		t.FailNow()
	}
	assert.NotNil(t, m)
	assert.Equal(t, "test", m.Name)
	assert.Equal(t, "_", m.Concatenator)
	assert.Equal(t, 2, len(m.SignalOrder))
}
