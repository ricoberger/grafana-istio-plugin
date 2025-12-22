import React, { ChangeEvent } from 'react';
import {
  ComboboxOption,
  InlineField,
  InlineFieldRow,
  InlineSwitch,
  MultiCombobox,
} from '@grafana/ui';
import { QueryEditorProps } from '@grafana/data';

import { DataSource } from '../datasource';
import { Options, Query } from '../types';
import { NamespaceField } from './NamespaceField';
import { ApplicationField } from './ApplicationField';

type Props = QueryEditorProps<DataSource, Query, Options>;

export function QueryEditor({
  datasource,
  query,
  range,
  onChange,
  onRunQuery,
}: Props) {
  return (
    <InlineFieldRow>
      <NamespaceField
        datasource={datasource}
        range={range}
        namespace={query.namespace}
        onNamespaceChange={(namespace) => {
          onChange({ ...query, namespace: namespace, application: '' });
          onRunQuery();
        }}
      />
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
      <InlineField label="Metrics">
        <MultiCombobox
          data-testid="metrics-combobox"
          width="auto"
          minWidth={32}
          maxWidth={32}
          isClearable={true}
          value={query.metrics}
          options={[
            { label: 'gRPC Requests', value: 'grpcRequests' },
            { label: 'gRPC Request Duration', value: 'grpcRequestDuration' },
            { label: 'gRPC Sent Messages', value: 'grpcSentMessages' },
            {
              label: 'gRPC Received Messages',
              value: 'grpcReceivedMessages',
            },
            { label: 'HTTP Requests', value: 'httpRequests' },
            { label: 'HTTP RequestDuration', value: 'httpRequestDuration' },
            { label: 'TCP Sent Bytes', value: 'tcpSentBytes' },
            { label: 'TCP Received Bytes', value: 'tcpReceivedBytes' },
          ]}
          onChange={(option: Array<ComboboxOption<string>>) => {
            onChange({
              ...query,
              metrics: Array.from(option.values()).map((value) => value.value),
            });
          }}
        />
      </InlineField>
      <InlineField label="Idle Edges">
        <InlineSwitch
          value={query.idleEdges || false}
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            onChange({ ...query, idleEdges: event.target.checked });
            onRunQuery();
          }}
        />
      </InlineField>
    </InlineFieldRow>
  );
}
