package devicesapi_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/suite"

	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi"
	"github.com/Enapter/grafana-plugins/pkg/http/enapterapi/v3/devicesapi"
)

type ClientSuite struct {
	suite.Suite
	ctx    context.Context
	token  string
	server *MockServer
	client *devicesapi.Client
}

func (s *ClientSuite) SetupTest() {
	s.ctx = context.Background()
	s.token = faker.Word()
	s.server = StartMockServer(s.T())
	client := devicesapi.NewClient(devicesapi.ClientParams{
		HTTPClient: s.server.NewClient(),
		BaseURL:    s.server.Address() + "/v3/devices",
		Token:      s.token,
	})
	s.client = client
}

func (s *ClientSuite) TearDownTest() {
	s.client.Close()
	s.server.Stop()
}

func (s *ClientSuite) TestReadyOK() {
	const errorJSON = `{"errors":[{"message":"Oops."}]}`
	s.server.ExpectGetManifestRequestAndReturnCode(
		http.StatusNotFound, errorJSON)
	err := s.client.Ready(s.ctx)
	s.Require().NoError(err)
}

func (s *ClientSuite) TestReadyNoMultiError() {
	const errorJSON = `{"error_message":"Oops."}`
	s.server.ExpectGetManifestRequestAndReturnCode(
		http.StatusNotFound, errorJSON)
	err := s.client.Ready(s.ctx)
	s.Require().Error(err)
	s.Require().Equal(
		`process response: multi-error: <not available>: empty error list`,
		err.Error())
}

func (s *ClientSuite) TestReadyUnexpectedAbsenseOfError() {
	s.server.ExpectGetManifestRequestAndReturnData(`{}`)
	err := s.client.Ready(s.ctx)
	s.Require().Error(err)
	s.Require().Equal(
		"unexpected absence of error",
		err.Error())
}

func (s *ClientSuite) TestGetManifestInvalidContentType() {
	s.server.ExpectGetManifestRequestAndReturnInvalidContentType()
	manifest, err := s.client.GetManifest(s.ctx, s.randomGetManifestParams())
	s.Require().Error(err)
	s.Require().Nil(manifest)
	s.Require().Equal(
		"process response: unexpected content type: "+
			"want application/json, have application/yaml",
		err.Error())
}

func (s *ClientSuite) TestGetManifest() {
	params := s.randomGetManifestParams()
	expectedManifest := `{"telemetry":{},"properties":{"foo":"bar"}}`
	s.server.ExpectGetManifestRequestCheckItAndReturnData(func(r *http.Request) {
		s.Require().Equal([]string{params.User}, r.Header["X-Enapter-Auth-User"])
		s.Require().Equal([]string{s.token}, r.Header["X-Enapter-Auth-Token"])
		s.Require().Equal(params.DeviceID, r.PathValue("device_id"))
	}, expectedManifest)
	manifest, err := s.client.GetManifest(s.ctx, params)
	s.Require().NoError(err)
	s.Require().Equal(expectedManifest, string(manifest))
}

func (s *ClientSuite) TestExecuteCommandInvalidContentType() {
	s.server.ExpectExecuteCommandRequestAndReturnInvalidContentType()
	manifest, err := s.client.ExecuteCommand(s.ctx, s.randomExecuteCommandParams())
	s.Require().Error(err)
	s.Require().Nil(manifest)
	s.Require().Equal(
		"process response: unexpected content type: "+
			"want application/json, have text/html",
		err.Error())
}

func (s *ClientSuite) TestExecuteCommandError() {
	const errorJSON = `{"errors":[{"message":"Oops."}]}`
	s.server.ExpectExecuteCommandRequestAndReturnCode(
		http.StatusBadRequest, errorJSON)
	_, err := s.client.ExecuteCommand(s.ctx, s.randomExecuteCommandParams())
	multiErr := new(enapterapi.MultiError)
	s.Require().ErrorAs(err, &multiErr)
	s.Require().Equal("Oops.", multiErr.Errors[0].Message)
}

func (s *ClientSuite) TestExecuteCommand() {
	params := s.randomExecuteCommandParams()
	expectedExecution := &devicesapi.CommandExecution{
		State: "SUCCESS",
		Response: devicesapi.CommandResponse{
			State:   "SUCCEEDED",
			Payload: map[string]any{},
		},
	}
	s.server.ExpectExecuteCommandRequestCheckItAndReturnData(func(r *http.Request) {
		s.Require().Equal([]string{params.User}, r.Header["X-Enapter-Auth-User"])
		s.Require().Equal([]string{s.token}, r.Header["X-Enapter-Auth-Token"])
		s.Require().Equal(params.DeviceID, r.PathValue("device_id"))
	}, expectedExecution)
	execution, err := s.client.ExecuteCommand(s.ctx, params)
	s.Require().NoError(err)
	s.Require().Equal(expectedExecution, execution)
}

func (s *ClientSuite) randomGetManifestParams() devicesapi.GetManifestParams {
	return devicesapi.GetManifestParams{
		User:     faker.Word(),
		DeviceID: faker.UUIDHyphenated(),
	}
}

func (s *ClientSuite) randomExecuteCommandParams() devicesapi.ExecuteCommandParams {
	return devicesapi.ExecuteCommandParams{
		User:     faker.Word(),
		DeviceID: faker.UUIDHyphenated(),
		Request: devicesapi.CommandRequest{
			Name: faker.Word(),
			Arguments: map[string]any{
				faker.Word(): faker.Sentence(),
			},
		},
	}
}

func TestClient(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ClientSuite))
}
