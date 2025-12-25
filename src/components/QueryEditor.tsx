import React, { ChangeEvent, useState } from 'react';
import {
  Collapse,
  Combobox,
  ComboboxOption,
  InlineField,
  InlineFieldRow,
  InlineSwitch,
  MultiCombobox,
} from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';

import { DataSource } from '../datasource';
import { DEFAULT_QUERIES, Options, Query, QueryType } from '../types';
import { NamespaceField } from './NamespaceField';
import { ApplicationField } from './ApplicationField';
import { WorkloadField } from './WorkloadField';
import { FiltersField } from './FiltersField';

type Props = QueryEditorProps<DataSource, Query, Options>;

export function QueryEditor({
  datasource,
  query,
  range,
  onChange,
  onRunQuery,
}: Props) {
  const [graphOptionsIsOpen, setGraphOptionsIsOpen] = useState(false);

  return (
    <>
      <InlineFieldRow>
        <InlineField label="Graph Type" labelWidth={25}>
          <Combobox<QueryType>
            value={query.queryType}
            options={[
              { label: 'Application Graph', value: 'applicationgraph' },
              { label: 'Workload Graph', value: 'workloadgraph' },
              { label: 'Namespace Graph', value: 'namespacegraph' },
            ]}
            onChange={(option: ComboboxOption<QueryType>) => {
              onChange({
                ...query,
                ...DEFAULT_QUERIES[option.value],
                queryType: option.value,
                namespace: query.namespace,
                metrics: query.metrics,
                idleEdges: query.idleEdges,
              });
              onRunQuery();
            }}
          />
        </InlineField>
      </InlineFieldRow>

      <InlineFieldRow>
        <NamespaceField
          datasource={datasource}
          range={range}
          namespace={query.namespace}
          onNamespaceChange={(namespace) => {
            onChange({
              ...query,
              namespace: namespace,
              application: '',
              workload: '',
            });
            onRunQuery();
          }}
        />

        {query.queryType === 'applicationgraph' && (
          <ApplicationField
            datasource={datasource}
            range={range}
            namespace={query.namespace}
            application={query.application}
            onApplicationChange={(application) => {
              onChange({ ...query, application: application });
              onRunQuery();
            }}
          />
        )}

        {query.queryType === 'workloadgraph' && (
          <WorkloadField
            datasource={datasource}
            range={range}
            namespace={query.namespace}
            workload={query.workload}
            onWorkloadChange={(workload) => {
              onChange({ ...query, workload: workload });
              onRunQuery();
            }}
          />
        )}
      </InlineFieldRow>

      <Collapse
        label="Graph Options"
        isOpen={graphOptionsIsOpen}
        onToggle={() => setGraphOptionsIsOpen(!graphOptionsIsOpen)}
      >
        <InlineFieldRow>
          <InlineField label="Metrics" labelWidth={25}>
            <MultiCombobox
              data-testid="metrics-combobox"
              width="auto"
              minWidth={32}
              maxWidth={32}
              isClearable={true}
              value={query.metrics}
              options={[
                { label: 'gRPC Requests', value: 'grpcRequests' },
                {
                  label: 'gRPC Request Duration',
                  value: 'grpcRequestDuration',
                },
                { label: 'gRPC Sent Messages', value: 'grpcSentMessages' },
                {
                  label: 'gRPC Received Messages',
                  value: 'grpcReceivedMessages',
                },
                { label: 'HTTP Requests', value: 'httpRequests' },
                {
                  label: 'HTTP RequestDuration',
                  value: 'httpRequestDuration',
                },
                { label: 'TCP Sent Bytes', value: 'tcpSentBytes' },
                { label: 'TCP Received Bytes', value: 'tcpReceivedBytes' },
              ]}
              onChange={(option: Array<ComboboxOption<string>>) => {
                onChange({
                  ...query,
                  metrics: Array.from(option.values()).map(
                    (value) => value.value,
                  ),
                });
              }}
            />
          </InlineField>
        </InlineFieldRow>

        <InlineFieldRow>
          <InlineField label="Idle Edges" labelWidth={25}>
            <InlineSwitch
              value={query.idleEdges || false}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onChange({ ...query, idleEdges: event.target.checked });
              }}
            />
          </InlineField>
        </InlineFieldRow>

        <InlineFieldRow>
          <FiltersField
            datasource={datasource}
            range={range}
            filterType="source"
            namespace={query.namespace}
            application={query.application}
            workload={query.workload}
            filters={query.sourceFilters}
            onFiltersChange={(filters) => {
              onChange({ ...query, sourceFilters: filters });
            }}
          />

          <FiltersField
            datasource={datasource}
            range={range}
            filterType="destination"
            namespace={query.namespace}
            application={query.application}
            workload={query.workload}
            filters={query.destinationFilters}
            onFiltersChange={(filters) => {
              onChange({ ...query, destinationFilters: filters });
            }}
          />
        </InlineFieldRow>
      </Collapse>
    </>
  );
}
