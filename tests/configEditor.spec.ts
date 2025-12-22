import { test, expect } from '@grafana/plugin-e2e';

import { Options, OptionsSecure } from '../src/types';

test('smoke: should render config editor', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({
    fileName: 'datasources.yml',
    name: 'Istio',
  });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByTestId('promtheus-url-input')).toBeVisible();
});

test('"Save & test" should be successful when configuration is valid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<Options, OptionsSecure>({
    fileName: 'datasources.yml',
    name: 'Istio',
  });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByTestId('promtheus-url-input').fill('http://wiremock:8080');
  await expect(configPage.saveAndTest()).toBeOK();
});

test('"Save & test" should fail when configuration is invalid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource<Options, OptionsSecure>({
    fileName: 'datasources.yml',
    name: 'Istio',
  });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByTestId('promtheus-url-input').fill('');
  await expect(configPage.saveAndTest()).not.toBeOK();
  await expect(configPage).toHaveAlert('error', {
    hasText:
      'Prometheus health check failed: Get "/api/v1/status/buildinfo": unsupported protocol scheme ""',
  });
});
