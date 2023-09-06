#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

provisioning_dir=/etc/grafana/provisioning
datasource_config=$provisioning_dir/datasources/enapter-api.yml
plugin_type=enapter-api

export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=$plugin_type
export GF_LOG_LEVEL=info

ENAPTER_API_URL=${ENAPTER_API_URL:-https://api.enapter.com}
ENAPTER_API_TOKEN=${ENAPTER_API_TOKEN:-${TELEMETRY_API_TOKEN:-}}

if [ -z "$ENAPTER_API_TOKEN" ]; then
	echo "ENAPTER_API_TOKEN is either empty or missing." > /dev/stderr
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
  - name: Enapter API
    type: $plugin_type
    access: proxy
    orgId: 1
    isDefault: true
    jsonData:
      enapterAPIURL: "$ENAPTER_API_URL"
    secureJsonData:
      enapterAPIToken: "$ENAPTER_API_TOKEN"
    version: 1
    editable: false
EOF

exec /run.sh
