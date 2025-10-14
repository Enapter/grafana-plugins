package grafana

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/hashicorp/go-hclog"
)

type Plugin struct {
	name   string
	logger hclog.Logger
}

func NewPlugin() *Plugin {
	return &Plugin{
		name: "enapter-api",
		logger: hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Debug,
			Output:     os.Stderr,
			JSONFormat: true,
		}),
	}
}

func (p *Plugin) Serve() error {
	var manageOpts datasource.ManageOpts
	return datasource.Manage(p.name, p.newDataSource, manageOpts)
}

func (p *Plugin) newDataSource(
	settings backend.DataSourceInstanceSettings,
) (instancemgmt.Instance, error) {
	return newDataSourceInstance(p.logger, settings)
}
