package internal

import (
	"testing"

	"github.com/rtlnl/phoenix/models"
	"github.com/stretchr/testify/assert"
)

func TestNewBatchOperator(t *testing.T) {
	dbc, c := GetTestRedisClient()
	defer c()

	bo := NewBatchOperator(dbc, models.Model{})

	assert.NotNil(t, bo)
}
