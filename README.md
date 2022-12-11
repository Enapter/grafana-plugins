# Enapter Telemetry Grafana Datasource Plugin

This repo contains a Grafana datasource plugin that helps to visualize and
analyze devices telemetry from [Enapter Cloud](https://handbook.enapter.com/software/cloud/cloud.html).

![Example dashboard.](./example-dashboard.png)

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
	enapter/grafana-with-telemetry-datasource-plugin:v4.0.1
```

3. Proceed to `http://127.0.0.1:3000`.
4. Edit the Telemetry panel.

## Usage

To visualize the device telemetry, you need to declare which data you want
using YAML. A basic query looks like this:

```yaml
telemetry:
  - device: YOUR_DEVICE
    attribute: YOUR_TELEMETRY
granularity: $__interval
aggregation: auto
```

To get more info about the query language check out the [Enapter Developers
docs](https://developers.enapter.com/docs/tutorial/custom-dashboards/query-language).

## Installation

To use the Enapter Telemetry datasource in your existing Grafana installation
you need to extract the packaged plugin into the Grafana plugins directory.

The path to the plugin directory is defined in [the Grafana configuration
file](https://grafana.com/docs/grafana/latest/administration/configuration/#plugins).

Let us assume that the path is `/var/lib/grafana/plugins` (the default). Then
to install the plugin:

1. Go to the
   [Releases](https://github.com/Enapter/telemetry-grafana-datasource-plugin/releases)
   web page.
2. Download the plugin distribution (`dist.tar.gz`).
3. Unarchive and extract the `dist` dir from the downloaded file.
4. Move the extracted `dist` dir to `/var/lib/grafana/plugins/telemetry/dist`.

## Configuration

Once the plugin is installed, a new datasource should be created:

1. Use Grafana web UI to [create a new
   datasource](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/)
   of type `telemetry`.
2. Make sure `Enapter Telemetry API base URL` field value is set to
   `https://api.enapter.com/telemetry` (default).
3. Set `Enapter API token` field value to the value of your token.
4. Save the changes.

## Development

You will need the following tools to develop the plugin:

- `make`
- `tar`
- `gzip`
- `docker`

To build the plugin distribution from source run:

```bash
make dist
```

To start a local Grafana instance with the plugin installed run:

```bash
make grafana-build grafana-run
```
