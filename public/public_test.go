package public

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPublicAPI(t *testing.T) {
	port, _ := strconv.Atoi(testDBPort)
	p, err := NewPublicAPI(testDBHost, testNamespace, port, "")
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, p)
}
