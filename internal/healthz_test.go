package internal

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthz(t *testing.T) {
	router.GET("/v1/healthz", Healthz)

	code, body, err := MockRequest(http.MethodGet, "/v1/healthz", nil)
	if err != nil {
		t.FailNow()
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.FailNow()
	}

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "{\"message\":\"I'm healthy\"}", string(b))
}
