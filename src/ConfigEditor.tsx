import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onTelemetryAPIBaseURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      telemetryAPIBaseURL: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onTelemetryAPITokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        telemetryAPIToken: event.target.value,
      },
    });
  };

  onResetTelemetryAPIToken = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        telemetryAPIToken: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        telemetryAPIToken: '',
      },
    });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as MySecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="Telemetry API base URL"
            labelWidth={20}
            inputWidth={40}
            onChange={this.onTelemetryAPIBaseURLChange}
            value={jsonData.telemetryAPIBaseURL || 'https://api.enapter.com/telemetry'}
            placeholder="Enapter Telemetry API base URL."
          />
        </div>

        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.telemetryAPIToken) as boolean}
              value={secureJsonData.telemetryAPIToken || ''}
              label="Telemetry API token"
              placeholder="Your Enapter Telemetry API token."
              labelWidth={20}
              inputWidth={40}
              onReset={this.onResetTelemetryAPIToken}
              onChange={this.onTelemetryAPITokenChange}
            />
          </div>
        </div>
      </div>
    );
  }
}
