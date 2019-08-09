package models

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/rtlnl/data-personalization-api/pkg/db"
	"github.com/rtlnl/data-personalization-api/utils"
)

var (
	testDBHost    = utils.GetEnv("DB_HOST", "127.0.0.1")
	testDBPort    = utils.GetEnv("DB_PORT", "3000")
	testNamespace = "test"
)

func TestNewModelModelExists(t *testing.T) {
	p, _ := strconv.Atoi(testDBPort)
	ac := db.NewAerospikeClient(testDBHost, testNamespace, p)

	// Test object creation
	_, err := NewModel("rtl_test", "homepage", "userID_articleID", ac)
	assert.NotNil(t, err)
}
