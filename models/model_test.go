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

func TestNewModel(t *testing.T) {
	p, _ := strconv.Atoi(testDBPort)
	ac := db.NewAerospikeClient(testDBHost, testNamespace, p)

	// Test object creation
	m, err := NewModel("rtl_test", "homepage", "userID_articleID", ac)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, m.PublicationPoint, "rtl_test")
	assert.Equal(t, m.Campaign, "homepage")
	assert.Equal(t, m.SignalType, "userID_articleID")
	assert.Equal(t, m.Version.String(), initVersion)
	assert.Equal(t, m.Stage, initStage)

	// Test if model is in database
	r, err := ac.GetOne("rtl_test", "homepage")
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, r.Key, "homepage")
	assert.Equal(t, r.Bins["signal_type"].(string), "userID_articleID")
	assert.Equal(t, r.Bins["version"].(string), initVersion)
	assert.Equal(t, StageType(r.Bins["stage"].(string)), initStage)
}
