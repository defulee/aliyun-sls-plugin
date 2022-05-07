import { defaults } from 'lodash';

import React, { PureComponent } from 'react';
import { Label, InlineFieldRow, InlineField, Select, CodeEditor } from '@grafana/ui';

import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, SlsQuery, SlsDataSourceOptions, Formatter } from './types';

// const { FormField } = LegacyForms;

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

  render() {
    const query = defaults(this.props.query, defaultQuery);

    return (
      <>
        <div style={{ width: '100%' }}>
          <Label className="width-12">
            Query Text:
          </Label>

          <CodeEditor
            language="json"
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
            <InlineField label="Format" grow>
              <Select options={formatOptions} onChange={this.onFormatChange} value={query.format} />
            </InlineField>
          </InlineFieldRow>
        </div>
      </>
    );
  }
}
