package grafana_test

import (
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/require"

	"github.com/Enapter/grafana-plugins/pkg/grafana"
)

func TestNewDataSourceInstance(t *testing.T) {
	logger := hclog.NewNullLogger()

	t.Run("should NOT fail if JSONData contains boolean", func(t *testing.T) {
		jsonData, err := json.Marshal(map[string]any{
			"enapterAPIURL":     "https://api.enapter.com",
			"enapterAPIVersion": "v3",
			"tlsSkipVerify":     true,
		})
		require.NoError(t, err)

		settings := backend.DataSourceInstanceSettings{
			JSONData: jsonData,
		}

		instance, err := grafana.NewDataSourceInstance(logger, settings)
		require.NoError(t, err)
		require.NotNil(t, instance)
	})
}
