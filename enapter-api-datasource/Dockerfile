FROM node:22.18.0-alpine AS frontend

WORKDIR /build

COPY ./.npmrc ./.npmrc
COPY ./package.json ./package.json
COPY ./package-lock.json ./package-lock.json
RUN --mount=type=cache,target=node_modules npm ci

COPY ./src ./src
COPY ./.config ./.config
COPY ./.eslintrc ./.eslintrc
COPY ./jest.config.js ./jest.config.js
COPY ./jest-setup.js ./jest-setup.js
COPY ./tsconfig.json ./tsconfig.json
RUN --mount=type=cache,target=node_modules npm run test:ci

COPY ./.cprc.json ./.cprc.json
COPY ./README.md ./README.md
COPY ./CHANGELOG.md ./CHANGELOG.md
COPY ./LICENSE ./LICENSE
RUN --mount=type=cache,target=node_modules npm run build


FROM golang:1.20-alpine AS backend

RUN apk add --no-cache --virtual .build-deps \
    binutils-gold \
	git \
	build-base \
	&& go install github.com/magefile/mage@v1.12.1 \
	&& go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.2

WORKDIR /build

COPY ./go.mod ./go.mod
COPY ./go.sum ./go.sum

RUN go mod download

COPY ./pkg ./pkg

RUN --mount=type=cache,target=/root/.cache/golangci golangci-lint run ./pkg/...
RUN --mount=type=cache,target=/root/.cache/go-build go test -v ./pkg/...

COPY ./src/plugin.json ./src/plugin.json
COPY ./Magefile.go ./Magefile.go
RUN --mount=type=cache,target=/root/.cache/go-build mage \
	build:darwin \
	build:darwinARM64 \
	build:linux \
	build:linuxARM \
	build:linuxARM64 \
	build:windows


FROM scratch

COPY --from=frontend /build/dist /

COPY --from=backend /build/dist/gpx_enapter_api_darwin_amd64 /
COPY --from=backend /build/dist/gpx_enapter_api_darwin_arm64 /
COPY --from=backend /build/dist/gpx_enapter_api_linux_amd64 /
COPY --from=backend /build/dist/gpx_enapter_api_linux_arm /
COPY --from=backend /build/dist/gpx_enapter_api_linux_arm64 /
COPY --from=backend /build/dist/gpx_enapter_api_windows_amd64.exe /
