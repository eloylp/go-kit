package handler_test

import (
	"net/http"
	"testing"

	"github.com/eloylp/kit/test/check"
	"github.com/eloylp/kit/test/handler"
)

func router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexHandler)
	mux.HandleFunc("/redirect", RedirectHandler)
	return mux
}

func IndexHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("This is content"))
}

func RedirectHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Location", "example.com")
	w.WriteHeader(http.StatusPermanentRedirect)
}

func TestTester(t *testing.T) {
	cases := []handler.Case{
		{
			"Index is showing correctly", "/", http.MethodGet, nil, nil,
			[]check.Function{check.HasStatus(http.StatusOK), check.Contains("is content")},
		},
		{
			"Redirect endpoint redirects with permanent", "/redirect", http.MethodGet, nil, nil,
			[]check.Function{
				check.HasStatus(http.StatusPermanentRedirect),
				check.HasHeaders(func() http.Header {
					h := http.Header{}
					h.Add("Location", "example.com")
					return h
				}()),
			},
		},
	}
	t.Run("Running handler tests ...", handler.Tester(cases, router(), nil, nil))
}
