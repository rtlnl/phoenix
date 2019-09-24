package internal

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInternalAPI(t *testing.T) {
	p, _ := strconv.Atoi(testDBPort)
	i, err := NewInternalAPI(testDBHost, testNamespace, testRegion, testEndpoint, testDisableSSL, p)
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, i)
}
