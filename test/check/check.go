package check

import (
	"crypto/md5"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// Function represents a checker function that will
// endure certain results in handler response.
type Function func(w *http.Response, body []byte) error

// Contains check for a valid substring in the response body
// of an HTTP handler.
func Contains(want string) Function {
	return func(_ *http.Response, body []byte) error {
		if !strings.Contains(string(body), want) {
			return fmt.Errorf("contains: body does not contain %s", want)
		}
		return nil
	}
}

// MatchesMD5 checks that the response body of an HTTP handler
// is equal to the wanted MD5 sum.
func MatchesMD5(want string) Function {
	return func(_ *http.Response, body []byte) error {
		m := md5.New()
		if _, err := m.Write(body); err != nil {
			return fmt.Errorf("matchMD5: internal error, cannot write to data stream")
		}
		got := fmt.Sprintf("%x", m.Sum(nil))
		if got != want {
			return fmt.Errorf("matchMD5: wanted %q is not %q", want, got)
		}
		return nil
	}
}

// HasStatus checks for the specified HTTP status code in the
// handler response.
func HasStatus(want int) Function {
	return func(w *http.Response, body []byte) error {
		got := w.StatusCode
		if got != want {
			return fmt.Errorf("hasStatus: wanted %v is not %v ", want, got)
		}
		return nil
	}
}

// HasHeaders checks for wanted headers and will return an error
// with the non matched ones.
// This only checks that the specified headers are present with its values.
// This means that it could be more headers than the expected. In this case,
// the test will pass.
func HasHeaders(want http.Header) Function {
	return func(w *http.Response, body []byte) error {
		var sb strings.Builder
		var dirty bool
		sb.WriteString("hasHeaders: \n")
		for k := range want {
			got := w.Header.Get(k)
			want := want.Get(k)
			if got != want {
				dirty = true
				sb.WriteString(fmt.Sprintf("%s: want %q got %q \n", k, want, got))
			}
		}
		if dirty {
			return fmt.Errorf(sb.String())
		}
		return nil
	}
}

// ContainsJSON checks for equality in for a JSON given
// against the response body of an HTTP handler.
func ContainsJSON(want string) Function {
	return func(_ *http.Response, body []byte) error {
		want = removeSpacesAndTabs(want)
		got := removeSpacesAndTabs(string(body))
		if got == want {
			return nil
		}
		var sb strings.Builder
		sb.WriteString("containsJSON: wanted\n")
		sb.WriteString(fmt.Sprintf("%q", want))
		sb.WriteString("is not\n")
		sb.WriteString(fmt.Sprintf("%q", got))
		return errors.New(sb.String())
	}
}

func removeSpacesAndTabs(s string) string {
	s = strings.Replace(s, "\t", "", -1)
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "\n", "", -1)
	return s
}
