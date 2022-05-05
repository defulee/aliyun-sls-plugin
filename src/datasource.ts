import { DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { SlsDataSourceOptions, MyQuery } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, SlsDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<SlsDataSourceOptions>) {
    super(instanceSettings);
  }
}
