package core

import "context"

type EnapterAPIPort interface {
	Ready(context.Context) error
	QueryTimeseries(
		context.Context, *QueryTimeseriesRequest,
	) (*QueryTimeseriesResponse, error)
	ExecuteCommand(
		context.Context, *ExecuteCommandRequest,
	) (*ExecuteCommandResponse, error)
	GetDeviceManifest(
		context.Context, *GetDeviceManifestRequest,
	) (*GetDeviceManifestResponse, error)
}

type QueryTimeseriesRequest struct {
	User  string
	Query string
}

type QueryTimeseriesResponse struct {
	Timeseries *Timeseries
}

type ExecuteCommandRequest struct {
	User        string
	CommandName string
	CommandArgs map[string]any
	DeviceID    string
	HardwareID  string
}

type ExecuteCommandResponse struct {
	State   string
	Payload map[string]any
}

type GetDeviceManifestRequest struct {
	User     string
	DeviceID string
}

type GetDeviceManifestResponse struct {
	Manifest []byte
}
