import {
  DataFrame,
  DataSourceInstanceSettings,
  CoreApp,
  ScopedVars,
  LegacyMetricFindQueryOptions,
  MetricFindValue,
  DataQueryRequest,
  DataQueryResponse,
} from '@grafana/data';
import { DataSourceWithBackend, getTemplateSrv } from '@grafana/runtime';
import { lastValueFrom, Observable } from 'rxjs';

import { Query, Options, DEFAULT_QUERY } from './types';
import { VariableSupport } from './variablesupport';

export class DataSource extends DataSourceWithBackend<Query, Options> {
  constructor(instanceSettings: DataSourceInstanceSettings<Options>) {
    super(instanceSettings);
    this.variables = new VariableSupport(this);
  }

  getDefaultQuery(_: CoreApp): Partial<Query> {
    return DEFAULT_QUERY;
  }

  applyTemplateVariables(query: Query, scopedVars: ScopedVars) {
    const sourceFilters = [];
    if (query.sourceFilters) {
      for (const sourceFilter of query.sourceFilters) {
        const sf = getTemplateSrv()
          .replace(sourceFilter, scopedVars)
          .split(',');
        sourceFilters.push(...sf);
      }
    }

    const destinationFilters = [];
    if (query.destinationFilters) {
      for (const destinationFilter of query.destinationFilters) {
        const df = getTemplateSrv()
          .replace(destinationFilter, scopedVars)
          .split(',');
        destinationFilters.push(...df);
      }
    }

    return {
      ...query,
      queryType: query.queryType || DEFAULT_QUERY.queryType,
      namespace: getTemplateSrv().replace(query.namespace, scopedVars),
      application: getTemplateSrv().replace(query.application, scopedVars),
      workload: getTemplateSrv().replace(query.workload, scopedVars),
      sourceFilters: sourceFilters,
      destinationFilters: destinationFilters,
    };
  }

  query(request: DataQueryRequest<Query>): Observable<DataQueryResponse> {
    return super.query(request);
  }

  async metricFindQuery(
    query: Query,
    options?: LegacyMetricFindQueryOptions,
  ): Promise<MetricFindValue[]> {
    const q = this.query({
      targets: [
        {
          ...query,
          refId: query.refId
            ? `metricsFindQuery-${query.refId}`
            : 'metricFindQuery',
        },
      ],
      range: options?.range,
    } as DataQueryRequest<Query>);

    const response = await lastValueFrom(q as Observable<DataQueryResponse>);

    if (
      response &&
      (!response.data.length || !response.data[0].fields.length)
    ) {
      return [];
    }

    return response
      ? (response.data[0] as DataFrame).fields[0].values.map((_) => ({
        text: _.toString(),
      }))
      : [];
  }

  filterQuery(query: Query): boolean {
    if (query.queryType === 'applications' && !query.namespace) {
      return false;
    }

    if (query.queryType === 'workloads' && !query.namespace) {
      return false;
    }

    if (
      query.queryType === 'filters' &&
      (!query.filterType || !query.namespace)
    ) {
      return false;
    }

    if (
      query.queryType === 'applicationgraph' &&
      (!query.namespace || !query.application)
    ) {
      return false;
    }

    if (
      query.queryType === 'workloadgraph' &&
      (!query.namespace || !query.workload)
    ) {
      return false;
    }

    if (query.queryType === 'namespacegraph' && !query.namespace) {
      return false;
    }

    return true;
  }
}
