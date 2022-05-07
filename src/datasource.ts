import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { SlsDataSourceOptions, SlsQuery } from './types';

export class DataSource extends DataSourceWithBackend<SlsQuery, SlsDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SlsDataSourceOptions>) {
    super(instanceSettings);
  }
}
