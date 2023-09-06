# Enapter API Grafana Datasource Plugin

This repo contains a Grafana datasource plugin that helps to visualize and
analyze devices data from [Enapter
Cloud](https://handbook.enapter.com/software/cloud/cloud.html) with the help of
[Enapter HTTP API](https://developers.enapter.com/docs/reference/http/intro).

![Example dashboard.](https://raw.githubusercontent.com/Enapter/api-grafana-datasource-plugin/9e43b0860b51bf8d9842f2e14396096cc8624627/example-dashboard.png)

## Quick start

1. Proceed to [the token settings page](https://cloud.enapter.com/settings/tokens)
   in Enapter Cloud to issue a new API token if you do not have one.
2. Use your Enapter API token to run the Grafana Docker image with the plugin
   already installed:

```bash
docker run \
	--env ENAPTER_API_TOKEN=<YOUR_ENAPTER_API_TOKEN> \
	--rm \
	--interactive \
	--tty \
	--publish 3000:3000 \
	enapter/grafana-with-enapter-api-datasource-plugin:v6.0.0
```

3. Proceed to `http://127.0.0.1:3000`.
4. Edit the Enapter Telemetry panel.

## Usage

To visualize the device telemetry, you need to declare which data you want
using YAML. A basic query looks like this:

```yaml
telemetry:
  - device: YOUR_DEVICE
    attribute: YOUR_TELEMETRY
```

To get more info about the query language check out the [Enapter Developers
docs](https://developers.enapter.com/docs/tutorial/custom-dashboards/query-language).

## Installation

To use the Enapter API datasource in your existing Grafana installation you
need to extract the packaged plugin into the Grafana plugins directory.

The path to the plugin directory is defined in [the Grafana configuration
file](https://grafana.com/docs/grafana/latest/administration/configuration/#plugins).

Let us assume that the path is `/var/lib/grafana/plugins` (the default on
Linux). Then to install the plugin:

1. Go to the
   [Releases](https://github.com/Enapter/api-grafana-datasource-plugin/releases)
   web page.
2. Download the plugin distribution (`dist.tar.gz`).
3. Unarchive and extract the `dist` dir from the downloaded file.
4. Move the extracted `dist` dir to `/var/lib/grafana/plugins/enapter-api/dist`.

## Configuration

⚠️ The plugin is at the moment
[unsigned](https://grafana.com/docs/grafana/latest/administration/plugin-management/#plugin-signatures).
To be able to run the plugin you need to allow your Grafana installation to
load it despite the lack of signature. This can be accomplished in two ways:

1. [Using the config option](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#allow_loading_unsigned_plugins): `allow_loading_unsigned_plugins = enapter-api`
2. Using the env var: `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=enapter-api`

Once the plugin is installed and allowed to be loaded, a new datasource should
be created:

1. Use Grafana web UI to [create a new
   datasource](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/)
   of type `enapter-api`.
2. Make sure `Enapter API URL` field value is set to `https://api.enapter.com`
   (default).
3. Set `Enapter API token` field value to the value of your API token.
4. Save the changes.

## Development

You will need the following tools to develop the plugin:

- `docker`
- `gzip`
- `jq`
- `make`
- `tar`

To build the plugin distribution from source run:

```bash
make dist
```

To start a local Grafana instance with the plugin installed run:

```bash
make grafana-build grafana-run
```
