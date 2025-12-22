import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export const DEFAULT_QUERY: Partial<Query> = {
  queryType: 'graph',
  namespace: '',
  application: '',
  metrics: ['grpcRequests', 'httpRequests', 'tcpSentBytes', 'tcpReceivedBytes'],
};

export type QueryType = 'namespaces' | 'applications' | 'graph';

export interface Query
  extends DataQuery,
  QueryModelApplications,
  QueryModelGraph {
  queryType: QueryType;
}

interface QueryModelApplications {
  namespace?: string;
}

interface QueryModelGraph {
  namespace?: string;
  application?: string;
  metrics?: string[];
  idleEdges?: boolean;
}

export type OptionsPrometheusAuthMethod = 'none' | 'basic' | 'token';

export interface Options extends DataSourceJsonData {
  prometheusUrl?: string;
  prometheusAuthMethod?: OptionsPrometheusAuthMethod;
  prometheusUsername?: string;
  istioWarningThreshold?: number;
  istioErrorThreshold?: number;
  istioWorkloadDashboard?: string;
  istioServiceDashboard?: string;
}

export interface OptionsSecure {
  prometheusPassword?: string;
  prometheusToken?: string;
}
