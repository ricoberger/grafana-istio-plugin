import React from 'react';
import { Combobox, ComboboxOption, InlineField } from '@grafana/ui';
import { useAsync } from 'react-use';

import { DataSource } from '../datasource';
import { TimeRange } from '@grafana/data';

interface Props {
  datasource: DataSource;
  range?: TimeRange;
  namespace?: string;
  application?: string;
  onApplicationChange: (value: string) => void;
}

export function ApplicationField({
  datasource,
  range,
  namespace,
  application,
  onApplicationChange,
}: Props) {
  const state = useAsync(async (): Promise<ComboboxOption[]> => {
    if (!namespace) {
      return [];
    }

    const result = await datasource.metricFindQuery(
      {
        refId: 'applications',
        queryType: 'applications',
        namespace: namespace,
      },
      { range: range },
    );

    const applications = result.map((value) => {
      return { value: value.text };
    });
    return applications;
  }, [datasource, namespace]);

  return (
    <InlineField label="Application" labelWidth={25}>
      <Combobox<string>
        data-testid="application-combobox"
        value={application}
        createCustomValue={true}
        options={state.value || []}
        onChange={(option: ComboboxOption<string>) => {
          onApplicationChange(option.value);
        }}
      />
    </InlineField>
  );
}
