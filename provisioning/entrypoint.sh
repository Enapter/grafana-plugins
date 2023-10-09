#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

plugins_dir=/var/lib/grafana/plugins
rm -rf $plugins_dir/enapter-*
mkdir -p $plugins_dir
cp -r /opt/plugins/enapter-* $plugins_dir

ENAPTER_API_URL=${ENAPTER_API_URL:-https://api.enapter.com}
ENAPTER_API_TOKEN=${ENAPTER_API_TOKEN:-${TELEMETRY_API_TOKEN:-}}
if [ -z "$ENAPTER_API_TOKEN" ]; then
	echo "ENAPTER_API_TOKEN is either empty or missing." > /dev/stderr
	exit 1
fi

cat > /etc/grafana/provisioning/datasources/enapter.yml <<EOF
apiVersion: 1
datasources:
  - name: Enapter API
    type: enapter-api
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

export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=enapter-api,enapter-commands
export GF_LOG_LEVEL=${GF_LOG_LEVEL:-info}

exec /run.sh
