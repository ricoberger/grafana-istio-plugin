import React, { ChangeEvent, useState } from 'react';
import { QueryEditorProps } from '@grafana/data';
import {
  InlineFieldRow,
  InlineField,
  Combobox,
  ComboboxOption,
  RadioButtonGroup,
  Input,
} from '@grafana/ui';

import { DataSource } from '../datasource';
import { DEFAULT_QUERIES, Options, Query, QueryType } from '../types';
import { NamespaceField } from './NamespaceField';

interface Props extends QueryEditorProps<DataSource, any, Options, Query> { }

export function VariableQueryEditor({
  datasource,
  query,
  range,
  onChange,
  onRunQuery,
}: Props) {
  const [graphType, setGraphType] = useState<string>(
    query.application ? 'application' : query.workload ? 'workload' : '',
  );

  return (
    <>
      <InlineFieldRow>
        <InlineField label="Variable Type" labelWidth={25}>
          <Combobox<QueryType>
            value={query.queryType}
            options={[
              {
                label: 'Namespaces',
                value: 'namespaces',
              },
              {
                label: 'Applications',
                value: 'applications',
              },
              {
                label: 'Workloads',
                value: 'workloads',
              },
              {
                label: 'Filters',
                value: 'filters',
              },
            ]}
            onChange={(option: ComboboxOption<QueryType>) => {
              onChange({
                ...query,
                ...DEFAULT_QUERIES[option.value],
                queryType: option.value,
              });
            }}
          />
        </InlineField>
      </InlineFieldRow>

      {(query.queryType === 'applications' ||
        query.queryType === 'workloads' ||
        query.queryType === 'filters') && (
          <>
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
            </InlineFieldRow>

            {query.queryType === 'filters' && (
              <>
                <InlineFieldRow>
                  <InlineField label="Filter Type" labelWidth={25}>
                    <RadioButtonGroup<string>
                      options={[
                        { label: 'Source', value: 'source' },
                        { label: 'Destination', value: 'destination' },
                      ]}
                      value={query.filterType || ''}
                      onChange={(value: string) => {
                        if (value === 'destination' || value === 'source') {
                          onChange({
                            ...query,
                            filterType: value,
                          });
                          onRunQuery();
                        }
                      }}
                    />
                  </InlineField>
                </InlineFieldRow>
                <InlineFieldRow>
                  <InlineField label="Graph Type" labelWidth={25}>
                    <RadioButtonGroup<string>
                      options={[
                        { label: 'Application Graph', value: 'application' },
                        { label: 'Workload Graph', value: 'workload' },
                      ]}
                      value={graphType || ''}
                      onChange={(value: string) => {
                        setGraphType(value);
                      }}
                    />
                  </InlineField>
                  {(graphType === 'application' || graphType === 'workload') && (
                    <InlineField
                      label={
                        graphType === 'application'
                          ? 'Application'
                          : graphType === 'workload'
                            ? 'Workload'
                            : ''
                      }
                      labelWidth={25}
                      interactive
                    >
                      <Input
                        onChange={(event: ChangeEvent<HTMLInputElement>) => {
                          if (graphType === 'application') {
                            onChange({
                              ...query,
                              application: event.target.value,
                              workload: '',
                            });
                          } else if (graphType === 'workload') {
                            onChange({
                              ...query,
                              application: '',
                              workload: event.target.value,
                            });
                          }
                        }}
                        value={
                          graphType === 'application'
                            ? query.application
                            : graphType === 'workload'
                              ? query.workload
                              : ''
                        }
                      />
                    </InlineField>
                  )}
                </InlineFieldRow>
              </>
            )}
          </>
        )}
    </>
  );
}
