# Enapter Grafana Plugins

This repo contains the following Enapter Grafana plugins:

- [Enapter API Datasource](enapter-api-datasource) — Visualize and analyze
  devices data from Enapter Cloud and Enapter Gateway.
- [Enapter Commands Panel](enapter-commands-panel) — Send commands to devices
  integrated into Enapter EMS.

## Quick start

1. Proceed to [the token settings
   page](https://cloud.enapter.com/settings/tokens) in Enapter Cloud to issue a
   new API token if you do not have one.
2. Use your Enapter API token to run the Grafana Docker image with plugins
   already installed:

```bash
docker run \
	--env ENAPTER_API_TOKEN=<YOUR_ENAPTER_API_TOKEN> \
	--env ENAPTER_API_VERSION=v1 \
	--rm \
	--interactive \
	--tty \
	--publish 3000:3000 \
	enapter/grafana-plugins
```

3. Proceed to `http://127.0.0.1:3000`.

## Installation

To use the Enapter Grafana plugins in your existing Grafana installation you
need to extract the packaged plugins into the Grafana plugins directory.

The path to the plugin directory is defined in [the Grafana configuration
file](https://grafana.com/docs/grafana/latest/administration/configuration/#plugins).

Let us assume that the path is `/var/lib/grafana/plugins` (the default on
Linux). Then to install plugins:

1. Go to the [Releases](https://github.com/Enapter/grafana-plugins/releases)
   web page.
2. Download the packaged plugins (`enapter-grafana-plugins.tar.gz`).
3. Unarchive and extract all the directories from the downloaded file.
4. Move all the extracted directories into `/var/lib/grafana/plugins`.
5. Make sure the files are owned by `grafana` user.

## Development

You will need the following tools to develop plugins:

- `docker`
- `make`

To start a local Grafana instance with plugins installed run:

```bash
make grafana-build grafana-run
```
