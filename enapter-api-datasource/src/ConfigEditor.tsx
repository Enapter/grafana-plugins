import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps, SelectableValue } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from './types';

const { SecretFormField, FormField, Select } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

const apiVersions = ['v1', 'v3'] as const;
type ApiVersion = (typeof apiVersions)[number];
type ApiVersionOption = SelectableValue<ApiVersion>;

const apiVersionOptions: ApiVersionOption[] = [
  { label: 'v1', value: 'v1' },
  { label: 'v3', value: 'v3' },
];

const toApiVersionOptionOrDefault = (
  version: string | undefined,
  defaultOption: ApiVersionOption = { label: 'v3', value: 'v3' }
): ApiVersionOption => {
  if (!version) {
    return defaultOption;
  }

  return apiVersionOptions.find((o) => o.value === version) || defaultOption;
};

const toValidApiVersion = (value: string | undefined, defaultValue: ApiVersion = 'v3'): ApiVersion => {
  if (!value) {
    return defaultValue;
  }

  return apiVersions.find((v) => v === String(value).toLowerCase()) || defaultValue;
};

export class ConfigEditor extends PureComponent<Props, State> {
  componentDidMount() {
    const { onOptionsChange, options } = this.props;

    onOptionsChange({
      ...options,
      jsonData: {
        ...options.jsonData,
        enapterAPIURL: options.jsonData.enapterAPIURL || 'https://api.enapter.com',
        enapterAPIVersion: toValidApiVersion(options.jsonData.enapterAPIVersion),
      },
    });
  }

  onEnapterAPIURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      enapterAPIURL: event.target.value || 'https://api.enapter.com',
    };
    onOptionsChange({ ...options, jsonData });
  };

  onEnapterAPIVersionChange = (option: ApiVersionOption) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      enapterAPIVersion: toValidApiVersion(option.value),
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onEnapterAPITokenChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        enapterAPIToken: event.target.value,
      },
    });
  };

  onResetEnapterAPIToken = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        enapterAPIToken: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        enapterAPIToken: '',
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
            label="Enapter API URL"
            labelWidth={10}
            inputWidth={30}
            onChange={this.onEnapterAPIURLChange}
            value={jsonData.enapterAPIURL}
            placeholder="Enapter API URL."
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Enapter API version"
            labelWidth={10}
            inputEl={
              <Select
                width={30}
                options={apiVersionOptions}
                value={toApiVersionOptionOrDefault(jsonData.enapterAPIVersion)}
                onChange={this.onEnapterAPIVersionChange}
              />
            }
          />
        </div>

        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.enapterAPIToken) as boolean}
              value={secureJsonData.enapterAPIToken || ''}
              label="Enapter API token"
              placeholder="Your Enapter API token."
              labelWidth={10}
              inputWidth={30}
              onReset={this.onResetEnapterAPIToken}
              onChange={this.onEnapterAPITokenChange}
            />
          </div>
        </div>
      </div>
    );
  }
}
