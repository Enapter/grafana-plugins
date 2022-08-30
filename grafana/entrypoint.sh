#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

provisioning_dir=/etc/grafana/provisioning
datasource_config=$provisioning_dir/datasources/telemetry.yml
plugin_type=enapter-telemetry

export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=$plugin_type
export GF_LOG_LEVEL=info
export GF_AUTH_ANONYMOUS_ENABLED=true
export GF_AUTH_ANONYMOUS_ORG_ROLE=Admin

TELEMETRY_API_BASE_URL=${TELEMETRY_API_BASE_URL:-https://api.enapter.com/telemetry}
TELEMETRY_API_TOKEN=${TELEMETRY_API_TOKEN:-$ENAPTER_API_TOKEN}

if [ -z "$TELEMETRY_API_TOKEN" ]; then
	echo "both TELEMETRY_API_TOKEN and ENAPTER_API_TOKEN are empty or missing" > /dev/stderr
	exit 1
fi

opt_plugins_dir=/opt/plugins
plugins_dir=/var/lib/grafana/plugins

rm -rf ${plugins_dir:?}/$plugin_type
mkdir -p $plugins_dir
cp -r $opt_plugins_dir/$plugin_type $plugins_dir/$plugin_type

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
    editable: false
EOF

exec /run.sh
