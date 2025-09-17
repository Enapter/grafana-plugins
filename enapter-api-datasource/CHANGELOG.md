# Changelog

## v7.4.0

- Add @offset query modifier.

## v7.3.0

- Support Platform API v3 healthcheck.

## v7.2.0

- Enable alerting.

## v7.1.0

- Handle device manifest queries.

## v7.0.1

- Fix resetting Enapter API URL in datasource config editor.

## v7.0.0

- Rename project.
- Fix commands-api client URL usage.

## v6.0.0

- Add support for executing commands.
- Make labels unique only for a single query.

## v5.1.1

- Fix the link to example dashboard in README.
- Fix plugin name and author.
- Add plugin description.

## v5.1.0

- Extend supported platforms.
- Fix the name of executable.

## v5.0.2

- Fix timestamp formatting to avoid losing milliseconds.

## v5.0.1

- Upgrade `grafana-plugin-sdk-go`.

## v5.0.0

- Revert handling alerts by splitting array data frame into multiple boolean
  data frames. Similar functionality will be implemented by Enapter Telemetry
  API.

## v4.3.1

- Fix plugin type in UI.

## v4.3.0

- Handle alerts by splitting string array data frame into multiple boolean data
  frames.

## v4.2.0

- Default aggregation to `auto`.

## v4.1.0

- Add default granularity.

## v4.0.1

- Rename Telemetry API token to Enapter API token.

## v4.0.0

- Migrate to the new Telemetry API: set the required `Accept` header.

## v3.0.0

- Migrate to the new Telemetry API: use manifest data types names.

## v2.0.2

- Fix: update default query.

## v2.0.1

- Upgrade frontend dependencies.

## v2.0.0

- Migrate to the new Telemetry API.

## v1.2.9

- Provide Telemetry API HTTP client as library.

## v1.2.8

- Fix error message when no token is found.

## v1.2.7

- Allow other plugins to be installed.

## v1.2.6

- Smart labels.

## v1.2.5

- Handle hidden queries correctly.

## v1.2.4

- Nameless data frame.

## v1.2.3

- Upgrade frontend dependencies.

## v1.2.2

- Handle anonymous users correctly.

## v1.2.1

- Return more user-friendly error message if the metric data type is not
  supported.
- Fix nulls in multi-metric data.
- Reduce query editor font size.
- Add default query.

## v1.2.0

- Handle multiple metrics.

## v1.1.0

- Fix token support.
- Make form inputs in config editor wider.
- Configure Telemetry API base URL to point to the Cloud by default.

## v1.0.0

Initial release.
