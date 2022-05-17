import { DataQuery, DataSourceJsonData } from '@grafana/data';

export enum Formatter {
  TimeSeries = 'TimeSeries',
  Table = 'Table',
}

export interface SlsQuery extends DataQuery {
  queryText?: string;
  format?: Formatter;
  timeField?: string;
  timezone?: string;
  timeFormat?: string;
}

export const defaultQuery: Partial<SlsQuery> = {
  format: Formatter.TimeSeries,
  timeField: 'time',
  timezone: 'Asia/Shanghai',
  timeFormat: 'yyyy-MM-dd HH:mm:ss',
};

/**
 * These are options configured for each DataSource instance.
 */
export interface SlsDataSourceOptions extends DataSourceJsonData {
  accessKeyId?: string;
  endpoint?: string;
  project?: string;
  logStore?: string;
}

/**
 * Value that is used in the backend, but never sent over HTTP to the frontend
 */
export interface SlsSecureJsonData {
  accessKeySecret?: string;
}
