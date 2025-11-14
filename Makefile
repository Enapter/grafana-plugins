default:

.PHONY: tag-release
tag-release:
	for i in $$(seq 5); do \
		git tag v$$(date +%Y.%m.%d)-$$i && exit; \
	done

export DOCKER_BUILDKIT=1
DOCKER_BUILD = docker build \
	--build-arg BUILDKIT_INLINE_CACHE=1

PLUGINS ?= $(shell find . -path './*/src/plugin.json' \
	| sed -E 's|\./(.+)/src/plugin.json|\1|')

.PHONY: $(PLUGINS)
$(PLUGINS):
	rm -rf ./$@/dist
	$(DOCKER_BUILD) \
		--output ./$@/dist \
		./$@

enapter-grafana-plugins.tar.gz: $(PLUGINS)
	rm -f $@
	tar --create --gzip --file $@ $(addsuffix /dist,$^)

GRAFANA_VERSION ?= 11.6
GRAFANA_TAG ?= enapter/grafana-plugins:$(GRAFANA_VERSION)-dev

.PHONY: grafana-build
grafana-build: $(PLUGINS)
	$(DOCKER_BUILD) \
		--build-arg GRAFANA_VERSION=$(GRAFANA_VERSION) \
		--tag $(GRAFANA_TAG) \
		.

GRAFANA_PORT ?= 3000

ifndef ENAPTER_API_TOKEN
ENAPTER_API_TOKEN = $(TELEMETRY_API_TOKEN)
endif

GRAFANA_CONTAINER ?= enapter-dashboards

DOCKER_NETWORK ?= bridge

.PHONY: grafana-run
grafana-run:
	docker run \
		--name $(GRAFANA_CONTAINER) \
		--rm \
		--tty \
		--env ENAPTER_API_URL=$(ENAPTER_API_URL) \
		--env ENAPTER_API_TOKEN=$(ENAPTER_API_TOKEN) \
		--env ENAPTER_API_VERSION=$(ENAPTER_API_VERSION) \
		--env DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN=$(DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN) \
		--env PROVISION_ENAPTER_BUILT_IN_DASHBOARDS=$(PROVISION_ENAPTER_BUILT_IN_DASHBOARDS) \
		--interactive \
		--network $(DOCKER_NETWORK) \
		--publish $(GRAFANA_PORT):3000 \
		$(GRAFANA_TAG)
