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
  const [input, setInput] = useState<string>(
    query.application ? 'application' : query.workload ? 'workload' : '',
  );

  return (
    <>
      <InlineFieldRow>
        <InlineField label="Type">
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

            {query.queryType === 'filters' && (
              <>
                <InlineField label="Type">
                  <RadioButtonGroup<string>
                    options={[
                      { label: 'Source', value: 'source' },
                      { label: 'Destination', value: 'destination' },
                    ]}
                    value={query.type || ''}
                    onChange={(value: string) => {
                      if (value === 'destination' || value === 'source') {
                        onChange({
                          ...query,
                          type: value,
                        });
                        onRunQuery();
                      }
                    }}
                  />
                </InlineField>
                <InlineField label="Input Type">
                  <RadioButtonGroup<string>
                    options={[
                      { label: 'Application', value: 'application' },
                      { label: 'Workload', value: 'workload' },
                    ]}
                    value={input || ''}
                    onChange={(value: string) => {
                      setInput(value);
                    }}
                  />
                </InlineField>
                {(input === 'application' || input === 'workload') && (
                  <InlineField label="Input" interactive>
                    <Input
                      onChange={(event: ChangeEvent<HTMLInputElement>) => {
                        if (input === 'application') {
                          onChange({
                            ...query,
                            application: event.target.value,
                            workload: '',
                          });
                        } else if (input === 'workload') {
                          onChange({
                            ...query,
                            application: '',
                            workload: event.target.value,
                          });
                        }
                      }}
                      value={
                        input === 'application'
                          ? query.application
                          : input === 'workload'
                            ? query.workload
                            : ''
                      }
                    />
                  </InlineField>
                )}
              </>
            )}
          </InlineFieldRow>
        )}
    </>
  );
}
