package plugin

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/hashicorp/go-hclog"

	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/commandsapi"
	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/plugin/internal/handlers"
	"github.com/Enapter/telemetry-grafana-datasource-plugin/pkg/telemetryapi"
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

	apiToken := settings.DecryptedSecureJSONData["telemetryAPIToken"]

	telemetryAPIClient, err := telemetryapi.NewClient(telemetryapi.ClientParams{
		BaseURL: jsonData["telemetryAPIBaseURL"],
		Token:   apiToken,
	})
	if err != nil {
		return nil, fmt.Errorf("new telemetry API client: %w", err)
	}

	commandsAPIClient := commandsapi.NewClient(commandsapi.ClientParams{
		Token: apiToken,
	})

	queryDataHandler := handlers.NewQueryData(
		logger, telemetryAPIClient, commandsAPIClient)
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
