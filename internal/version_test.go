package internal

// import (
// 	"io/ioutil"
// 	"net/http"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestLongVersion(t *testing.T) {
// 	router.GET("/", LongVersion)

// 	code, body, err := MockRequest(http.MethodGet, "/", nil)
// 	if err != nil {
// 		t.FailNow()
// 	}

// 	b, err := ioutil.ReadAll(body)
// 	if err != nil {
// 		t.FailNow()
// 	}

// 	assert.Equal(t, http.StatusOK, code)
// 	assert.Equal(t, "{\"version\":\"Internal Personalization APIs v0.0.1\"}", string(b))
// }
