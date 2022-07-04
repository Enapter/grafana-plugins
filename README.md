# Enapter Telemetry Grafana Datasource Plugin

This repo contains a Grafana datasource plugin that helps to visualize and
analyze device telemetry.

In order to use the plugin one needs to obtain an Enapter Telemetry API
token. At the moment we provide it on an individual basis. Please, contact
us at [developers@enapter.com](mailto:developers@enapter.com) to get your token.

## Quick start

1. Use your Telemetry API token to run the Grafana Docker image with the plugin
   already installed:

```bash
docker run \
	--env TELEMETRY_API_TOKEN=<YOUR_TELEMETRY_API_TOKEN> \
	--rm \
	--interactive \
	--tty \
	--publish 3000:3000 \
	enapter/grafana-with-telemetry-datasource-plugin
```

2. Proceed to `http://127.0.0.1:3000`.
3. Edit the Telemetry panel.

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

1. Use Grafana web UI to
   [create a new datasource](https://grafana.com/docs/grafana/latest/datasources/add-a-data-source/)
   of type `telemetry`.
2. Make sure `Telemetry API base URL` field value is set to `https://api.enapter.com/telemetry` (default).
3. Set `Telemetry API token` field value to the value of your token.
4. Save the changes.

## Usage

TODO: Add link to docs.

## Development

You will need the following tools to develop the plugin:

- `make`
- `tar`
- `gzip`
- `docker`

To build the plugin distribution from source run `make dist`.
