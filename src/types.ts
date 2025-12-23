import { DataSourceJsonData } from '@grafana/data';
import { DataQuery } from '@grafana/schema';

export const DEFAULT_QUERIES: Record<QueryType, Partial<Query>> = {
  namespaces: {},
  applications: {
    namespace: '',
  },
  workloads: {
    namespace: '',
  },
  applicationgraph: {
    namespace: '',
    application: '',
    metrics: [
      'grpcRequests',
      'httpRequests',
      'tcpSentBytes',
      'tcpReceivedBytes',
    ],
  },
  workloadgraph: {
    namespace: '',
    workload: '',
    metrics: [
      'grpcRequests',
      'httpRequests',
      'tcpSentBytes',
      'tcpReceivedBytes',
    ],
  },
};

export const DEFAULT_QUERY: Partial<Query> = {
  queryType: 'applicationgraph',
  namespace: '',
  application: '',
  metrics: ['grpcRequests', 'httpRequests', 'tcpSentBytes', 'tcpReceivedBytes'],
};

export type QueryType =
  | 'namespaces'
  | 'applications'
  | 'workloads'
  | 'applicationgraph'
  | 'workloadgraph';

export interface Query
  extends DataQuery,
  QueryModelApplications,
  QueryModelWorkloads,
  QueryModelApplicationGraph,
  QueryModelWorkloadGraph {
  queryType: QueryType;
}

interface QueryModelApplications {
  namespace?: string;
}

interface QueryModelWorkloads {
  namespace?: string;
}

interface QueryModelApplicationGraph {
  namespace?: string;
  application?: string;
  metrics?: string[];
  idleEdges?: boolean;
}

interface QueryModelWorkloadGraph {
  namespace?: string;
  workload?: string;
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
