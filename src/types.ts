import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  text: string;
}

export const defaultQuery: Partial<MyQuery> = {
  text: `telemetry:
  - device: YOUR_DEVICE
    attribute: YOUR_TELEMETRY
granularity: $__interval
aggregation: auto`,
};

/**
 * These are options configured for each DataSource instance.
 */
export interface MyDataSourceOptions extends DataSourceJsonData {
  enapterAPIURL?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface MySecureJsonData {
  enapterAPIToken?: string;
}
