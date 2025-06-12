package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/grafana-plugins/pkg/assetsapi"
	"github.com/Enapter/grafana-plugins/pkg/commandsapi"
	"github.com/Enapter/grafana-plugins/pkg/plugin/internal/handlers"
	"github.com/Enapter/grafana-plugins/pkg/telemetryapi"
)

var _ instancemgmt.InstanceDisposer = (*dataSource)(nil)

type dataSource struct {
	logger             hclog.Logger
	telemetryAPIClient telemetryapi.Client

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

	telemetryAPIClient := telemetryapi.NewClient(telemetryapi.ClientParams{
		BaseURL: apiURL + "/telemetry",
		Token:   apiToken,
	})

	commandsAPIClient := commandsapi.NewClient(commandsapi.ClientParams{
		APIURL: apiURL,
		Token:  apiToken,
	})

	assetsAPIClient := assetsapi.NewClient(assetsapi.ClientParams{
		APIURL: apiURL,
		Token:  apiToken,
	})

	queryDataHandler := handlers.NewQueryData(
		logger, telemetryAPIClient, commandsAPIClient, assetsAPIClient)
	checkHealthHandler := handlers.NewCheckHealth(logger, telemetryAPIClient)

	logger.Info("created new data source")

	return &dataSource{
		logger:             logger,
		telemetryAPIClient: telemetryAPIClient,
		QueryDataHandler:   queryDataHandler,
		CheckHealthHandler: checkHealthHandler,
	}, nil
}

func (d *dataSource) Dispose() {
	d.telemetryAPIClient.Close()

	d.logger.Info("disposed data source")
}
