package internal

import (
	"testing"

	"github.com/rtlnl/data-personalization-api/models"
	"github.com/stretchr/testify/assert"
)

func TestNewBatchOperator(t *testing.T) {
	// get aerospike client
	ac, c := GetTestAerospikeClient()
	defer c()

	bo := NewBatchOperator(ac, &models.Model{})

	assert.NotNil(t, bo)
}
