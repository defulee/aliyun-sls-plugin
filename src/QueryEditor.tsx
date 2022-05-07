import { defaults } from 'lodash';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms, Select } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, SlsQuery, SlsDataSourceOptions, Formatter } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, SlsQuery, SlsDataSourceOptions>;

const formatOptions = [
  { label: 'TimeSeries', value: Formatter.TimeSeries, description: 'time series format' },
  { label: 'Table', value: Formatter.Table, description: 'table format' },
];

export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, query: event.target.value });
  };

  onFormatterChange = (event: SelectableValue<Formatter>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, format: event.value });
    onRunQuery();
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { query: queryText, format: format } = query;

    return (
      <div className="gf-form">

        <FormField
          labelWidth={8}
          value={queryText || ''}
          onChange={this.onQueryTextChange}
          label="Query Text"
          tooltip="Aliyun sls query text"
        />

        <div className="gf-form">
          <Select menuShouldPortal options={formatOptions} value={format} onChange={this.onFormatterChange} />
        </div>
      </div>
    );
  }
}
