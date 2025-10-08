package grafana

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/pkg/core"
	"github.com/Enapter/grafana-plugins/pkg/http"
)

var _ instancemgmt.InstanceDisposer = (*dataSourceInstance)(nil)

type dataSourceInstance struct {
	logger              hclog.Logger
	enapterAPIv1Adapter *http.EnapterAPIv1Adapter
	backend.QueryDataHandler
	backend.CheckHealthHandler
}

func newDataSourceInstance(
	logger hclog.Logger, settings backend.DataSourceInstanceSettings,
) (_ *dataSourceInstance, retErr error) {
	logger = logger.Named(fmt.Sprintf("data_source[%q]", settings.Name))

	defer func() {
		if retErr != nil {
			logger.Error("failed to create data source",
				"error", retErr.Error())
		}
	}()

	var jsonData map[string]string
	if err := json.Unmarshal(settings.JSONData, &jsonData); err != nil {
		return nil, fmt.Errorf("JSON data: %w", err)
	}

	apiURL := jsonData["enapterAPIURL"]
	apiToken := settings.DecryptedSecureJSONData["enapterAPIToken"]

	enapterAPIv1Adapter := http.NewEnapterAPIv1Adapter(http.EnapterAPIv1AdapterParams{
		Logger:   logger,
		APIURL:   apiURL,
		APIToken: apiToken,
	})

	dataSource := core.NewDataSource(core.DataSourceParams{
		Logger:     logger,
		EnapterAPI: enapterAPIv1Adapter,
	})

	logger.Info("created new data source")

	return &dataSourceInstance{
		logger:              logger,
		enapterAPIv1Adapter: enapterAPIv1Adapter,
		QueryDataHandler:    dataSource,
		CheckHealthHandler:  dataSource,
	}, nil
}

func (d *dataSourceInstance) Dispose() {
	d.enapterAPIv1Adapter.Close()

	d.logger.Info("disposed data source")
}
