package db

import (
	"encoding/json"
	"github.com/rtlnl/phoenix/utils"
	"os"
	"testing"
)

import "github.com/stretchr/testify/assert"

var (
	testRedisHost = utils.GetDefault(os.Getenv("DB_HOST"),"127.0.0.1:6379")
)

func TestNewRedis(t *testing.T) {
	c, err := NewRedisClient(testRedisHost)
	if err != nil {
		t.Fail()
	}
	defer c.Close()

	assert.NotNil(t, c)
}

func TestRedisGetOne(t *testing.T) {
	c, err := NewRedisClient(testRedisHost)
	if err != nil {
		t.Fail()
	}
	defer c.Close()

	var m []map[string]string
	m = append(m, map[string]string{
		"item":  "5",
		"score": "0.2",
		"type":  "movie",
	})
	m = append(m, map[string]string{
		"item":  "6",
		"score": "0.3",
		"type":  "series",
	})

	// serialize
	enc, err := json.Marshal(m)
	if err != nil {
		t.Fail()
	}

	err = c.AddOne("hello", "world", string(enc))
	if err != nil {
		t.Fail()
	}

	values, err := c.GetOne("hello", "world")
	if err != nil {
		t.Fail()
	}

	// deserialize
	var me []map[string]string
	err = json.Unmarshal([]byte(values), &me)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 2, len(me))

	err = c.DropTable("hello")
	if err != nil {
		t.Fail()
	}
}

func TestRedisGetOneNotExists(t *testing.T) {
	c, err := NewRedisClient(testRedisHost)
	if err != nil {
		t.Fail()
	}
	defer c.Close()

	val, err := c.GetOne("i", "dont-exist")

	assert.Equal(t, "", val)
	assert.NotNil(t, err)
	assert.Equal(t, "key dont-exist not found", err.Error())
}

func TestRedisAddOne(t *testing.T) {
	c, err := NewRedisClient(testRedisHost)
	if err != nil {
		t.Fail()
	}
	defer c.Close()

	var m []map[string]string
	m = append(m, map[string]string{
		"item":  "5",
		"score": "0.2",
		"type":  "movie",
	})
	m = append(m, map[string]string{
		"item":  "6",
		"score": "0.3",
		"type":  "series",
	})

	// serialize
	enc, err := json.Marshal(m)
	if err != nil {
		t.Fail()
	}

	err = c.AddOne("hello", "world", string(enc))
	if err != nil {
		t.Fail()
	}

	// clean up
	err = c.DropTable("hello")
	if err != nil {
		t.Fail()
	}
}

func TestRedisDeleteOne(t *testing.T) {
	c, err := NewRedisClient(testRedisHost)
	if err != nil {
		t.Fail()
	}
	defer c.Close()

	var m []map[string]string
	m = append(m, map[string]string{
		"item":  "5",
		"score": "0.2",
		"type":  "movie",
	})
	m = append(m, map[string]string{
		"item":  "6",
		"score": "0.3",
		"type":  "series",
	})

	// serialize
	enc, err := json.Marshal(m)
	if err != nil {
		t.Fail()
	}

	err = c.AddOne("hello", "world", string(enc))
	if err != nil {
		t.Fail()
	}

	err = c.DeleteOne("hello", "world")
	if err != nil {
		t.Fail()
	}

	v, err := c.GetAllRecords("hello")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 0, len(v))

	// clean up
	err = c.DropTable("hello")
	if err != nil {
		t.Fail()
	}
}

func TestRedisDropTable(t *testing.T) {
	c, err := NewRedisClient(testRedisHost)
	if err != nil {
		t.Fail()
	}
	defer c.Close()

	var m []map[string]string
	m = append(m, map[string]string{
		"item":  "5",
		"score": "0.2",
		"type":  "movie",
	})
	m = append(m, map[string]string{
		"item":  "6",
		"score": "0.3",
		"type":  "series",
	})

	// serialize
	enc, err := json.Marshal(m)
	if err != nil {
		t.Fail()
	}

	err = c.AddOne("hello", "world", string(enc))
	if err != nil {
		t.Fail()
	}

	err = c.AddOne("hello", "bananas", string(enc))
	if err != nil {
		t.Fail()
	}

	err = c.DropTable("hello")
	if err != nil {
		t.Fail()
	}

	// it should be empty indeed
	val, err := c.GetOne("hello", "bananas")
	if err == nil {
		t.Fail()
	}

	assert.Equal(t, "", val)

	// it should be empty indeed
	vals, err := c.GetAllRecords("hello")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 0, len(vals))
}

func TestRedisGetAllRecords(t *testing.T) {
	c, err := NewRedisClient(testRedisHost)
	if err != nil {
		t.Fail()
	}
	defer c.Close()

	var m []map[string]string
	m = append(m, map[string]string{
		"item":  "5",
		"score": "0.2",
		"type":  "movie",
	})
	m = append(m, map[string]string{
		"item":  "6",
		"score": "0.3",
		"type":  "series",
	})

	// serialize
	enc, err := json.Marshal(m)
	if err != nil {
		t.Fail()
	}

	err = c.AddOne("hello", "world", string(enc))
	if err != nil {
		t.Fail()
	}

	err = c.AddOne("hello", "apples", string(enc))
	if err != nil {
		t.Fail()
	}

	values, err := c.GetAllRecords("hello")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, 2, len(values))
}
