import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor', async ({
  panelEditPage,
  readProvisionedDataSource,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(
    panelEditPage.getQueryEditorRow('A').getByTestId('namespace-combobox'),
  ).toBeVisible();
  await expect(
    panelEditPage.getQueryEditorRow('A').getByTestId('application-combobox'),
  ).toBeVisible();
});

test('should trigger new query when namespace and application is set', async ({
  panelEditPage,
  readProvisionedDataSource,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);

  const queryReqNamespaces = panelEditPage.waitForQueryDataRequest();
  await expect(await queryReqNamespaces).toBeTruthy();

  const queryReqApplications = panelEditPage.waitForQueryDataRequest();

  await panelEditPage
    .getQueryEditorRow('A')
    .getByTestId('namespace-combobox')
    .fill('echoserver');
  await panelEditPage
    .getQueryEditorRow('A')
    .getByTestId('namespace-combobox')
    .press('Enter');

  await expect(await queryReqApplications).toBeTruthy();

  const queryReq = panelEditPage.waitForQueryDataRequest();

  await panelEditPage
    .getQueryEditorRow('A')
    .getByTestId('application-combobox')
    .fill('echoserver');
  await panelEditPage
    .getQueryEditorRow('A')
    .getByTestId('application-combobox')
    .press('Enter');

  await expect(await queryReq).toBeTruthy();

  await panelEditPage.setVisualization('Node Graph');
  await expect(panelEditPage.refreshPanel()).toBeOK();
});
