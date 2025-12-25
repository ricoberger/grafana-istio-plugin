import React from 'react';
import { Combobox, ComboboxOption, InlineField } from '@grafana/ui';
import { useAsync } from 'react-use';

import { DataSource } from '../datasource';
import { TimeRange } from '@grafana/data';

interface Props {
  datasource: DataSource;
  range?: TimeRange;
  namespace?: string;
  onNamespaceChange: (value: string) => void;
}

export function NamespaceField({
  datasource,
  range,
  namespace,
  onNamespaceChange,
}: Props) {
  const state = useAsync(async (): Promise<ComboboxOption[]> => {
    const result = await datasource.metricFindQuery(
      {
        refId: 'namespaces',
        queryType: 'namespaces',
      },
      { range: range },
    );

    const namespaces = result.map((value) => {
      return { value: value.text };
    });
    return namespaces;
  }, [datasource]);

  return (
    <InlineField label="Namespace" labelWidth={25}>
      <Combobox<string>
        data-testid="namespace-combobox"
        value={namespace}
        createCustomValue={true}
        options={state.value || []}
        onChange={(option: ComboboxOption<string>) => {
          onNamespaceChange(option.value);
        }}
      />
    </InlineField>
  );
}
