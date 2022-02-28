# Enapter Telemetry Grafana Datasource Plugin

This repo contains a Grafana datasource plugin that helps to visualize and
analyze device telemetry.

## Installation

To install the plugin:

1. Go to the
   [Releases](https://github.com/Enapter/telemetry-grafana-datasource-plugin/releases)
   web page.
2. Download the plugin distribution (`dist.tar.gz`).
3. Unarchive and extract the `dist` dir from the downloaded file.
4. Move the extracted `dist` dir to `/var/lib/grafana/plugins/telemetry/dist`

## Development

You will need the following tools to develop the plugin:

- `make`
- `tar`
- `gzip`
- `docker`

To build the plugin distribution from source run `make`.
