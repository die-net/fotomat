package thumbnail

// +build go1.6

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
	"time"
)

// Go versions prior to 1.6 have a racy net/http/httptest.(*Server).Close()
// that makes this fail the race detector.  See
// https://go-review.googlesource.com/#/c/15151/ for more detail.

func TestProxyTimeout(t *testing.T) {
	ps := newProxyServer(time.Second, time.Nanosecond)
	defer ps.close()

	body, status := ps.get("timeout")
	assert.Equal(t, http.StatusGatewayTimeout, status, string(body))
}
