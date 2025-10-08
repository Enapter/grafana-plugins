package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/pkg/http"
	"github.com/Enapter/grafana-plugins/pkg/plugin/internal/handlers"
)

var _ instancemgmt.InstanceDisposer = (*dataSource)(nil)

type dataSource struct {
	logger              hclog.Logger
	enapterAPIv1Adapter *http.EnapterAPIv1Adapter

	backend.QueryDataHandler
	backend.CheckHealthHandler
}

func newDataSource(logger hclog.Logger, settings backend.DataSourceInstanceSettings,
) (inst instancemgmt.Instance, retErr error) {
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

	queryDataHandler := handlers.NewQueryData(logger, enapterAPIv1Adapter)
	checkHealthHandler := handlers.NewCheckHealth(logger, enapterAPIv1Adapter)

	logger.Info("created new data source")

	return &dataSource{
		logger:              logger,
		enapterAPIv1Adapter: enapterAPIv1Adapter,
		QueryDataHandler:    queryDataHandler,
		CheckHealthHandler:  checkHealthHandler,
	}, nil
}

func (d *dataSource) Dispose() {
	d.enapterAPIv1Adapter.Close()

	d.logger.Info("disposed data source")
}
