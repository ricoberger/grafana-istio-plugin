package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/ricoberger/grafana-istio-plugin/pkg/models"
	"github.com/ricoberger/grafana-istio-plugin/pkg/prometheus"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/experimental/concurrent"
	"go.opentelemetry.io/otel/codes"
)

// handleNamespacesQueries handles the queries to get a list of namespaces. It
// uses the concurrent package to handle multiple queries in parallel. The
// namespaces are retrieved from the "destination_workload_namespace",
// "source_workload_namespace", and "destination_service_namespace" labels.
func (d *Datasource) handleNamespacesQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleNamespacesQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleNamespaces, 10)
}

func (d *Datasource) handleNamespaces(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleNamespaces")
	defer span.End()

	queries := []prometheus.LabelValuesQuery{{
		Label: "destination_workload_namespace",
		Matches: []string{
			"istio_requests_total",
			"istio_tcp_sent_bytes_total",
			"istio_tcp_received_bytes_total",
		},
	}, {
		Label: "source_workload_namespace",
		Matches: []string{
			"istio_requests_total",
			"istio_tcp_sent_bytes_total",
			"istio_tcp_received_bytes_total",
		},
	}}

	return d.handelLabelValues(ctx, queries, query.DataQuery.TimeRange)
}

// handleApplicationQueries handles the queries to get a list of applications.
// It uses the concurrent package to handle multiple queries in parallel. The
// applications are retrieved from the "destination_app" label.
func (d *Datasource) handleApplicationsQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleApplicationsQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleApplications, 10)
}

func (d *Datasource) handleApplications(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleApplications")
	defer span.End()

	var qm models.QueryModelApplications
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}

	queries := []prometheus.LabelValuesQuery{{
		Label: "destination_app",
		Matches: []string{
			fmt.Sprintf("istio_requests_total{destination_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_sent_bytes_total{destination_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_received_bytes_total{destination_workload_namespace=\"%s\"}", qm.Namespace),
		},
	}, {
		Label: "source_app",
		Matches: []string{
			fmt.Sprintf("istio_requests_total{source_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_sent_bytes_total{source_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_received_bytes_total{source_workload_namespace=\"%s\"}", qm.Namespace),
		},
	}}

	return d.handelLabelValues(ctx, queries, query.DataQuery.TimeRange)
}

// handleLabelValues retrieves the values for the given labels and filter from
// the "istio_requests_total", "istio_tcp_sent_bytes_total", and
// "istio_tcp_received_bytes_total" metrics. It performs the retrieval in
// parallel for each label and combines the results into a single response.
func (d *Datasource) handelLabelValues(ctx context.Context, queries []prometheus.LabelValuesQuery, timeRange backend.TimeRange) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleLabelValues")
	defer span.End()

	var errors []error
	errorsMutex := &sync.Mutex{}

	var values [][]string
	valuesMutex := &sync.Mutex{}

	var queriesWG sync.WaitGroup
	queriesWG.Add(len(queries))

	for _, query := range queries {
		go func(query prometheus.LabelValuesQuery) {
			defer queriesWG.Done()

			d.logger.Debug("Get label values", "label", query.Label, "matches", query.Matches, "timeRangeFrom", timeRange.From, "timeRangeTo", timeRange.To)
			labelValues, err := d.prometheusClient.GetLabelValues(ctx, query, timeRange)
			if err != nil {
				d.logger.Error("Failed to get values", "error", err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				errorsMutex.Lock()
				errors = append(errors, err)
				errorsMutex.Unlock()
				return
			}
			d.logger.Debug("Retrieved label values", "label", query.Label, "matches", query.Matches, "values", labelValues)

			valuesMutex.Lock()
			values = append(values, labelValues)
			valuesMutex.Unlock()
		}(query)
	}

	queriesWG.Wait()

	if len(errors) > 0 {
		span.RecordError(errors[0])
		span.SetStatus(codes.Error, errors[0].Error())
		return backend.ErrorResponseWithErrorSource(errors[0])
	}

	var allValues []string
	for _, v := range values {
		allValues = append(allValues, v...)
	}
	slices.Sort(allValues)
	allValues = slices.Compact(allValues)

	frame := data.NewFrame(
		"Values",
		data.NewField("values", nil, allValues),
	)

	frame.SetMeta(&data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
		Type:                   data.FrameTypeTable,
	})

	var response backend.DataResponse
	response.Frames = append(response.Frames, frame)

	return response
}

