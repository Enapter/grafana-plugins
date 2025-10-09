package telemetryapi_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockServer struct {
	t                 *testing.T
	server            *httptest.Server
	timeseriesHandler http.HandlerFunc
}

func StartMockServer(t *testing.T) *MockServer {
	t.Helper()

	s := new(MockServer)

	s.t = t
	s.timeseriesHandler = s.unexpectedRequestHandler
	s.server = httptest.NewServer(s)

	return s
}

func (s *MockServer) Stop() {
	s.server.Close()
}

func (s *MockServer) Address() string {
	return s.server.URL
}

func (s *MockServer) NewClient() *http.Client {
	return s.server.Client()
}

func (s *MockServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/v3/telemetry/query_timeseries":
		switch r.Method {
		case http.MethodPost:
			s.timeseriesHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	default:
		http.NotFound(w, r)
	}
}

func (s *MockServer) ExpectTimeseriesRequestAndReturnZeroContentLength() {
	s.replaceTimeseriesHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
	})
}

func (s *MockServer) ExpectTimeseriesRequestAndReturnInvalidContentType() {
	s.replaceTimeseriesHandler(func(w http.ResponseWriter, r *http.Request) {
		data := []byte("{}")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(data)
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectTimeseriesRequestAndReturnCode(code int, description string) {
	s.replaceTimeseriesHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, err := w.Write([]byte(description))
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectTimeseriesRequestCheckItAndReturnData(
	checkFn func(*http.Request), types string, data string,
) {
	s.replaceTimeseriesHandler(func(w http.ResponseWriter, r *http.Request) {
		checkFn(r)
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("X-Enapter-Timeseries-Data-Types", types)
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(data))
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectTimeseriesRequestAndReturnData(types []string, data string) {
	s.replaceTimeseriesHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		for _, t := range types {
			w.Header().Add("X-Enapter-Timeseries-Data-Types", t)
		}
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(data))
		require.NoError(s.t, err)
	})
}

func (s *MockServer) replaceTimeseriesHandler(h http.HandlerFunc) {
	s.replaceHandler(&s.timeseriesHandler, h)
}

func (s *MockServer) replaceHandler(p *http.HandlerFunc, h http.HandlerFunc) {
	old := *p
	*p = func(w http.ResponseWriter, r *http.Request) {
		defer func() { *p = old }()
		h(w, r)
	}
}

func (s *MockServer) unexpectedRequestHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "unexpected request", http.StatusExpectationFailed)
}
