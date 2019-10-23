package public

import (
	"strconv"
	"testing"

	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/stretchr/testify/assert"
)

func TestNewPublicAPI(t *testing.T) {
	port, _ := strconv.Atoi(testDBPort)
	rl := logs.NewStdoutLog()

	p, err := NewPublicAPI(testDBHost, testNamespace, port, "", rl)
	if err != nil {
		t.Fail()
	}

	assert.NotNil(t, p)
}