// handleGraphQueries handles the queries to get graph for an application. It
// uses the concurrent package to handle multiple queries in parallel. The graph
// is generated for an specific application in a namespace and contains all the
// requested metrics.
func (d *Datasource) handleGraphQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleGraphQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleGraph, 10)
}

func (d *Datasource) handleGraph(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleGraph")
	defer span.End()

	var qm models.QueryModelGraph
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}
	interval := int64(query.DataQuery.TimeRange.Duration().Seconds())

	var errors []error
	errorsMutex := &sync.Mutex{}

	var metrics []prometheus.Metric
	metricsMutex := &sync.Mutex{}

	var metricsWG sync.WaitGroup
	metricsWG.Add(len(qm.Metrics))

	for _, metric := range qm.Metrics {
		go func(metric string) {
			defer metricsWG.Done()

			d.logger.Debug("Get metric", "metric", metric, "namespace", qm.Namespace, "application", qm.Application, "timeRangeFrom", query.DataQuery.TimeRange.From, "timeRangeTo", query.DataQuery.TimeRange.To, "interval", interval)

			applicationTargetMetrics, err := d.prometheusClient.GetMetrics(ctx, metric, d.metricToQueryApplicationTarget(qm.Namespace, qm.Application, metric, interval), query.DataQuery.TimeRange)
			if err != nil {
				d.logger.Error("Failed to get metric", "error", err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				errorsMutex.Lock()
				errors = append(errors, err)
				errorsMutex.Unlock()
				return
			}
			d.logger.Debug("Retrieved metrics where application is target", "metric", metric, "namespace", qm.Namespace, "application", qm.Application, "metrics", applicationTargetMetrics)

			applicationSourceMetrics, err := d.prometheusClient.GetMetrics(ctx, metric, d.metricToQueryApplicationSource(qm.Namespace, qm.Application, metric, interval), query.DataQuery.TimeRange)
			if err != nil {
				d.logger.Error("Failed to get metric", "error", err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				errorsMutex.Lock()
				errors = append(errors, err)
				errorsMutex.Unlock()
				return
			}
			d.logger.Debug("Retrieved metrics where application is source", "metric", metric, "namespace", qm.Namespace, "application", qm.Application, "metrics", applicationSourceMetrics)

			metricsMutex.Lock()
			metrics = append(metrics, applicationTargetMetrics...)
			metrics = append(metrics, applicationSourceMetrics...)
			metricsMutex.Unlock()
		}(metric)
	}

	metricsWG.Wait()

	if len(errors) > 0 {
		span.RecordError(errors[0])
		span.SetStatus(codes.Error, errors[0].Error())
		return backend.ErrorResponseWithErrorSource(errors[0])
	}

	edges := d.metricsToEdges(metrics)
	nodes := d.edgesToNodes(edges)

	edgeFields := models.Fields{}
	edgeIds := edgeFields.Add("id", nil, []string{})
	edgeSources := edgeFields.Add("source", nil, []string{})
	edgeTargets := edgeFields.Add("target", nil, []string{})
	edgeMainStat := edgeFields.Add("mainstat", nil, []string{}, &data.FieldConfig{DisplayName: "Traffic Rates"})
	edgeSecondaryStat := edgeFields.Add("secondarystat", nil, []string{}, &data.FieldConfig{DisplayName: "Response Time / Throughput"})
	edgeColors := edgeFields.Add("color", nil, []string{}, &data.FieldConfig{DisplayName: "Health"})
	edgeDetailsGRPCRate := edgeFields.Add("detail__grpcrate", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Rate"})
	edgeDetailsGRPCErr := edgeFields.Add("detail__grpcperr", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Error"})
	edgeDetailsGRPCDuration := edgeFields.Add("detail__grpcduration", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Duration"})
	edgeDetailsGRPCSentMessages := edgeFields.Add("detail__grpcsentmessages", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Sent Messages"})
	edgeDetailsGRPCReceivedMessages := edgeFields.Add("detail__grpcreceivedmessages", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Received Messages"})
	edgeDetailsHTTPRate := edgeFields.Add("detail__httprate", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Rate"})
	edgeDetailsHTTPErr := edgeFields.Add("detail__httperr", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Error"})
	edgeDetailsHTTPDuration := edgeFields.Add("detail__httpduration", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Duration"})
	edgeDetailsTCPSentBytes := edgeFields.Add("detail__tcpsentbytes", nil, []string{}, &data.FieldConfig{DisplayName: "TCP Sent"})
	edgeDetailsTCPReceivedBytes := edgeFields.Add("detail__tcpreceivedbytes", nil, []string{}, &data.FieldConfig{DisplayName: "TCP Received"})

	for _, edge := range edges {
		edgeField := getEdgeField(edge, float64(interval))

		edgeIds.Append(edgeField.ID)
		edgeSources.Append(edgeField.Source)
		edgeTargets.Append(edgeField.Target)
		edgeMainStat.Append(strings.Join(edgeField.MainStat, " | "))
		edgeSecondaryStat.Append(strings.Join(edgeField.SecondaryStat, " | "))
		edgeColors.Append(edgeField.Color)
		edgeDetailsGRPCRate.Append(edgeField.DetailsGRPCRate)
		edgeDetailsGRPCErr.Append(edgeField.DetailsGRPCErr)
		edgeDetailsGRPCDuration.Append(edgeField.DetailsGRPCDuration)
		edgeDetailsGRPCSentMessages.Append(edgeField.DetailsGRPCSentMessages)
		edgeDetailsGRPCReceivedMessages.Append(edgeField.DetailsGRPCReceivedMessages)
		edgeDetailsHTTPRate.Append(edgeField.DetailsHTTPRate)
		edgeDetailsHTTPErr.Append(edgeField.DetailsHTTPErr)
		edgeDetailsHTTPDuration.Append(edgeField.DetailsHTTPDuration)
		edgeDetailsTCPSentBytes.Append(edgeField.DetailsTCPSentBytes)
		edgeDetailsTCPReceivedBytes.Append(edgeField.DetailsTCPReceivedBytes)
	}

	nodeFields := models.Fields{}
	nodeIds := nodeFields.Add("id", nil, []string{})
	nodeTitles := nodeFields.Add("title", nil, []string{}, &data.FieldConfig{DisplayName: "Type"})
	nodeSubTitles := nodeFields.Add("subtitle", nil, []string{}, &data.FieldConfig{DisplayName: "Name (Namespace)"})
	nodeMainStat := nodeFields.Add("mainstat", nil, []string{}, &data.FieldConfig{DisplayName: "Traffic Rates"})
	nodeSecondaryStat := nodeFields.Add("secondarystat", nil, []string{}, &data.FieldConfig{DisplayName: "Response Time / Throughput"})
	nodeColors := nodeFields.Add("color", nil, []string{}, &data.FieldConfig{DisplayName: "Health"})
	nodeDetailsGRPCRate := nodeFields.Add("detail__grpcrate", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Rate"})
	nodeDetailsGRPCErr := nodeFields.Add("detail__grpcperr", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Error"})
	nodeDetailsGRPCDuration := nodeFields.Add("detail__grpcduration", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Duration"})
	nodeDetailsGRPCSentMessages := nodeFields.Add("detail__grpcsentmessages", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Sent Messages"})
	nodeDetailsGRPCReceivedMessages := nodeFields.Add("detail__grpcreceivedmessages", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Received Messages"})
	nodeDetailsHTTPRate := nodeFields.Add("detail__httprate", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Rate"})
	nodeDetailsHTTPErr := nodeFields.Add("detail__httperr", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Error"})
	nodeDetailsHTTPDuration := nodeFields.Add("detail__httpduration", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Duration"})
	nodeDetailsTCPSentBytes := nodeFields.Add("detail__tcpsentbytes", nil, []string{}, &data.FieldConfig{DisplayName: "TCP Sent"})
	nodeDetailsTCPReceivedBytes := nodeFields.Add("detail__tcpreceivedbytes", nil, []string{}, &data.FieldConfig{DisplayName: "TCP Received"})
	nodeLink := nodeFields.Add("link", nil, []string{}, &data.FieldConfig{
		Links: []data.DataLink{
			{
				Title: "Istio Dashboard",
				URL:   "${__data.fields[\"link\"]}",
			},
		},
	})

	for _, node := range nodes {
		nodeField := getEdgeField(node, float64(interval))

		nodeIds.Append(nodeField.ID)
		nodeTitles.Append(node.TargetType)
		nodeSubTitles.Append(fmt.Sprintf("%s (%s)", node.TargetName, node.TargetNamespace))
		nodeMainStat.Append(strings.Join(nodeField.MainStat, " | "))
		nodeSecondaryStat.Append(strings.Join(nodeField.SecondaryStat, " | "))
		nodeColors.Append(nodeField.Color)
		nodeDetailsGRPCRate.Append(nodeField.DetailsGRPCRate)
		nodeDetailsGRPCErr.Append(nodeField.DetailsGRPCErr)
		nodeDetailsGRPCDuration.Append(nodeField.DetailsGRPCDuration)
		nodeDetailsGRPCSentMessages.Append(nodeField.DetailsGRPCSentMessages)
		nodeDetailsGRPCReceivedMessages.Append(nodeField.DetailsGRPCReceivedMessages)
		nodeDetailsHTTPRate.Append(nodeField.DetailsHTTPRate)
		nodeDetailsHTTPErr.Append(nodeField.DetailsHTTPErr)
		nodeDetailsHTTPDuration.Append(nodeField.DetailsHTTPDuration)
		nodeDetailsTCPSentBytes.Append(nodeField.DetailsTCPSentBytes)
		nodeDetailsTCPReceivedBytes.Append(nodeField.DetailsTCPReceivedBytes)

		switch node.TargetType {
		case "Service":
			nodeLink.Append(fmt.Sprintf("%s?var-service=%s", d.istioServiceDashboard, node.TargetService))
		case "Workload":
			nodeLink.Append(fmt.Sprintf("%s?var-namespace=%s&var-workload=%s", d.istioWorkloadDashboard, node.TargetNamespace, node.TargetName))
		default:
			nodeLink.Append("")
		}
	}

	edgeFrame := data.NewFrame("edges", edgeFields...).SetMeta(&data.FrameMeta{PreferredVisualization: data.VisTypeNodeGraph})
	nodeFrame := data.NewFrame("nodes", nodeFields...).SetMeta(&data.FrameMeta{PreferredVisualization: data.VisTypeNodeGraph})

	var response backend.DataResponse
	response.Frames = append(response.Frames, edgeFrame)
	response.Frames = append(response.Frames, nodeFrame)

	return response
}

func (d *Datasource) metricToQueryApplicationTarget(namespace, application, metric string, interval int64) string {
	switch metric {
	case models.MetricGRPCRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{destination_workload_namespace="%s", destination_app="%s", request_protocol="grpc"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status)`, namespace, application, interval)
	case models.MetricGRPCRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{destination_workload_namespace="%s", destination_app="%s", request_protocol="grpc"}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status))`, namespace, application, interval)
	case models.MetricGRPCSentMessages:
		return fmt.Sprintf(`sum(increase(istio_request_messages_total{source_workload_namespace="%s", source_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status)`, namespace, application, interval)
	case models.MetricGRPCReceivedMessages:
		return fmt.Sprintf(`sum(increase(istio_response_messages_total{source_workload_namespace="%s", source_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status)`, namespace, application, interval)
	case models.MetricHTTPRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{destination_workload_namespace="%s", destination_app="%s", request_protocol="http"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, response_code)`, namespace, application, interval)
	case models.MetricHTTPRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{destination_workload_namespace="%s", destination_app="%s", request_protocol="http"}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status))`, namespace, application, interval)
	case models.MetricTCPSentBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_sent_bytes_total{destination_workload_namespace="%s", destination_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)`, namespace, application, interval)
	case models.MetricTCPReceivedBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_received_bytes_total{destination_workload_namespace="%s", destination_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)`, namespace, application, interval)
	default:
		return ""
	}
}

func (d *Datasource) metricToQueryApplicationSource(namespace, application, metric string, interval int64) string {
	switch metric {
	case models.MetricGRPCRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{source_workload_namespace="%s", source_app="%s", request_protocol="grpc"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status)`, namespace, application, interval)
	case models.MetricGRPCRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{source_workload_namespace="%s", source_app="%s", request_protocol="grpc"}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status))`, namespace, application, interval)
	case models.MetricGRPCSentMessages:
		return fmt.Sprintf(`sum(increase(istio_request_messages_total{source_workload_namespace="%s", source_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status)`, namespace, application, interval)
	case models.MetricGRPCReceivedMessages:
		return fmt.Sprintf(`sum(increase(istio_response_messages_total{source_workload_namespace="%s", source_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status)`, namespace, application, interval)
	case models.MetricHTTPRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{source_workload_namespace="%s", source_app="%s", request_protocol="http"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, response_code)`, namespace, application, interval)
	case models.MetricHTTPRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{source_workload_namespace="%s", source_app="%s", request_protocol="http"}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status))`, namespace, application, interval)
	case models.MetricTCPSentBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_sent_bytes_total{source_workload_namespace="%s", source_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)`, namespace, application, interval)
	case models.MetricTCPReceivedBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_received_bytes_total{source_workload_namespace="%s", source_app="%s"}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)`, namespace, application, interval)
	default:
		return ""
	}
}

func (d *Datasource) metricsToEdges(metrics []prometheus.Metric) []models.Edge {
	edges := make(map[string]models.Edge)

	for _, m := range metrics {
		workloadToServiceId := fmt.Sprintf("Workload: %s (%s) - Service: %s (%s)", m.Labels["source_workload"], m.Labels["source_workload_namespace"], m.Labels["destination_service_name"], m.Labels["destination_service_namespace"])
		workloadToServiceSource := fmt.Sprintf("Workload: %s (%s)", m.Labels["source_workload"], m.Labels["source_workload_namespace"])
		workloadToServiceSourceType := "Workload"
		workloadToServiceSourceName := m.Labels["source_workload"]
		workloadToServiceSourceNamespace := m.Labels["source_workload_namespace"]
		workloadToServiceTarget := fmt.Sprintf("Service: %s (%s)", m.Labels["destination_service_name"], m.Labels["destination_service_namespace"])
		workloadToServiceTargetType := "Service"
		workloadToServiceTargetName := m.Labels["destination_service_name"]
		workloadToServiceTargetNamespace := m.Labels["destination_service_namespace"]
		workloadToServiceTargetService := m.Labels["destination_service"]

		serviceToWorkloadId := fmt.Sprintf("Service: %s (%s) - Workload: %s (%s)", m.Labels["destination_service_name"], m.Labels["destination_service_namespace"], m.Labels["destination_workload"], m.Labels["destination_workload_namespace"])
		serviceToWorkloadSource := fmt.Sprintf("Service: %s (%s)", m.Labels["destination_service_name"], m.Labels["destination_service_namespace"])
		serviceToServiceSourceType := "Service"
		serviceToServiceSourceName := m.Labels["destination_service_name"]
		serviceToServiceSourceNamespace := m.Labels["destination_service_namespace"]
		serviceToWorkloadTarget := fmt.Sprintf("Workload: %s (%s)", m.Labels["destination_workload"], m.Labels["destination_workload_namespace"])
		serviceToServiceTargetType := "Workload"
		serviceToServiceTargetName := m.Labels["destination_workload"]
		serviceToServiceTargetNamespace := m.Labels["destination_workload_namespace"]
		serviceToServiceTargetService := m.Labels["destination_service"]

		if _, ok := edges[workloadToServiceId]; !ok {
			edges[workloadToServiceId] = models.Edge{
				ID:                   workloadToServiceId,
				Source:               workloadToServiceSource,
				SourceType:           workloadToServiceSourceType,
				SourceName:           workloadToServiceSourceName,
				SourceNamespace:      workloadToServiceSourceNamespace,
				Target:               workloadToServiceTarget,
				TargetType:           workloadToServiceTargetType,
				TargetName:           workloadToServiceTargetName,
				TargetNamespace:      workloadToServiceTargetNamespace,
				TargetService:        workloadToServiceTargetService,
				GRPCResponseCodes:    make(map[string]float64),
				GRPCRequestsSuccess:  0,
				GRPCRequestsError:    0,
				GRPCRequestDuration:  0,
				GRPCSentMessages:     0,
				GRPCReceivedMessages: 0,
				HTTPResponseCodes:    make(map[string]float64),
				HTTPRequestsSuccess:  0,
				HTTPRequestsError:    0,
				HTTPRequestDuration:  0,
				TCPSentBytes:         0,
				TCPReceivedBytes:     0,
			}
		}

		if _, ok := edges[serviceToWorkloadId]; !ok {
			edges[serviceToWorkloadId] = models.Edge{
				ID:                   serviceToWorkloadId,
				Source:               serviceToWorkloadSource,
				SourceType:           serviceToServiceSourceType,
				SourceName:           serviceToServiceSourceName,
				SourceNamespace:      serviceToServiceSourceNamespace,
				Target:               serviceToWorkloadTarget,
				TargetType:           serviceToServiceTargetType,
				TargetName:           serviceToServiceTargetName,
				TargetNamespace:      serviceToServiceTargetNamespace,
				TargetService:        serviceToServiceTargetService,
				GRPCResponseCodes:    make(map[string]float64),
				GRPCRequestsSuccess:  0,
				GRPCRequestsError:    0,
				GRPCRequestDuration:  0,
				GRPCSentMessages:     0,
				GRPCReceivedMessages: 0,
				HTTPResponseCodes:    make(map[string]float64),
				HTTPRequestsSuccess:  0,
				HTTPRequestsError:    0,
				HTTPRequestDuration:  0,
				TCPSentBytes:         0,
				TCPReceivedBytes:     0,
			}
		}

		if workloadToService, ok := edges[workloadToServiceId]; ok {
			if serviceToWorkloadEdge, ok := edges[serviceToWorkloadId]; ok {
				switch m.Labels["metric"] {
				case models.MetricGRPCRequests:
					workloadToService.GRPCResponseCodes[m.Labels["grpc_response_status"]] += m.Value
					serviceToWorkloadEdge.GRPCResponseCodes[m.Labels["grpc_response_status"]] += m.Value
				case models.MetricGRPCRequestDuration:
					workloadToService.GRPCRequestDuration = (workloadToService.GRPCRequestDuration + m.Value) / 2
					serviceToWorkloadEdge.GRPCRequestDuration = (serviceToWorkloadEdge.GRPCRequestDuration + m.Value) / 2
				case models.MetricGRPCSentMessages:
					workloadToService.GRPCSentMessages += m.Value
					serviceToWorkloadEdge.GRPCSentMessages += m.Value
				case models.MetricGRPCReceivedMessages:
					workloadToService.GRPCReceivedMessages += m.Value
					serviceToWorkloadEdge.GRPCReceivedMessages += m.Value
				case models.MetricHTTPRequests:
					workloadToService.HTTPResponseCodes[m.Labels["response_code"]] += m.Value
					serviceToWorkloadEdge.HTTPResponseCodes[m.Labels["response_code"]] += m.Value
				case models.MetricHTTPRequestDuration:
					workloadToService.HTTPRequestDuration = (workloadToService.HTTPRequestDuration + m.Value) / 2
					serviceToWorkloadEdge.HTTPRequestDuration = (serviceToWorkloadEdge.HTTPRequestDuration + m.Value) / 2
				case models.MetricTCPSentBytes:
					workloadToService.TCPSentBytes += m.Value
					serviceToWorkloadEdge.TCPSentBytes += m.Value
				case models.MetricTCPReceivedBytes:
					workloadToService.TCPReceivedBytes += m.Value
					serviceToWorkloadEdge.TCPReceivedBytes += m.Value
				}

				edges[workloadToServiceId] = workloadToService
				edges[serviceToWorkloadId] = serviceToWorkloadEdge
			}
		}
	}

	edgesSlice := make([]models.Edge, 0, len(edges))
	for _, edge := range edges {
		for code, count := range edge.GRPCResponseCodes {
			if code == "2" || code == "4" || code == "12" || code == "13" || code == "14" || code == "15" {
				edge.GRPCRequestsError += count
			} else {
				edge.GRPCRequestsSuccess += count
			}
		}
		for code, count := range edge.HTTPResponseCodes {
			if code[0] == '5' {
				edge.HTTPRequestsError += count
			} else {
				edge.HTTPRequestsSuccess += count
			}
		}

		edgesSlice = append(edgesSlice, edge)
	}

	return edgesSlice
}

func (d *Datasource) edgesToNodes(edges []models.Edge) []models.Edge {
	nodes := make(map[string]models.Edge)

	for _, e := range edges {
		nodeId := e.Target
		nodeType := e.TargetType
		nodeName := e.TargetName
		nodeNamespace := e.TargetNamespace
		nodeService := e.TargetService

		if _, ok := nodes[nodeId]; !ok {
			nodes[nodeId] = models.Edge{
				ID:                   nodeId,
				TargetType:           nodeType,
				TargetName:           nodeName,
				TargetNamespace:      nodeNamespace,
				TargetService:        nodeService,
				GRPCResponseCodes:    make(map[string]float64),
				GRPCRequestsSuccess:  0,
				GRPCRequestsError:    0,
				GRPCRequestDuration:  0,
				GRPCSentMessages:     0,
				GRPCReceivedMessages: 0,
				HTTPResponseCodes:    make(map[string]float64),
				HTTPRequestsSuccess:  0,
				HTTPRequestsError:    0,
				HTTPRequestDuration:  0,
				TCPSentBytes:         0,
				TCPReceivedBytes:     0,
			}
		}

		if node, ok := nodes[nodeId]; ok {
			for code, count := range e.GRPCResponseCodes {
				node.GRPCResponseCodes[code] += count
			}
			node.GRPCRequestsSuccess += e.GRPCRequestsSuccess
			node.GRPCRequestsError += e.GRPCRequestsError
			if e.GRPCRequestDuration > 0 {
				if node.GRPCRequestDuration == 0 {
					node.GRPCRequestDuration = e.GRPCRequestDuration
				} else {
					node.GRPCRequestDuration = (node.GRPCRequestDuration + e.GRPCRequestDuration) / 2
				}
			}
			node.GRPCSentMessages += e.GRPCSentMessages
			node.GRPCReceivedMessages += e.GRPCReceivedMessages
			for code, count := range e.HTTPResponseCodes {
				node.HTTPResponseCodes[code] += count
			}
			node.HTTPRequestsSuccess += e.HTTPRequestsSuccess
			node.HTTPRequestsError += e.HTTPRequestsError
			if e.HTTPRequestDuration > 0 {
				if node.HTTPRequestDuration == 0 {
					node.HTTPRequestDuration = e.HTTPRequestDuration
				} else {
					node.HTTPRequestDuration = (node.HTTPRequestDuration + e.HTTPRequestDuration) / 2
				}
			}
			node.TCPSentBytes += e.TCPSentBytes
			node.TCPReceivedBytes += e.TCPReceivedBytes

			nodes[nodeId] = node
		}
	}

	sourceIds := make(map[string]bool)
	for _, e := range edges {
		nodeId := e.Source
		nodeType := e.SourceType
		nodeName := e.SourceName
		nodeNamespace := e.SourceNamespace

		if _, ok := nodes[nodeId]; !ok {
			sourceIds[nodeId] = true
			nodes[nodeId] = models.Edge{
				ID:                   nodeId,
				TargetType:           nodeType,
				TargetName:           nodeName,
				TargetNamespace:      nodeNamespace,
				GRPCResponseCodes:    make(map[string]float64),
				GRPCRequestsSuccess:  0,
				GRPCRequestsError:    0,
				GRPCRequestDuration:  0,
				GRPCSentMessages:     0,
				GRPCReceivedMessages: 0,
				HTTPResponseCodes:    make(map[string]float64),
				HTTPRequestsSuccess:  0,
				HTTPRequestsError:    0,
				HTTPRequestDuration:  0,
				TCPSentBytes:         0,
				TCPReceivedBytes:     0,
			}
		}

		if _, isSource := sourceIds[nodeId]; !isSource {
			continue
		}

		if node, ok := nodes[nodeId]; ok {
			for code, count := range e.GRPCResponseCodes {
				node.GRPCResponseCodes[code] += count
			}
			node.GRPCRequestsSuccess += e.GRPCRequestsSuccess
			node.GRPCRequestsError += e.GRPCRequestsError
			node.GRPCRequestDuration = (node.GRPCRequestDuration + e.GRPCRequestDuration) / 2
			node.GRPCSentMessages += e.GRPCSentMessages
			node.GRPCReceivedMessages += e.GRPCReceivedMessages
			for code, count := range e.HTTPResponseCodes {
				node.HTTPResponseCodes[code] += count
			}
			node.HTTPRequestsSuccess += e.HTTPRequestsSuccess
			node.HTTPRequestsError += e.HTTPRequestsError
			node.HTTPRequestDuration = (node.HTTPRequestDuration + e.HTTPRequestDuration) / 2
			node.TCPSentBytes += e.TCPSentBytes
			node.TCPReceivedBytes += e.TCPReceivedBytes

			nodes[nodeId] = node
		}
	}

	nodesSlice := make([]models.Edge, 0, len(edges))
	for _, node := range nodes {
		nodesSlice = append(nodesSlice, node)
	}

	return nodesSlice
}

func getEdgeField(edge models.Edge, interval float64) models.EdgeField {
	edgeField := models.EdgeField{}
	edgeField.ID = edge.ID
	edgeField.Source = edge.Source
	edgeField.Target = edge.Target

	var grpcErrRate float64
	var httpErrRate float64

	if edge.GRPCRequestsSuccess+edge.GRPCRequestsError > 0 {
		edgeField.DetailsGRPCRate = fmt.Sprintf("%.2frps", (edge.GRPCRequestsSuccess+edge.GRPCRequestsError)/interval)
		if edge.GRPCRequestsError > 0 {
			grpcErrRate = (edge.GRPCRequestsError / (edge.GRPCRequestsSuccess + edge.GRPCRequestsError)) * 100
			edgeField.DetailsGRPCErr = fmt.Sprintf("%.2f%%", grpcErrRate)
		}
	}

	if edge.GRPCRequestDuration > 0 {
		edgeField.DetailsGRPCDuration = fmt.Sprintf("%.2fms", edge.GRPCRequestDuration)
	}

	if edge.GRPCSentMessages > 0 {
		edgeField.DetailsGRPCSentMessages = fmt.Sprintf("%.2fmps", edge.GRPCSentMessages/interval)
	}

	if edge.GRPCReceivedMessages > 0 {
		edgeField.DetailsGRPCReceivedMessages = fmt.Sprintf("%.2fmps", edge.GRPCReceivedMessages/interval)
	}

	if edge.HTTPRequestsSuccess+edge.HTTPRequestsError > 0 {
		edgeField.DetailsHTTPRate = fmt.Sprintf("%.2frps", (edge.HTTPRequestsSuccess+edge.HTTPRequestsError)/interval)
		if edge.HTTPRequestsError > 0 {
			httpErrRate = (edge.HTTPRequestsError / (edge.HTTPRequestsSuccess + edge.HTTPRequestsError)) * 100
			edgeField.DetailsHTTPErr = fmt.Sprintf("%.2f%%", httpErrRate)
		}
	}

	if edge.HTTPRequestDuration > 0 {
		edgeField.DetailsHTTPDuration = fmt.Sprintf("%.2fms", edge.HTTPRequestDuration)
	}

	if edge.TCPSentBytes > 0 {
		edgeField.DetailsTCPSentBytes = fmt.Sprintf("%.2fbps", edge.TCPSentBytes/interval)
	}

	if edge.TCPReceivedBytes > 0 {
		edgeField.DetailsTCPReceivedBytes = fmt.Sprintf("%.2fbps", edge.TCPReceivedBytes/interval)
	}

	if edge.HTTPRequestsSuccess+edge.HTTPRequestsError > edge.GRPCRequestsSuccess+edge.GRPCRequestsError {
		edgeField.MainStat = append(edgeField.MainStat, edgeField.DetailsHTTPRate)
		if edgeField.DetailsHTTPErr != "" {
			edgeField.MainStat = append(edgeField.MainStat, edgeField.DetailsHTTPErr)
		}

		if httpErrRate >= 5 {
			edgeField.Color = "#f2495c"
		} else if httpErrRate > 0 {
			edgeField.Color = "#fade2a"
		} else {
			edgeField.Color = "#73bf69"
		}

		if edgeField.DetailsHTTPDuration != "" {
			edgeField.SecondaryStat = append(edgeField.SecondaryStat, edgeField.DetailsHTTPDuration)
		}
		if edge.TCPSentBytes+edge.TCPReceivedBytes > 0 {
			edgeField.SecondaryStat = append(edgeField.SecondaryStat, fmt.Sprintf("%.2fbps", (edge.TCPSentBytes+edge.TCPReceivedBytes)/interval))
		}
	} else if edge.GRPCRequestsSuccess+edge.GRPCRequestsError > 0 {
		edgeField.MainStat = append(edgeField.MainStat, edgeField.DetailsGRPCRate)
		if edgeField.DetailsGRPCErr != "" {
			edgeField.MainStat = append(edgeField.MainStat, edgeField.DetailsGRPCErr)
		}

		if grpcErrRate >= 5 {
			edgeField.Color = "#f2495c"
		} else if grpcErrRate > 0 {
			edgeField.Color = "#fade2a"
		} else {
			edgeField.Color = "#73bf69"
		}

		if edgeField.DetailsGRPCDuration != "" {
			edgeField.SecondaryStat = append(edgeField.SecondaryStat, edgeField.DetailsGRPCDuration)
		}
		if edge.TCPSentBytes+edge.TCPReceivedBytes > 0 {
			edgeField.SecondaryStat = append(edgeField.SecondaryStat, fmt.Sprintf("%.2fbps", (edge.TCPSentBytes+edge.TCPReceivedBytes)/interval))
		}
	} else if edge.TCPSentBytes+edge.TCPReceivedBytes > 0 {
		edgeField.MainStat = append(edgeField.MainStat, fmt.Sprintf("%.2fbps", (edge.TCPSentBytes+edge.TCPReceivedBytes)/interval))
		edgeField.Color = "#5794f2"
	} else {
		edgeField.Color = "#ccccdc"
	}

	return edgeField
}
