package devicesapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi/v3/devicesapi"
)

type MockServer struct {
	t                     *testing.T
	server                *httptest.Server
	getManifestHandler    http.HandlerFunc
	executeCommandHandler http.HandlerFunc
}

func StartMockServer(t *testing.T) *MockServer {
	t.Helper()

	s := new(MockServer)

	s.t = t
	s.getManifestHandler = s.unexpectedRequestHandler
	s.executeCommandHandler = s.unexpectedRequestHandler

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v3/devices/{device_id}/manifest",
		s.handleGetManifest)
	mux.HandleFunc("POST /v3/devices/{device_id}/execute_command",
		s.handleExecuteCommand)

	s.server = httptest.NewServer(mux)

	return s
}

func (s *MockServer) handleGetManifest(w http.ResponseWriter, r *http.Request) {
	s.getManifestHandler(w, r)
}

func (s *MockServer) handleExecuteCommand(w http.ResponseWriter, r *http.Request) {
	s.executeCommandHandler(w, r)
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

func (s *MockServer) ExpectGetManifestRequestAndReturnInvalidContentType() {
	s.replaceGetManifestHandler(func(w http.ResponseWriter, r *http.Request) {
		data := []byte("{}")
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(data)
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectGetManifestRequestAndReturnCode(code int, description string) {
	s.replaceGetManifestHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, err := w.Write([]byte(description))
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectGetManifestRequestCheckItAndReturnData(
	checkFn func(*http.Request), manifest string,
) {
	s.replaceGetManifestHandler(func(w http.ResponseWriter, r *http.Request) {
		checkFn(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"manifest":` + manifest + `}`))
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectGetManifestRequestAndReturnData(manifest string) {
	s.replaceGetManifestHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"manifest":` + manifest + `}`))
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectExecuteCommandRequestAndReturnInvalidContentType() {
	s.replaceExecuteCommandHandler(func(w http.ResponseWriter, r *http.Request) {
		data := []byte("{}")
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(data)
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectExecuteCommandRequestAndReturnCode(
	code int, description string,
) {
	s.replaceExecuteCommandHandler(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(code)
		_, err := w.Write([]byte(description))
		require.NoError(s.t, err)
	})
}

func (s *MockServer) ExpectExecuteCommandRequestCheckItAndReturnData(
	checkFn func(*http.Request), execution *devicesapi.CommandExecution,
) {
	s.replaceExecuteCommandHandler(func(w http.ResponseWriter, r *http.Request) {
		checkFn(r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]any{
			"execution": execution,
		})
		require.NoError(s.t, err)
	})
}

func (s *MockServer) replaceGetManifestHandler(h http.HandlerFunc) {
	s.replaceHandler(&s.getManifestHandler, h)
}

func (s *MockServer) replaceExecuteCommandHandler(h http.HandlerFunc) {
	s.replaceHandler(&s.executeCommandHandler, h)
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
