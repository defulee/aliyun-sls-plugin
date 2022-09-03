import { defaults } from 'lodash';

import React, { ChangeEvent, PureComponent } from 'react';
import { LegacyForms, Label, InlineFieldRow, InlineField, Select, CodeEditor } from '@grafana/ui';

import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, SlsQuery, SlsDataSourceOptions, Formatter } from './types';

const { FormField } = LegacyForms;

type Props = QueryEditorProps<DataSource, SlsQuery, SlsDataSourceOptions>;

const formatOptions = [
  { label: 'TimeSeries', value: Formatter.TimeSeries },
  { label: 'Table', value: Formatter.Table },
];

export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (value: string) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, queryText: value });
    onRunQuery();
  };

  onFormatChange = (event: SelectableValue<Formatter>) => {
    const { onChange, query, onRunQuery } = this.props;
    onChange({ ...query, format: event.value });
    onRunQuery();
  };

  onTimeFieldChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, timeField: event.target.value });
  };

  onTimezoneChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, timezone: event.target.value });
  };

  onTimeFormatChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, timeFormat: event.target.value });
  };

  onNumberFieldChange = (event: ChangeEvent<HTMLInputElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, numberField: event.target.value });
  };

  render() {
    const query = defaults(this.props.query, defaultQuery);

    return (
      <>
        <div style={{ width: '100%' }}>
          <Label className="width-12">Query Text:</Label>
          <CodeEditor
            language="sql"
            showLineNumbers={true}
            value={query.queryText || ''}
            width="100%"
            height="200px"
            onBlur={this.onQueryTextChange}
          />
        </div>
        <br />
        <div className="gf-form">
          <InlineFieldRow>
            <InlineField label="Format" labelWidth={8}>
              <Select options={formatOptions} onChange={this.onFormatChange} value={query.format} />
            </InlineField>
          </InlineFieldRow>

          {query.format === Formatter.TimeSeries && (
            <FormField
              labelWidth={8}
              value={query.timeField || 'time'}
              onChange={this.onTimeFieldChange}
              label="TimeField"
              tooltip="TimeField is used for time series format x-axis values. Default field name is 'time'."
              placeholder="time field name"
            />
          )}

          {query.format === Formatter.TimeSeries && (
            <FormField
              labelWidth={8}
              value={query.timezone || 'Asia/Shanghai'}
              onChange={this.onTimezoneChange}
              label="Timezone"
              tooltip="Timezone to parse time field. Default timezone is 'Asia/Shanghai'."
              placeholder="eg. Asia/Shanghai"
            />
          )}

          {query.format === Formatter.TimeSeries && (
            <FormField
              labelWidth={8}
              value={query.timeFormat || 'yyyy-MM-dd HH:mm:ss'}
              onChange={this.onTimeFormatChange}
              label="TimeFormat"
              tooltip="TimeFormat to parse time field. Default timeFormat is 'yyyy-MM-dd HH:mm:ss'."
              placeholder="eg. yyyy-MM-dd HH:mm:ss"
            />
          )}

          <FormField
            labelWidth={8}
            value={query.numberField || 'qpm'}
            onChange={this.onTimeFieldChange}
            label="NumberField"
            tooltip="NumberField is used for time series format y-axis value. Default field name is 'qpm'."
            placeholder="number field name"
          />
        </div>
      </>
    );
  }
}
