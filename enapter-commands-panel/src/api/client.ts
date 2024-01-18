import { getBackendSrv } from '@grafana/runtime';
import { BlueprintManifest } from 'types/manifest-schema';
import { DataSourceSettings } from '@grafana/data';

type BackendResponse<T> = {
  results: {
    A: {
      frames: [
        {
          data: {
            values: [T, Array<{ code: string; message: string }>];
          };
        }
      ];
      error?: any;
    };
  };
};

export type CommandResponse = BackendResponse<[state: string, payload: Record<string, any>]>;

export class GrafanaApiError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'GrafanaApiError';
  }
}

export class ApiClient {
  private readonly host: string;

  constructor(host = '') {
    this.host = `${host}`;
  }

  async fetchDatasources() {
    return this.server.get<DataSourceSettings[]>(`${this.host}/api/datasources`);
  }

  async fetchCommands(deviceId: string, datasourceUid: string) {
    const response = await this.server
      .post<BackendResponse<[BlueprintManifest]>>(
        `${this.host}/api/ds/query`,
        this.wrapRequestBody({
          queryType: 'manifest',
          payload: { deviceId },
          datasource: { uid: datasourceUid },
        })
      )
      .catch((e) => {
        this.handleFetchError(e);
      });

    const grafanaBackendResponse = response.results.A;
    const grafanaApiError = grafanaBackendResponse.error;

    if (grafanaApiError) {
      throw new GrafanaApiError(grafanaApiError);
    }

    const telemetryBackendResponse = grafanaBackendResponse.frames[0].data;
    const manifest = telemetryBackendResponse.values[0][0];
    const errors = telemetryBackendResponse.values[1] || [];

    return { commands: manifest.commands, errors };
  }

  async runCommand<T>(body: T) {
    const response = await this.server
      .post<CommandResponse>(`${this.host}/api/ds/query`, this.wrapRequestBody(body))
      .catch((e) => {
        this.handleFetchError(e);
      });

    const grafanaBackendResponse = response.results.A;
    const grafanaApiError = grafanaBackendResponse.error;

    if (grafanaApiError) {
      throw new GrafanaApiError(grafanaApiError);
    }

    const telemetryBackendResponse = grafanaBackendResponse.frames[0].data;
    const state = telemetryBackendResponse.values[0]?.[0];
    const payload = telemetryBackendResponse.values[1]?.[0] || {};
    const errors = state === 'succeeded' ? [] : telemetryBackendResponse.values[1] || [];

    return { state, payload, errors };
  }

  private get server() {
    return getBackendSrv();
  }

  private wrapRequestBody<T>(body: T): { queries: [{ refId: string } & T] } {
    return {
      queries: [
        {
          refId: 'A',
          ...body,
        },
      ],
    };
  }

  private handleFetchError(e: any): never {
    console.error(e);

    if (e.data?.error) {
      throw new GrafanaApiError(e.data.error.message);
    }

    if (e.data?.results?.A?.error) {
      throw new GrafanaApiError(e.data.results.A.error);
    }

    throw new GrafanaApiError(e.statusText);
  }
}
