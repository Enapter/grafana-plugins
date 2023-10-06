ARG GRAFANA_VERSION=9.5.2

FROM grafana/grafana:${GRAFANA_VERSION}

COPY ./enapter-api-datasource/dist /opt/plugins/enapter-api-datasource/dist

COPY ./provisioning/entrypoint.sh ./opt/grafana-entrypoint.sh
COPY ./provisioning/home-dashboard.json /usr/share/grafana/public/dashboards/home.json

ENTRYPOINT ["./opt/grafana-entrypoint.sh"]
