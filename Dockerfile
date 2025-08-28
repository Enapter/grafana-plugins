ARG GRAFANA_VERSION=11.6.2

FROM grafana/grafana:${GRAFANA_VERSION}

USER root
RUN apk update && apk add sqlite
USER grafana

COPY ./enapter-api-datasource/dist /opt/enapter/grafana/plugins/enapter-api-datasource/dist
COPY ./enapter-commands-panel/dist /opt/enapter/grafana/plugins/enapter-commands-panel/dist

COPY ./grafana/home-dashboard.json /usr/share/grafana/public/dashboards/home.json
COPY ./grafana/default-dashboard-provider.yml /etc/grafana/provisioning/dashboards/default-dashboard-provider.yml
COPY ./grafana/vucm-dashboard.json /opt/enapter/grafana/dashboards/enapter-vucm-dashboard.json

COPY ./grafana/entrypoint.sh /opt/enapter/grafana/entrypoint.sh

ENTRYPOINT ["/opt/enapter/grafana/entrypoint.sh"]
