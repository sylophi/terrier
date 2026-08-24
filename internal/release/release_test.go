package release

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

// A bare "HTTP 403" reads as something being broken, when the answer is to
// run the command again later.
func TestRateLimitIsExplained(t *testing.T) {
	reset := time.Now().Add(20 * time.Minute)
	resp := &http.Response{StatusCode: http.StatusForbidden, Header: http.Header{}}
	resp.Header.Set("X-RateLimit-Remaining", "0")
	resp.Header.Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

	err := statusError(resp)
	if !strings.Contains(err.Error(), "rate limiting") {
		t.Errorf("got %q, want it to name the rate limit", err)
	}
	if !strings.Contains(err.Error(), reset.Format("15:04")) {
		t.Errorf("got %q, want it to name when the limit resets", err)
	}
}

func TestRateLimitWithoutAResetHeader(t *testing.T) {
	resp := &http.Response{StatusCode: http.StatusForbidden, Header: http.Header{}}
	resp.Header.Set("X-RateLimit-Remaining", "0")
	if err := statusError(resp); !strings.Contains(err.Error(), "rate limiting") {
		t.Errorf("got %q, want it to name the rate limit", err)
	}
}

// A 403 that is not the rate limit, and anything else, stays as it was.
func TestOtherFailuresReportTheStatus(t *testing.T) {
	for _, code := range []int{http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
		resp := &http.Response{StatusCode: code, Header: http.Header{}}
		if err := statusError(resp); err.Error() != "HTTP "+strconv.Itoa(code) {
			t.Errorf("status %d gave %q", code, err)
		}
	}
}
