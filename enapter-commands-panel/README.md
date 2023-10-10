# Enapter Commands Panel Plugin

Enapter Commands panel plugin allows you to create buttons which send commands
to devices integrated into Enapter EMS.

## Usage

To add a button to your dashboard create a panel of type Enapter Commands.
Follow the UI in the panel options editor to define which command to which
device your button sends.

## Configuration

⚠️ The plugin is at the moment
[unsigned](https://grafana.com/docs/grafana/latest/administration/plugin-management/#plugin-signatures).
To be able to run the plugin you need to allow your Grafana installation to
load it despite the lack of signature. This can be accomplished in two ways:

1. [Using the config option](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#allow_loading_unsigned_plugins): `allow_loading_unsigned_plugins = enapter-commands`
2. Using the env var: `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=enapter-commands`
