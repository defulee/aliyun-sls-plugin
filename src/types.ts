import { DataQuery, DataSourceJsonData } from '@grafana/data';

export enum Formatter {
  TimeSeries = 'TimeSeries',
  Table = 'table',
}

export interface SlsQuery extends DataQuery {
  queryText?: string;
  format?: Formatter;
}

export const defaultQuery: Partial<SlsQuery> = {
  format: Formatter.TimeSeries,
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
