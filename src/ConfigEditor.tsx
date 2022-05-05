import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { SlsDataSourceOptions, SlsSecureJsonData } from './types';

const { SecretFormField, FormField } = LegacyForms;

interface Props extends DataSourcePluginOptionsEditorProps<SlsDataSourceOptions> {}

interface State {}

export class ConfigEditor extends PureComponent<Props, State> {
  onAccessKeyIdChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      accessKeyId: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  // Secure field (only sent to the backend)
  onAccessKeySecretChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonData: {
        accessKeySecret: event.target.value,
      },
    });
  };

  onResetAccessKeySecret = () => {
    const { onOptionsChange, options } = this.props;
    onOptionsChange({
      ...options,
      secureJsonFields: {
        ...options.secureJsonFields,
        accessKeySecret: false,
      },
      secureJsonData: {
        ...options.secureJsonData,
        accessKeySecret: '',
      },
    });
  };

  onEndpointChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      endpoint: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onProjectChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      project: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  onLogStoreChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onOptionsChange, options } = this.props;
    const jsonData = {
      ...options.jsonData,
      logStore: event.target.value,
    };
    onOptionsChange({ ...options, jsonData });
  };

  render() {
    const { options } = this.props;
    const { jsonData, secureJsonFields } = options;
    const secureJsonData = (options.secureJsonData || {}) as SlsSecureJsonData;

    return (
      <div className="gf-form-group">
        <div className="gf-form">
          <FormField
            label="AccessKeyID"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onAccessKeyIdChange}
            value={jsonData.accessKeyId || ''}
            placeholder="your ak id"
          />
        </div>

        <div className="gf-form-inline">
          <div className="gf-form">
            <SecretFormField
              isConfigured={(secureJsonFields && secureJsonFields.accessKeySecret) as boolean}
              value={secureJsonData.accessKeySecret || ''}
              label="AccessKeySecret"
              placeholder="your ak secret"
              labelWidth={6}
              inputWidth={20}
              onReset={this.onResetAccessKeySecret}
              onChange={this.onAccessKeySecretChange}
            />
          </div>
        </div>

        <div className="gf-form">
          <FormField
            label="Endpoint"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onEndpointChange}
            value={jsonData.endpoint || ''}
            placeholder="your endpoint"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="Project"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onProjectChange}
            value={jsonData.project || ''}
            placeholder="your project name"
          />
        </div>

        <div className="gf-form">
          <FormField
            label="LogStore"
            labelWidth={6}
            inputWidth={20}
            onChange={this.onLogStoreChange}
            value={jsonData.logStore || ''}
            placeholder="your LogStore name"
          />
        </div>
      </div>
    );
  }
}
