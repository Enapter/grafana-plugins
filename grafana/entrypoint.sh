#!/bin/bash

set -e

provisioning_dir=/etc/grafana/provisioning
datasource_config=$provisioning_dir/datasources/telemetry.yml
plugin_type=enapter-telemetry

export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=$plugin_type
export GF_LOG_LEVEL=info
export GF_AUTH_ANONYMOUS_ENABLED=true
export GF_AUTH_ANONYMOUS_ORG_ROLE=Admin

TELEMETRY_API_BASE_URL=${TELEMETRY_API_BASE_URL:-https://api.enapter.com/telemetry}

if [ -z "$TELEMETRY_API_TOKEN" ]; then
	echo "TELEMETRY_API_TOKEN is empty or missing" > /dev/stderr
	exit
fi

cat > $datasource_config <<EOF
apiVersion: 1
datasources:
  - name: Enapter Telemetry
    type: $plugin_type
    access: proxy
    orgId: 1
    isDefault: true
    jsonData:
      telemetryAPIBaseURL: "$TELEMETRY_API_BASE_URL"
    secureJsonData:
      telemetryAPIToken: "$TELEMETRY_API_TOKEN"
    version: 1
    editable: true
EOF

exec /run.sh
