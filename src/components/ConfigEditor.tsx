import React, { ChangeEvent } from 'react';
import {
  InlineField,
  Input,
  RadioButtonGroup,
  SecretInput,
  useStyles2,
} from '@grafana/ui';
import {
  DataSourcePluginOptionsEditorProps,
  GrafanaTheme2,
} from '@grafana/data';

import { Options, OptionsPrometheusAuthMethod, OptionsSecure } from '../types';
import { css } from '@emotion/css';

interface Props
  extends DataSourcePluginOptionsEditorProps<Options, OptionsSecure> { }

export function ConfigEditor({ options, onOptionsChange }: Props) {
  const styles = useStyles2((theme: GrafanaTheme2) => {
    return {
      container: css({
        paddingTop: theme.spacing(5),
      }),
    };
  });

  const { jsonData, secureJsonFields, secureJsonData } = options;

  return (
    <>
      <h3>Prometheus</h3>
      <InlineField label="Url" labelWidth={25} interactive>
        <Input
          data-testid="promtheus-url-input"
          onChange={(event: ChangeEvent<HTMLInputElement>) => {
            onOptionsChange({
              ...options,
              jsonData: {
                ...jsonData,
                prometheusUrl: event.target.value,
              },
            });
          }}
          value={jsonData.prometheusUrl}
          width={40}
        />
      </InlineField>

      <InlineField label="Authentication Method" labelWidth={25}>
        <RadioButtonGroup<OptionsPrometheusAuthMethod>
          options={[
            { label: 'None', value: 'none' },
            { label: 'Basic Auth', value: 'basic' },
            { label: 'Bearer Token', value: 'token' },
          ]}
          value={options.jsonData.prometheusAuthMethod || 'none'}
          onChange={(value: OptionsPrometheusAuthMethod) => {
            onOptionsChange({
              ...options,
              jsonData: {
                ...options.jsonData,
                prometheusAuthMethod: value,
              },
            });
          }}
        />
      </InlineField>

      {jsonData.prometheusAuthMethod === 'basic' && (
        <>
          <InlineField label="Username" labelWidth={25} interactive>
            <Input
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onOptionsChange({
                  ...options,
                  jsonData: {
                    ...jsonData,
                    prometheusUsername: event.target.value,
                  },
                });
              }}
              value={jsonData.prometheusUsername}
              width={40}
            />
          </InlineField>
          <InlineField label="Password" labelWidth={25} interactive>
            <SecretInput
              required
              isConfigured={secureJsonFields.apiKey}
              value={secureJsonData?.prometheusPassword}
              width={40}
              onReset={() => {
                onOptionsChange({
                  ...options,
                  secureJsonFields: {
                    ...options.secureJsonFields,
                    prometheusPassword: false,
                  },
                  secureJsonData: {
                    ...options.secureJsonData,
                    prometheusPassword: '',
                  },
                });
              }}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onOptionsChange({
                  ...options,
                  secureJsonData: {
                    prometheusPassword: event.target.value,
                  },
                });
              }}
            />
          </InlineField>
        </>
      )}

      {jsonData.prometheusAuthMethod === 'token' && (
        <>
          <InlineField label="Bearer Token" labelWidth={25} interactive>
            <SecretInput
              required
              isConfigured={secureJsonFields.apiKey}
              value={secureJsonData?.prometheusPassword}
              width={40}
              onReset={() => {
                onOptionsChange({
                  ...options,
                  secureJsonFields: {
                    ...options.secureJsonFields,
                    prometheusPassword: false,
                  },
                  secureJsonData: {
                    ...options.secureJsonData,
                    prometheusPassword: '',
                  },
                });
              }}
              onChange={(event: ChangeEvent<HTMLInputElement>) => {
                onOptionsChange({
                  ...options,
                  secureJsonData: {
                    prometheusPassword: event.target.value,
                  },
                });
              }}
            />
          </InlineField>
        </>
      )}

      <div className={styles.container}>
        <h3>Istio</h3>
        <InlineField label="Warning Threshold" labelWidth={25} interactive>
          <Input
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              onOptionsChange({
                ...options,
                jsonData: {
                  ...jsonData,
                  istioWarningThreshold: parseFloat(event.target.value),
                },
              });
            }}
            value={jsonData.istioWarningThreshold}
            width={40}
          />
        </InlineField>
        <InlineField label="Error Threshold" labelWidth={25} interactive>
          <Input
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              onOptionsChange({
                ...options,
                jsonData: {
                  ...jsonData,
                  istioErrorThreshold: parseFloat(event.target.value),
                },
              });
            }}
            value={jsonData.istioErrorThreshold}
            width={40}
          />
        </InlineField>

        <InlineField label="Workload Dashboard" labelWidth={25} interactive>
          <Input
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              onOptionsChange({
                ...options,
                jsonData: {
                  ...jsonData,
                  istioWorkloadDashboard: event.target.value,
                },
              });
            }}
            value={jsonData.istioWorkloadDashboard}
            width={40}
          />
        </InlineField>
        <InlineField label="Service Dashboard" labelWidth={25} interactive>
          <Input
            onChange={(event: ChangeEvent<HTMLInputElement>) => {
              onOptionsChange({
                ...options,
                jsonData: {
                  ...jsonData,
                  istioServiceDashboard: event.target.value,
                },
              });
            }}
            value={jsonData.istioServiceDashboard}
            width={40}
          />
        </InlineField>
      </div>
    </>
  );
}
