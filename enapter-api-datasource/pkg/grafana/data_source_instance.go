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

type enapterAPIAdapter interface {
	core.EnapterAPIPort
	Close()
}

type dataSourceInstance struct {
	logger            hclog.Logger
	enapterAPIAdapter enapterAPIAdapter
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
	apiVersion := jsonData["enapterAPIVersion"]
	apiToken := settings.DecryptedSecureJSONData["enapterAPIToken"]

	var enapterAPIAdapter enapterAPIAdapter

	switch apiVersion {
	case "v1":
		enapterAPIAdapter = http.NewEnapterAPIv1Adapter(
			http.EnapterAPIv1AdapterParams{
				Logger:   logger,
				APIURL:   apiURL,
				APIToken: apiToken,
			})
	case "v3":
		enapterAPIAdapter = http.NewEnapterAPIv3Adapter(
			http.EnapterAPIv3AdapterParams{
				Logger:   logger,
				APIURL:   apiURL,
				APIToken: apiToken,
			})
	default:
		return nil, fmt.Errorf(`%w: want "v1" or "v3", have %q`,
			errUnsupportedAPIVersion, apiVersion)
	}

	dataSource := core.NewDataSource(core.DataSourceParams{
		Logger:     logger,
		EnapterAPI: enapterAPIAdapter,
	})

	logger.Info("created new data source",
		"api_url", apiURL,
		"api_version", apiVersion,
	)

	return &dataSourceInstance{
		logger:             logger,
		enapterAPIAdapter:  enapterAPIAdapter,
		QueryDataHandler:   dataSource,
		CheckHealthHandler: dataSource,
	}, nil
}

func (d *dataSourceInstance) Dispose() {
	d.enapterAPIAdapter.Close()

	d.logger.Info("disposed data source")
}
