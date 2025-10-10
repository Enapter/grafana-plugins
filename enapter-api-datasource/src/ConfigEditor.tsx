import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<MyDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onEnapterAPIURLChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      enapterAPIURL: event.target.value || 'https://api.enapter.com',
    };
    onOptionsChange({ ...options, jsonData });
  };

  onEnapterAPIVersionChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      enapterAPIVersion: event.target.value || 'v3',
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
            inputWidth={30}
            onChange={this.onEnapterAPIVersionChange}
            value={jsonData.enapterAPIVersion}
            placeholder="v1 or v3"
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
