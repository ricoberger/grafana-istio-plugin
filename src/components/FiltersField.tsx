import React from 'react';
import { ComboboxOption, InlineField, MultiCombobox } from '@grafana/ui';
import { useAsync } from 'react-use';

import { DataSource } from '../datasource';
import { TimeRange } from '@grafana/data';
import { QueryModelFiltersFilterType } from '../types';

interface Props {
  datasource: DataSource;
  range?: TimeRange;
  filterType: QueryModelFiltersFilterType;
  namespace?: string;
  application?: string;
  workload?: string;
  filters?: string[];
  onFiltersChange: (value: string[]) => void;
}

export function FiltersField({
  datasource,
  range,
  filterType,
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
        filterType: filterType,
        namespace: namespace,
        application: application,
        workload: workload,
      },
      { range: range },
    );

    const workloads = result.map((value) => {
      return { label: value.text, value: value.text };
    });
    return workloads;
  }, [datasource, namespace, application, workload, filterType, range]);

  return (
    <InlineField
      label={filterType === 'source' ? 'Source Filters' : 'Destination Filters'}
      labelWidth={25}
    >
      <MultiCombobox
        data-testid={`${filterType}-filters-combobox`}
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
