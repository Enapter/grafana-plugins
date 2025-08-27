#!/bin/bash

set -euo pipefail
IFS=$'\n\t'

plugins_dir=/var/lib/grafana/plugins
rm -rf $plugins_dir/enapter-*
mkdir -p $plugins_dir

cp -r /opt/enapter/grafana/plugins/enapter-api-datasource $plugins_dir

DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN="${DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN:-"0"}"
case "${DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN}" in
	"0")
		cp -r /opt/enapter/grafana/plugins/enapter-commands-panel $plugins_dir
		;;
	"1")
		echo "DEBUG: Enapter Commands Panel plugin disabled." > /dev/stderr
		;;
	*)
		echo "ERROR: Unexpected value of DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN: \`${DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN}\`. Please use either 0 or 1." > /dev/stderr
		exit 1
esac

ENAPTER_API_URL=${ENAPTER_API_URL:-https://api.enapter.com}
ENAPTER_API_TOKEN=${ENAPTER_API_TOKEN:-${TELEMETRY_API_TOKEN:-}}
if [ -z "$ENAPTER_API_TOKEN" ]; then
	echo "ERROR: ENAPTER_API_TOKEN is either empty or missing." > /dev/stderr
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

dashboards_dir=/var/lib/grafana/dashboards
rm -rf $dashboards_dir/enapter-*
mkdir -p $dashboards_dir

PROVISION_ENAPTER_VUCM_DASHBOARD="${PROVISION_ENAPTER_VUCM_DASHBOARD:-"0"}"
case "${PROVISION_ENAPTER_VUCM_DASHBOARD}" in
	"0")
		echo "DEBUG: Skip provisioning Enapter VUCM dashboard." > /dev/stderr
		;;
	"1")
		cp /opt/enapter/grafana/dashboards/enapter-vucm-dashboard.json $dashboards_dir
		;;
	*)
		echo "ERROR: Unexpected value of PROVISION_ENAPTER_VUCM_DASHBOARD: \`${PROVISION_ENAPTER_VUCM_DASHBOARD}\`. Please use either 0 or 1." > /dev/stderr
		exit 1
esac

export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=enapter-api,enapter-commands
export GF_LOG_LEVEL=${GF_LOG_LEVEL:-info}
export GF_PANELS_DISABLE_SANITIZE_HTML=${GF_PANELS_DISABLE_SANITIZE_HTML:-true}

# Use the data migration CLI utility to force migration of database before
# starting the server.
grafana cli admin data-migration encrypt-datasource-passwords
# Disable the default "Getting Started" panel. See:
# https://github.com/grafana/grafana/issues/8402.
sqlite3 /var/lib/grafana/grafana.db "UPDATE user SET help_flags1 = 1 WHERE login = 'admin';"

exec /run.sh
