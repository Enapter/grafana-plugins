#!/bin/bash

set -e

config_dir=/etc/grafana/provisioning/datasources
filename=$config_dir/telemetry.yml
plugin_type=enapter-telemetry

export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=$plugin_type
export GF_LOG_LEVEL=info
export GF_USERS_ALLOW_SIGN_UP=true
export GF_USERS_AUTO_ASSIGN_ORG_ROLE=Admin

TELEMETRY_API_BASE_URL=${TELEMETRY_API_BASE_URL:-https://api.enapter.com/telemetry}

if [ -z "$TELEMETRY_API_TOKEN" ]; then
	echo "TELEMETRY_API_TOKEN is empty or missing" > /dev/stderr
	exit
fi

cat > $filename <<EOF
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
