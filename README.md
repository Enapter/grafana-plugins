# Enapter Grafana Plugins

This repo contains the following Enapter Grafana plugins:

- [Enapter API Datasource](enapter-api-datasource) — Visualize and analyze
  devices data from Enapter Cloud and Enapter Gateway.
- [Enapter Commands Panel](enapter-commands-panel) — Send commands to devices
  integrated into Enapter EMS.

## Quick start with Cloud Connection

This setup used when your local grafana setup connects to Enapter Cloud for Data visualization and commands.

1. Proceed to [the token settings
   page](https://cloud.enapter.com/settings/tokens) in Enapter Cloud to issue a
   new API token if you do not have one.
2. Use your Enapter API token to run the Grafana Docker image with plugins
   already installed:

```bash
docker run \
	--env ENAPTER_API_TOKEN=<YOUR_ENAPTER_API_TOKEN> \
	--rm \
	--interactive \
	--tty \
	--publish 3000:3000 \
	enapter/grafana-plugins
```

3. Proceed to `http://127.0.0.1:3000`.

## Quick start with Enapter Gateway 3.0 Connection

In case you are running Enapter Gateway 3.0, you can enable Grafana for creating custom dashboards which can be shown on external screens, for example TV, operator PC, etc.

To run grafana

1. Connect over SSH to your gateway under `enapter` user

```bash
ssh enapter@enapter-gateway.local
```
2. Edit `docker-compose.yaml` file

```bash
sudo nano /user/etc/docker-compose/docker-compose.yml
```
3. Uncomment lines so config looks like this

```yaml
services:
  grafana:
    user: root
    ports:
      - '0.0.0.0:3000:3000'
    env_file: /user/etc/enapter/enapter-token.env
    environment:
      - 'ENAPTER_API_URL=http://10.88.0.1/api'
      # NOTE: Making Commands Plugin support Platform V3 is on the way.
      - 'DISABLE_ENAPTER_COMMANDS_PANEL_PLUGIN=1'
    volumes:
      - /user/grafana-data:/var/lib/grafana
    image: enapter/grafana-plugins
```

4. Save changes and exit with CTRL+X shortcut
5. Restart `docker-compose`

```bash
sudo systemctl restart enapter-docker-compose
```
6. Open web browser and navigate to http://gateway_ip:3000 or [http://enapter-gateway.local:3000]
7. Use default login `admin` and password `admin` to log in to Grafana.

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
