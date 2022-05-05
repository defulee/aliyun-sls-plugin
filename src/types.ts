import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface MyQuery extends DataQuery {
  queryText?: string;
  constant: number;
  withStreaming: boolean;
}

export const defaultQuery: Partial<MyQuery> = {
  constant: 6.5,
  withStreaming: false,
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
