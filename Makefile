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

GRAFANA_VERSION ?= 8.4.4
GRAFANA_TAG = grafana/grafana:$(GRAFANA_VERSION)-enapter-telemetry-datasource-plugin

.PHONY: grafana-build
grafana-build:
	$(DOCKER_BUILD) \
		--build-arg GRAFANA_VERSION=$(GRAFANA_VERSION) \
		--target grafana \
		--tag $(GRAFANA_TAG) \
		.

GRAFANA_PORT ?= 3000

.PHONY: grafana-run
grafana-run:
	docker run \
		--rm \
		--tty \
		--env TELEMETRY_API_BASE_URL=$(TELEMETRY_API_BASE_URL) \
		--env TELEMETRY_API_TOKEN=$(TELEMETRY_API_TOKEN) \
		--interactive \
		--publish $(GRAFANA_PORT):3000 \
		$(GRAFANA_TAG)
