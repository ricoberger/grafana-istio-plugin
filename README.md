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
    <tr>
      <td><img src="https://raw.githubusercontent.com/ricoberger/grafana-istio-plugin/refs/heads/main/src/img/screenshots/example-bookinfo-options-and-details.png" /></td>
      <td><img src="https://raw.githubusercontent.com/ricoberger/grafana-istio-plugin/refs/heads/main/src/img/screenshots/example-dashboard.png" /></td>
    </tr>
  </table>
</div>

> [!NOTE]
>
> The plugin requires the
> [Istio Standard Metrics](https://istio.io/latest/docs/reference/config/metrics/)
> to be available in your Prometheus instance.
>
> The plugin was only tested with Istio Ambient Mode. However, it should also
> work for Istio Sidecar deployments, but might give some wrong results, because
> the plugin currently doesn't handle the reporter of the metrics.

## Usage

### Query Options

- Graph Type: Select between **Application Graph**, **Workload Graph** and
  **Namespace Graph**, to visualize an application, workload or whole namespace.
- Namespace: Select the **Namespace** of the application or workload or if the
  **Namespace Graph** type is selected, the namespace which should be
  visualized.
- Application / Workload: Select the **Application** or **Workload** which
  should be visualized.
- Metrics: Select the metrics which should be included in the visualization. The
  available metrics are: **gRPC Requests**, **gRPC Request Duration**, **gRPC
  Sent Messages**, **gRPC Received Messages**, **HTTP Requests**, **HTTP Request
  Duration**, **TCP Sent Bytes** and **TCP Received Bytes**. If a metric is not
  selected, it might be that an edge between two nodes in the graph is not
  shown, because there is no traffic for the selected metrics.
- Idle Edges: If selected the graph will also shown **Idle Edges**, which means
  edges which do not have any traffic in the selected time range.
- Filters: Add multiple **Source Filters** and **Destination Filters** for
  workloads, which should not be shown in the graph.

### Variable Query Options

- Variable Type: Select the type of the variable. The available types are
  **Namespaces**, **Applications**, **Workloads** and **Filters**.
- Namespace: Select the **Namespace** for an application, workload or filter
  variable.
- Filter Type: Select the type of the filter, when the variable type is set to
  **Filters**. The available filter types are **Source** and **Destination**.
- Graph Type: Select the graph type for which the filter variable is used. The
  available options are **Application Graph** and **Workload Graph**. Depending
  on the selection also a **Application** or **Workload** is required to
  determine the source / destination filters for the application or workload.

### Legend

- **Color:** Depending on the traffic type an edge / node can have multiple
  colors:
  - **Green:** The edge / node has gRPC or HTTP traffic without an error rate.
  - **Yellow:** The edge / node has gRPC or HTTP traffic with an error rate
    above the configured warning threshold.
  - **Red:** The edge / node has gRPC or HTTP traffic with an error rate above
    the configured error threshold.
  - **Blue:** The edge / node has TCP traffic.
  - **Gray:** The edge / node has no traffic in the selected time range.
- **Main / Secondary Stats:** The main statistics which are shown on an edge /
  node:
  - For edges / nodes with more HTTP then gRPC traffic, we show the number of
    HTTP requests per second and the error rate in percent for HTTP traffic. The
    secondary stats contains the P95 request duration in milliseconds if
    available and the total number of bytes sent and received per second.
  - For edges / nodes with more gRPC then HTTP traffic, we show the number of
    gRPC requests per second and the error rate in percent for gRPC traffic. The
    secondary stats contains the P95 request duration in milliseconds if
    available and the total number of bytes sent and received per second.
  - For edges / nodes with only TCP traffic, we show the total number of sent
    and received bytes per second.
  - For _workload_ nodes, we always try to show the server side statistics
    first. If the workload does not have any server side traffic, we show the
    client side statistics.

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
  e.g.
  `/d/istio-workload-dashboard/istio-workload-dashboard?orgId=1&var-datasource=prometheus&var-qrep=waypoint&var-qrep=source&var-qrep=destination`.
  The plugin adds the following query parameters to the provided dashboard url:
  `&var-service=<SERVICE>&from=<FROM>&to=<TO>`. ``
- **Istio Service Dashboard:** The link to the
  [Istio service dashboard](https://grafana.com/grafana/dashboards/7636-istio-service-dashboard/),
  e.g.
  `/d/istio-service-dashboard/istio-service-dashboard?orgId=1&var-datasource=prometheus&var-qrep=waypoint&var-qrep=source&var-qrep=destination`.
  The plugin adds the following query parameters to the provided dashboard url:
  `&var-namespace=<WORKLOAD-NAMESPACE>&var-workload=<WORKLOAD-NAME>&from=<FROM>&to=<TO>`.
  ``

![Configuration](https://raw.githubusercontent.com/ricoberger/grafana-istio-plugin/refs/heads/main/src/img/screenshots/configuration.png)

## Contributing

If you want to contribute to the project, please read through the
[contribution guideline](https://github.com/ricoberger/grafana-istio-plugin/blob/main/CONTRIBUTING.md).
Please also follow our
[code of conduct](https://github.com/ricoberger/grafana-istio-plugin/blob/main/CODE_OF_CONDUCT.md)
in all your interactions with the project.
