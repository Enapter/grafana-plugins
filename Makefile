default:

.PHONY: dist.tar.gz
dist.tar.gz: dist
	rm --force $@
	tar --create --gzip --file $@ $<

export DOCKER_BUILDKIT=1
DOCKER_BUILD = docker build \
	--build-arg BUILDKIT_INLINE_CACHE=1

.PHONY: dist
dist:
	rm --recursive --force $@
	$(DOCKER_BUILD) \
		--target $@ \
		--output $@ \
		.

PLUGIN_VERSION = $(shell jq -r .version package.json)
GRAFANA_TAG = enapter/grafana-with-enapter-api-datasource-plugin:v$(PLUGIN_VERSION)-dev

.PHONY: grafana-build
grafana-build:
	$(DOCKER_BUILD) \
		--target grafana \
		--tag $(GRAFANA_TAG) \
		.

GRAFANA_PORT ?= 3000

ifndef ENAPTER_API_TOKEN
ENAPTER_API_TOKEN = $(TELEMETRY_API_TOKEN)
endif

.PHONY: grafana-run
grafana-run:
	docker run \
		--rm \
		--tty \
		--env ENAPTER_API_URL=$(ENAPTER_API_URL) \
		--env ENAPTER_API_TOKEN=$(ENAPTER_API_TOKEN) \
		--interactive \
		--publish $(GRAFANA_PORT):3000 \
		$(GRAFANA_TAG)
