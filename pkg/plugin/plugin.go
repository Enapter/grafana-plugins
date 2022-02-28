package plugin

import (
	"os"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	hclog "github.com/hashicorp/go-hclog"
)

func Serve() error {
	p := &plugin{
		name: "enapter-telemetry",
		logger: hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Debug,
			Output:     os.Stderr,
			JSONFormat: true,
		}),
	}
	return p.serve()
}

type plugin struct {
	name   string
	logger hclog.Logger
}

func (p *plugin) serve() error {
	var manageOpts datasource.ManageOpts
	return datasource.Manage(p.name, p.newDataSource, manageOpts)
}

func (p *plugin) newDataSource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return newDataSource(p.logger, settings)
}
