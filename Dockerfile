ARG GRAFANA_VERSION=11.6.2

FROM grafana/grafana:${GRAFANA_VERSION}

# Install `sqlite` to hack the DB to remove the default getting started panel
# in the entrypoint script.
USER root
RUN apk update && apk add sqlite
USER grafana

# Install the plugins.
COPY ./enapter-api-datasource/dist /opt/enapter/grafana/plugins/enapter-api-datasource/dist
COPY ./enapter-commands-panel/dist /opt/enapter/grafana/plugins/enapter-commands-panel/dist

# Install a custom home dashboard.
COPY ./grafana/dashboards/home.json /usr/share/grafana/public/dashboards/home.json

# Prepare optional dashboards to be provisioned by the entrypoint script.
COPY ./grafana/dashboards/provisioning /opt/enapter/grafana/dashboards/provisioning
COPY ./grafana/dashboards/lib /opt/enapter/grafana/dashboards/lib

# Install the entrypoint script.
COPY ./grafana/entrypoint.sh /opt/enapter/grafana/entrypoint.sh

ENTRYPOINT ["/opt/enapter/grafana/entrypoint.sh"]
