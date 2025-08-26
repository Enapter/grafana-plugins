ARG GRAFANA_VERSION=11.6.2

FROM grafana/grafana:${GRAFANA_VERSION}

USER root
RUN apk update && apk add sqlite
USER grafana

COPY ./enapter-api-datasource/dist /opt/enapter/grafana/plugins/enapter-api-datasource/dist
COPY ./enapter-commands-panel/dist /opt/enapter/grafana/plugins/enapter-commands-panel/dist

COPY ./provisioning/entrypoint.sh /opt/enapter/grafana/entrypoint.sh
COPY ./provisioning/home-dashboard.json /usr/share/grafana/public/dashboards/home.json

ENTRYPOINT ["/opt/enapter/grafana/entrypoint.sh"]
