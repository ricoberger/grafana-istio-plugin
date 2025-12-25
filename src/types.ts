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
  filters: {
    filterType: 'source',
    namespace: '',
    application: '',
    workload: '',
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
  namespacegraph: {
    namespace: '',
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
  | 'filters'
  | 'applicationgraph'
  | 'workloadgraph'
  | 'namespacegraph';

export interface Query
  extends DataQuery,
  QueryModelApplications,
  QueryModelWorkloads,
  QueryModelFilters,
  QueryModelApplicationGraph,
  QueryModelWorkloadGraph,
  QueryModelNamespaceGraph {
  queryType: QueryType;
}

interface QueryModelApplications {
  namespace?: string;
}

interface QueryModelWorkloads {
  namespace?: string;
}

export type QueryModelFiltersFilterType = 'source' | 'destination';

interface QueryModelFilters {
  filterType?: QueryModelFiltersFilterType;
  namespace?: string;
  application?: string;
  workload?: string;
}

interface QueryModelApplicationGraph {
  namespace?: string;
  application?: string;
  metrics?: string[];
  idleEdges?: boolean;
  sourceFilters?: string[];
  destinationFilters?: string[];
}

interface QueryModelWorkloadGraph {
  namespace?: string;
  workload?: string;
  metrics?: string[];
  idleEdges?: boolean;
  sourceFilters?: string[];
  destinationFilters?: string[];
}

interface QueryModelNamespaceGraph {
  namespace?: string;
  workload?: string;
  metrics?: string[];
  idleEdges?: boolean;
  sourceFilters?: string[];
  destinationFilters?: string[];
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
