import React from 'react';
import { ComboboxOption, InlineField, MultiCombobox } from '@grafana/ui';
import { useAsync } from 'react-use';

import { DataSource } from '../datasource';
import { TimeRange } from '@grafana/data';

interface Props {
  datasource: DataSource;
  range?: TimeRange;
  type: 'source' | 'destination';
  namespace?: string;
  application?: string;
  workload?: string;
  filters?: string[];
  onFiltersChange: (value: string[]) => void;
}

export function FiltersField({
  datasource,
  range,
  type,
  namespace,
  application,
  workload,
  filters,
  onFiltersChange,
}: Props) {
  const state = useAsync(async (): Promise<ComboboxOption[]> => {
    if (!namespace) {
      return [];
    }

    const result = await datasource.metricFindQuery(
      {
        refId: 'filters',
        queryType: 'filters',
        type: type,
        namespace: namespace,
        application: application,
        workload: workload,
      },
      { range: range },
    );

    const workloads = result.map((value) => {
      return { value: value.text };
    });
    return workloads;
  }, [datasource, namespace]);

  return (
    <InlineField
      label={type === 'source' ? 'Source Filters' : 'Destination Filters'}
    >
      <MultiCombobox
        data-testid="filters-combobox"
        width="auto"
        minWidth={32}
        maxWidth={32}
        isClearable={true}
        value={filters}
        createCustomValue={true}
        options={state.value || []}
        onChange={(option: Array<ComboboxOption<string>>) => {
          onFiltersChange(
            Array.from(option.values()).map((value) => value.value),
          );
        }}
      />
    </InlineField>
  );
}
