# Changelog

## v5.0.1

- Upgrade `grafana-plugin-sdk-go`

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

- Default aggregation to `auto`

## v4.1.0

- Add default granularity

## v4.0.1

- Rename Telemetry API token to Enapter API token

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

- Smart labels

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
