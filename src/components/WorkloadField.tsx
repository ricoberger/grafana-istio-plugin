import React from 'react';
import { Combobox, ComboboxOption, InlineField } from '@grafana/ui';
import { useAsync } from 'react-use';

import { DataSource } from '../datasource';
import { TimeRange } from '@grafana/data';

interface Props {
  datasource: DataSource;
  range?: TimeRange;
  namespace?: string;
  workload?: string;
  onWorkloadChange: (value: string) => void;
}

export function WorkloadField({
  datasource,
  range,
  namespace,
  workload,
  onWorkloadChange,
}: Props) {
  const state = useAsync(async (): Promise<ComboboxOption[]> => {
    if (!namespace) {
      return [];
    }

    const result = await datasource.metricFindQuery(
      {
        refId: 'workloads',
        queryType: 'workloads',
        namespace: namespace,
      },
      { range: range },
    );

    const workloads = result.map((value) => {
      return { value: value.text };
    });
    return workloads;
  }, [datasource, namespace]);

  return (
    <InlineField label="Workload">
      <Combobox<string>
        data-testid="workload-combobox"
        value={workload}
        createCustomValue={true}
        options={state.value || []}
        onChange={(option: ComboboxOption<string>) => {
          onWorkloadChange(option.value);
        }}
      />
    </InlineField>
  );
}
