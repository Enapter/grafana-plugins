# Enapter API Datasource Plugin

Enapter API datasource plugin helps to visualize and analyze devices data from
[Enapter Cloud](https://handbook.enapter.com/software/cloud/cloud.html) with
the help of [Enapter HTTP
API](https://developers.enapter.com/docs/reference/http/intro).

![Example dashboard.](https://raw.githubusercontent.com/Enapter/api-grafana-datasource-plugin/9e43b0860b51bf8d9842f2e14396096cc8624627/example-dashboard.png)

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
