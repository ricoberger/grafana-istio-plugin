# Grafana Istio Plugin

The Grafana Istio Pluggin allows you to visualize your Istio metrics from
Prometheus via the Grafana
[node graph panel](https://grafana.com/docs/grafana/latest/visualizations/panels-visualizations/visualizations/node-graph/).

<div align="center">
  <table>
    <tr>
      <td><img src="https://raw.githubusercontent.com/ricoberger/grafana-istio-plugin/refs/heads/main/src/img/screenshots/example-bookinfo.png" /></td>
      <td><img src="https://raw.githubusercontent.com/ricoberger/grafana-istio-plugin/refs/heads/main/src/img/screenshots/example-otel-demo.png" /></td>
    </tr>
  </table>
</div>

## Installation

1. Before you can install the plugin, you have to add
   `ricoberger-istio-datasource` to the
   [`allow_loading_unsigned_plugins`](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#allow_loading_unsigned_plugins)
   configuration option or to the `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS`
   environment variable.
2. The plugin can then be installed by adding
   `ricoberger-istio-datasource@<VERSION>@https://github.com/ricoberger/grafana-istio-plugin/releases/download/v<VERSION>/ricoberger-istio-datasource-<VERSION>.zip`
   to the
   [`preinstall_sync`](https://grafana.com/docs/grafana/latest/setup-grafana/configure-grafana/#preinstall_sync)
   configuration option or the `GF_PLUGINS_PREINSTALL_SYNC` environment
   variable.

### Configuration File

```ini
[plugins]
allow_loading_unsigned_plugins = ricoberger-istio-datasource
preinstall_sync = ricoberger-istio-datasource@0.1.0@https://github.com/ricoberger/grafana-istio-plugin/releases/download/v0.1.0/ricoberger-istio-datasource-0.1.0.zip
```

### Environment Variables

```bash
export GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=ricoberger-istio-datasource
export GF_PLUGINS_PREINSTALL_SYNC=ricoberger-istio-datasource@0.1.0@https://github.com/ricoberger/grafana-istio-plugin/releases/download/v0.1.0/ricoberger-istio-datasource-0.1.0.zip
```

## Configuration

- **Prometheus Url:** The url of the Prometheus instance, e.g.
  `http://localhost:9090`.
- **Prometheus Authentication Method:** The authentication method which should
  be used for the Prometheus instance. The plugin supports basic authentication
  and bearer token authentication.
- **Istio Warning Threshold:** The threshold in percent which defines when a
  edge or node should be marked `yellow`. The default value is `0`.
- **Istio Error Threshold:** The threshold in percent which defines when a edge
  or node should be marked `red`. The default value is `5`.
- **Istio Workload Dashboard:** The link to the
  [Istio workload dashboard](https://grafana.com/grafana/dashboards/7630-istio-workload-dashboard/),
  e.g. `/d/istio-workload-dashboard/istio-workload-dashboard`.
- **Istio Service Dashboard:** The link to the
  [Istio service dashboard](https://grafana.com/grafana/dashboards/7636-istio-service-dashboard/),
  e.g. `/d/istio-service-dashboard/istio-service-dashboard`.

![Configuration](https://raw.githubusercontent.com/ricoberger/grafana-istio-plugin/refs/heads/main/src/img/screenshots/configuration.png)

## Contributing

If you want to contribute to the project, please read through the
[contribution guideline](https://github.com/ricoberger/grafana-istio-plugin/blob/main/CONTRIBUTING.md).
Please also follow our
[code of conduct](https://github.com/ricoberger/grafana-istio-plugin/blob/main/CODE_OF_CONDUCT.md)
in all your interactions with the project.
