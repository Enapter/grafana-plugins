ARG GRAFANA_VERSION=9.0.2

FROM node:14-alpine AS frontend

WORKDIR /build

COPY ./package.json ./package.json
COPY ./yarn.lock ./yarn.lock

RUN --mount=type=cache,target=node_modules yarn install --pure-lockfile

COPY ./src ./src
RUN --mount=type=cache,target=node_modules yarn test

COPY ./README.md ./README.md
COPY ./CHANGELOG.md ./CHANGELOG.md
COPY ./LICENSE ./LICENSE
RUN --mount=type=cache,target=node_modules yarn build


FROM golang:1.17-alpine AS backend

RUN apk add --no-cache --virtual .build-deps \
    git \
    build-base \
    && go install github.com/magefile/mage@v1.12.1 \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.43.0

WORKDIR /build

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go mod download

COPY ./pkg ./pkg

RUN --mount=type=cache,target=/root/.cache/golangci golangci-lint run ./pkg/...
RUN --mount=type=cache,target=/root/.cache/go-build go test -v ./pkg/...

COPY ./src/plugin.json ./src/plugin.json
COPY ./Magefile.go ./Magefile.go
RUN --mount=type=cache,target=/root/.cache/go-build mage build:backend


FROM scratch AS dist

COPY --from=frontend /build/dist /
COPY --from=backend /build/dist/gpx_telemetry_linux_amd64 /


FROM grafana/grafana:${GRAFANA_VERSION} AS grafana

COPY --from=dist / /opt/plugins/enapter-telemetry/dist

COPY ./grafana/entrypoint.sh ./opt/grafana-entrypoint.sh
COPY ./grafana/home-dashboard.json /usr/share/grafana/public/dashboards/home.json

ENTRYPOINT ["./opt/grafana-entrypoint.sh"]
