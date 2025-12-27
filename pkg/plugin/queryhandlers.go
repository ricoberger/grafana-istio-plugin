package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
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
// applications are retrieved from the "destination_app" and "source_app" label.
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

// handleWorkloadQueries handles the queries to get a list of workloads. It uses
// the concurrent package to handle multiple queries in parallel. The workloads
// are retrieved from the "destination_workload" and "source_workload" label.
func (d *Datasource) handleWorkloadsQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleWorkloadsQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleWorkloads, 10)
}

func (d *Datasource) handleWorkloads(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleWorkloads")
	defer span.End()

	var qm models.QueryModelWorkloads
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}

	queries := []prometheus.LabelValuesQuery{{
		Label: "destination_workload",
		Matches: []string{
			fmt.Sprintf("istio_requests_total{destination_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_sent_bytes_total{destination_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_received_bytes_total{destination_workload_namespace=\"%s\"}", qm.Namespace),
		},
	}, {
		Label: "source_workload",
		Matches: []string{
			fmt.Sprintf("istio_requests_total{source_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_sent_bytes_total{source_workload_namespace=\"%s\"}", qm.Namespace),
			fmt.Sprintf("istio_tcp_received_bytes_total{source_workload_namespace=\"%s\"}", qm.Namespace),
		},
	}}

	return d.handelLabelValues(ctx, queries, query.DataQuery.TimeRange)
}

// handleFilterQueries handles the queries to get a list of workloads for a
// namespace, application or workload which can be used asa filters. This means
// which should not be included in the generated graph. It uses the concurrent
// package to handle multiple queries in parallel.
func (d *Datasource) handleFiltersQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleFiltersQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleFilters, 10)
}

func (d *Datasource) handleFilters(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleFilters")
	defer span.End()

	var qm models.QueryModelFilters
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}

	var namespaceLabel string
	var workloadLabel string
	var queries []string

	switch qm.FilterType {
	case "source":
		namespaceLabel = "source_workload_namespace"
		workloadLabel = "source_workload"

		destinationLabel := ""
		if qm.Application != "" {
			destinationLabel = fmt.Sprintf(`, destination_app="%s"`, qm.Application)
		} else if qm.Workload != "" {
			destinationLabel = fmt.Sprintf(`, destination_workload="%s"`, qm.Workload)
		}

		queries = []string{
			fmt.Sprintf("sum(istio_requests_total{destination_workload_namespace=\"%s\" %s}) by (source_workload_namespace, source_workload)", qm.Namespace, destinationLabel),
			fmt.Sprintf("sum(istio_tcp_sent_bytes_total{destination_workload_namespace=\"%s\" %s}) by (source_workload_namespace, source_workload)", qm.Namespace, destinationLabel),
			fmt.Sprintf("sum(istio_tcp_received_bytes_total{destination_workload_namespace=\"%s\" %s}) by (source_workload_namespace, source_workload)", qm.Namespace, destinationLabel),
		}
	case "destination":
		namespaceLabel = "destination_workload_namespace"
		workloadLabel = "destination_workload"

		sourceLabel := ""
		if qm.Application != "" {
			sourceLabel = fmt.Sprintf(`, source_app="%s"`, qm.Application)
		} else if qm.Workload != "" {
			sourceLabel = fmt.Sprintf(`, source_workload="%s"`, qm.Workload)
		}

		queries = []string{
			fmt.Sprintf("sum(istio_requests_total{source_workload_namespace=\"%s\" %s}) by (destination_workload_namespace, destination_workload)", qm.Namespace, sourceLabel),
			fmt.Sprintf("sum(istio_tcp_sent_bytes_total{source_workload_namespace=\"%s\" %s}) by (destination_workload_namespace, destination_workload)", qm.Namespace, sourceLabel),
			fmt.Sprintf("sum(istio_tcp_received_bytes_total{source_workload_namespace=\"%s\" %s}) by (destination_workload_namespace, destination_workload)", qm.Namespace, sourceLabel),
		}
	}

	var errors []error
	errorsMutex := &sync.Mutex{}

	var values []string
	valuesMutex := &sync.Mutex{}

	var queriesWG sync.WaitGroup
	queriesWG.Add(len(queries))

	for _, q := range queries {
		go func(q string) {
			defer queriesWG.Done()

			d.logger.Debug("Get metrics", "query", q, "timeRangeFrom", query.DataQuery.TimeRange.From, "timeRangeTo", query.DataQuery.TimeRange.To)
			metrics, err := d.prometheusClient.GetMetrics(ctx, "", q, query.DataQuery.TimeRange)
			if err != nil {
				d.logger.Error("Failed to get metrics", "error", err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				errorsMutex.Lock()
				errors = append(errors, err)
				errorsMutex.Unlock()
				return
			}
			d.logger.Debug("Retrieved metrics", "query", q, "metrics", metrics)

			var vs []string
			for _, metric := range metrics {
				if namespace, ok := metric.Labels[namespaceLabel]; ok {
					if workload, ok := metric.Labels[workloadLabel]; ok {
						vs = append(vs, fmt.Sprintf("%s/%s", namespace, workload))
					}
				}
			}

			valuesMutex.Lock()
			values = append(values, vs...)
			valuesMutex.Unlock()
		}(q)
	}

	queriesWG.Wait()

	if len(errors) > 0 {
		span.RecordError(errors[0])
		span.SetStatus(codes.Error, errors[0].Error())
		return backend.ErrorResponseWithErrorSource(errors[0])
	}

	slices.Sort(values)
	values = slices.Compact(values)

	frame := data.NewFrame(
		"Values",
		data.NewField("values", nil, values),
	)

	frame.SetMeta(&data.FrameMeta{
		PreferredVisualization: data.VisTypeTable,
		Type:                   data.FrameTypeTable,
	})

	var response backend.DataResponse
	response.Frames = append(response.Frames, frame)

	return response
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

// handleApplicationGraphQueries handles the queries to get graph for an
// application. It uses the concurrent package to handle multiple queries in
// parallel. The graph is generated for a specific application in a namespace
// and contains all the requested metrics.
func (d *Datasource) handleApplicationGraphQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleApplicationGraphQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleApplicationGraph, 10)
}

func (d *Datasource) handleApplicationGraph(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleApplicationGraph")
	defer span.End()

	var qm models.QueryModelApplicationGraph
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}

	return d.handleGraph(ctx, qm.Namespace, qm.Application, "", qm.Metrics, qm.SourceFilters, qm.DestinationFilters, qm.IdleEdges, query.DataQuery.TimeRange)
}

// handleWorkloadGraphQueries handles the queries to get graph for a workload.
// It uses the concurrent package to handle multiple queries in parallel. The
// graph is generated for a specific workload in a namespace and contains all
// the requested metrics.
func (d *Datasource) handleWorkloadGraphQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleWorkloadGraphQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleWorkloadGraph, 10)
}

func (d *Datasource) handleWorkloadGraph(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleWorkloadGraph")
	defer span.End()

	var qm models.QueryModelWorkloadGraph
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}

	return d.handleGraph(ctx, qm.Namespace, "", qm.Workload, qm.Metrics, qm.SourceFilters, qm.DestinationFilters, qm.IdleEdges, query.DataQuery.TimeRange)
}

// handleNamespaceGraphQueries handles the queries to get graph for a namespace.
// It uses the concurrent package to handle multiple queries in parallel. The
// graph is generated for a specific namespace and contains all the requested
// metrics.
func (d *Datasource) handleNamespaceGraphQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleNamespaceGraphQueries")
	defer span.End()

	return concurrent.QueryData(ctx, req, d.handleNamespaceGraph, 10)
}

func (d *Datasource) handleNamespaceGraph(ctx context.Context, query concurrent.Query) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleNamespaceGraph")
	defer span.End()

	var qm models.QueryModelNamespaceGraph
	err := json.Unmarshal(query.DataQuery.JSON, &qm)
	if err != nil {
		d.logger.Error("Failed to unmarshal query model", "error", err.Error())
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return backend.ErrorResponseWithErrorSource(err)
	}

	return d.handleGraph(ctx, qm.Namespace, "", "", qm.Metrics, qm.SourceFilters, qm.DestinationFilters, qm.IdleEdges, query.DataQuery.TimeRange)
}

// handleGraph creates the graph for the given namespace, application or
// workload. The function can be used for all the three graph types we support.
// It retrieves all the requested metrics, generates the edges and nodes based
// on the metrics and returns the graph as data frames.
func (d *Datasource) handleGraph(ctx context.Context, namespace, application, workload string, metrics, sourceFilters, destinationFilters []string, idleEdges bool, timeRange backend.TimeRange) backend.DataResponse {
	ctx, span := tracing.DefaultTracer().Start(ctx, "handleGraph")
	defer span.End()

	interval := int64(timeRange.Duration().Seconds())

	var errors []error
	errorsMutex := &sync.Mutex{}

	var prometheusMetrics []prometheus.Metric
	prometheusMetricsMutex := &sync.Mutex{}

	var metricsWG sync.WaitGroup
	metricsWG.Add(len(metrics))

	// Get all metrics in parallel for the given namespace, application or
	// workload. We need to get the metrics where the namespace / application /
	// workload is the detination orthe source to build the full graph.
	for _, metric := range metrics {
		go func(metric string) {
			defer metricsWG.Done()

			d.logger.Debug("Get metric", "metric", metric, "namespace", namespace, "application", application, "workload", workload, "timeRangeFrom", timeRange.From, "timeRangeTo", timeRange.To, "interval", interval)

			destinationMetrics, err := d.prometheusClient.GetMetrics(ctx, metric, d.metricToPrometheusDestinationsQuery(namespace, application, workload, metric, idleEdges, interval), timeRange)
			if err != nil {
				d.logger.Error("Failed to get metric", "error", err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				errorsMutex.Lock()
				errors = append(errors, err)
				errorsMutex.Unlock()
				return
			}
			d.logger.Debug("Retrieved metrics where application is destination", "metric", metric, "namespace", namespace, "application", application, "workload", workload, "metrics", destinationMetrics)

			sourceMetrics, err := d.prometheusClient.GetMetrics(ctx, metric, d.metricToPrometheusSourcesQuery(namespace, application, workload, metric, idleEdges, interval), timeRange)
			if err != nil {
				d.logger.Error("Failed to get metric", "error", err.Error())
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())

				errorsMutex.Lock()
				errors = append(errors, err)
				errorsMutex.Unlock()
				return
			}
			d.logger.Debug("Retrieved metrics where application is source", "metric", metric, "namespace", namespace, "application", application, "workload", workload, "metrics", sourceMetrics)

			prometheusMetricsMutex.Lock()
			prometheusMetrics = append(prometheusMetrics, destinationMetrics...)
			prometheusMetrics = append(prometheusMetrics, sourceMetrics...)
			prometheusMetricsMutex.Unlock()
		}(metric)
	}

	metricsWG.Wait()

	if len(errors) > 0 {
		span.RecordError(errors[0])
		span.SetStatus(codes.Error, errors[0].Error())
		return backend.ErrorResponseWithErrorSource(errors[0])
	}

	// Deduplicate the metrics (metrics where all labels are the same), generate
	// the edges based on the metrics and then generate the nodes based on the
	// edges.
	prometheusMetrics = d.deduplicateMetrics(prometheusMetrics)
	edges := d.metricsToEdges(prometheusMetrics, sourceFilters, destinationFilters)
	nodes := d.edgesToNodes(edges)

	// Generate the data frames for the edges and nodes, the data for the
	// "details__*" fields is generated using the "getEdgeField" and
	// "getNodeField" functions.
	edgeFields := models.Fields{}
	edgeIds := edgeFields.Add("id", nil, []string{})
	edgeSources := edgeFields.Add("source", nil, []string{})
	edgeDestinations := edgeFields.Add("target", nil, []string{})
	edgeMainStat := edgeFields.Add("mainstat", nil, []string{}, &data.FieldConfig{DisplayName: "Main Stats"})
	edgeSecondaryStat := edgeFields.Add("secondarystat", nil, []string{}, &data.FieldConfig{DisplayName: "Secondary Stats"})
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
		edgeField := d.getEdgeField(edge, float64(interval))

		edgeIds.Append(edgeField.ID)
		edgeSources.Append(edgeField.Source)
		edgeDestinations.Append(edgeField.Destination)
		edgeMainStat.Append(strings.Join(edgeField.MainStat, " | "))
		edgeSecondaryStat.Append(strings.Join(edgeField.SecondaryStat, " | "))
		edgeColors.Append(edgeField.Color)
		edgeDetailsGRPCRate.Append(strings.Join(edgeField.DetailsGRPCRate, " | "))
		edgeDetailsGRPCErr.Append(strings.Join(edgeField.DetailsGRPCErr, " | "))
		edgeDetailsGRPCDuration.Append(strings.Join(edgeField.DetailsGRPCDuration, " | "))
		edgeDetailsGRPCSentMessages.Append(strings.Join(edgeField.DetailsGRPCSentMessages, " | "))
		edgeDetailsGRPCReceivedMessages.Append(strings.Join(edgeField.DetailsGRPCReceivedMessages, " | "))
		edgeDetailsHTTPRate.Append(strings.Join(edgeField.DetailsHTTPRate, " | "))
		edgeDetailsHTTPErr.Append(strings.Join(edgeField.DetailsHTTPErr, " | "))
		edgeDetailsHTTPDuration.Append(strings.Join(edgeField.DetailsHTTPDuration, " | "))
		edgeDetailsTCPSentBytes.Append(strings.Join(edgeField.DetailsTCPSentBytes, " | "))
		edgeDetailsTCPReceivedBytes.Append(strings.Join(edgeField.DetailsTCPReceivedBytes, " | "))
	}

	nodeFields := models.Fields{}
	nodeIds := nodeFields.Add("id", nil, []string{})
	nodeTitles := nodeFields.Add("title", nil, []string{}, &data.FieldConfig{DisplayName: "Type"})
	nodeSubTitles := nodeFields.Add("subtitle", nil, []string{}, &data.FieldConfig{DisplayName: "Name (Namespace)"})
	nodeMainStat := nodeFields.Add("mainstat", nil, []string{}, &data.FieldConfig{DisplayName: "Main Stats"})
	nodeSecondaryStat := nodeFields.Add("secondarystat", nil, []string{}, &data.FieldConfig{DisplayName: "Secondary Stats"})
	nodeColors := nodeFields.Add("color", nil, []string{}, &data.FieldConfig{DisplayName: "Health"})
	nodeDetailsGRPCRate := nodeFields.Add("detail__grpcrate", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Rate"})
	nodeDetailsGRPCErr := nodeFields.Add("detail__grpcperr", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Error"})
	nodeDetailsGRPCSentMessages := nodeFields.Add("detail__grpcsentmessages", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Sent Messages"})
	nodeDetailsGRPCReceivedMessages := nodeFields.Add("detail__grpcreceivedmessages", nil, []string{}, &data.FieldConfig{DisplayName: "gRPC Received Messages"})
	nodeDetailsHTTPRate := nodeFields.Add("detail__httprate", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Rate"})
	nodeDetailsHTTPErr := nodeFields.Add("detail__httperr", nil, []string{}, &data.FieldConfig{DisplayName: "HTTP Error"})
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
		nodeField := d.getNodeField(node, float64(interval))

		nodeIds.Append(nodeField.ID)
		nodeTitles.Append(node.Type)
		nodeSubTitles.Append(fmt.Sprintf("%s (%s)", node.Name, node.Namespace))
		nodeMainStat.Append(strings.Join(nodeField.MainStat, " | "))
		nodeSecondaryStat.Append(strings.Join(nodeField.SecondaryStat, " | "))
		nodeColors.Append(nodeField.Color)
		nodeDetailsGRPCRate.Append(strings.Join(nodeField.DetailsGRPCRate, " | "))
		nodeDetailsGRPCErr.Append(strings.Join(nodeField.DetailsGRPCErr, " | "))
		nodeDetailsGRPCSentMessages.Append(strings.Join(nodeField.DetailsGRPCSentMessages, " | "))
		nodeDetailsGRPCReceivedMessages.Append(strings.Join(nodeField.DetailsGRPCReceivedMessages, " | "))
		nodeDetailsHTTPRate.Append(strings.Join(nodeField.DetailsHTTPRate, " | "))
		nodeDetailsHTTPErr.Append(strings.Join(nodeField.DetailsHTTPErr, " | "))
		nodeDetailsTCPSentBytes.Append(strings.Join(nodeField.DetailsTCPSentBytes, " | "))
		nodeDetailsTCPReceivedBytes.Append(strings.Join(nodeField.DetailsTCPReceivedBytes, " | "))

		// Depending on the node type we link to the appropriate Istio dashboard
		// with the correct variables set.
		// - Service dashboard: https://grafana.com/grafana/dashboards/7636-istio-service-dashboard/
		// - Workload dashboard: https://grafana.com/grafana/dashboards/7630-istio-workload-dashboard/
		switch node.Type {
		case "Service":
			nodeLink.Append(fmt.Sprintf("%s&var-service=%s&from=%d&to=%d", d.istioServiceDashboard, node.Service, timeRange.From.UnixMilli(), timeRange.To.UnixMilli()))
		case "Workload":
			nodeLink.Append(fmt.Sprintf("%s&var-namespace=%s&var-workload=%s&from=%d&to=%d", d.istioWorkloadDashboard, node.Namespace, node.Name, timeRange.From.UnixMilli(), timeRange.To.UnixMilli()))
		default:
			nodeLink.Append("")
		}
	}

	// Generate the backend data response with the edge and node data frames.
	// Alos set the preferred visualization to "node graph" for both frames, so
	// that Grafana knows how to visualize them.
	edgeFrame := data.NewFrame("edges", edgeFields...).SetMeta(&data.FrameMeta{PreferredVisualization: data.VisTypeNodeGraph})
	nodeFrame := data.NewFrame("nodes", nodeFields...).SetMeta(&data.FrameMeta{PreferredVisualization: data.VisTypeNodeGraph})

	var response backend.DataResponse
	response.Frames = append(response.Frames, edgeFrame)
	response.Frames = append(response.Frames, nodeFrame)

	return response
}

// metricToPrometheusDestinationsQuery generates the Prometheus query for the
// given metric where the application or workload is the destination.
//
// If the "idleEdges" parameter is set to true, the query will also include
// edges with zero traffic. Otherwise, these edges will be filtered out using a
// "> 0" operator.
//
// If the "application" parameter is set, the query will filter by the
// "destination_app" label. If the "workload" parameter is set, the query will
// filter by the "destination_workload" label.
func (d *Datasource) metricToPrometheusDestinationsQuery(namespace, application, workload, metric string, idleEdges bool, interval int64) string {
	operator := "> 0"
	if idleEdges {
		operator = ""
	}

	destinationLabel := ""
	if application != "" {
		destinationLabel = fmt.Sprintf(`, destination_app="%s"`, application)
	} else if workload != "" {
		destinationLabel = fmt.Sprintf(`, destination_workload="%s"`, workload)
	}

	switch metric {
	case models.MetricGRPCRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{destination_workload_namespace="%s", request_protocol="grpc" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status) %s`, namespace, destinationLabel, interval, operator)
	case models.MetricGRPCRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{destination_workload_namespace="%s", request_protocol="grpc" %s}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)) %s`, namespace, destinationLabel, interval, operator)
	case models.MetricGRPCSentMessages:
		return fmt.Sprintf(`sum(increase(istio_request_messages_total{destination_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, destinationLabel, interval, operator)
	case models.MetricGRPCReceivedMessages:
		return fmt.Sprintf(`sum(increase(istio_response_messages_total{destination_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, destinationLabel, interval, operator)
	case models.MetricHTTPRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{destination_workload_namespace="%s", request_protocol="http" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, response_code) %s`, namespace, destinationLabel, interval, operator)
	case models.MetricHTTPRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{destination_workload_namespace="%s", request_protocol="http" %s}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)) %s`, namespace, destinationLabel, interval, operator)
	case models.MetricTCPSentBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_sent_bytes_total{destination_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, destinationLabel, interval, operator)
	case models.MetricTCPReceivedBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_received_bytes_total{destination_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, destinationLabel, interval, operator)
	default:
		return ""
	}
}

// metricToPrometheusSourcesQuery generates the Prometheus query for the given
// metric where the application or workload is the source.
//
// If the "idleEdges" parameter is set to true, the query will also include
// edges with zero traffic. Otherwise, these edges will be filtered out using a
// "> 0" operator.
//
// If the "application" parameter is set, the query will filter by the
// "source_app" label. If the "workload" parameter is set, the query will
// filter by the "source_workload" label.
func (d *Datasource) metricToPrometheusSourcesQuery(namespace, application, workload, metric string, idleEdges bool, interval int64) string {
	operator := "> 0"
	if idleEdges {
		operator = ""
	}

	sourceLabel := ""
	if application != "" {
		sourceLabel = fmt.Sprintf(`, source_app="%s"`, application)
	} else if workload != "" {
		sourceLabel = fmt.Sprintf(`, source_workload="%s"`, workload)
	}

	switch metric {
	case models.MetricGRPCRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{source_workload_namespace="%s", request_protocol="grpc" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, grpc_response_status) %s`, namespace, sourceLabel, interval, operator)
	case models.MetricGRPCRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{source_workload_namespace="%s", request_protocol="grpc" %s}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)) %s`, namespace, sourceLabel, interval, operator)
	case models.MetricGRPCSentMessages:
		return fmt.Sprintf(`sum(increase(istio_request_messages_total{source_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, sourceLabel, interval, operator)
	case models.MetricGRPCReceivedMessages:
		return fmt.Sprintf(`sum(increase(istio_response_messages_total{source_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, sourceLabel, interval, operator)
	case models.MetricHTTPRequests:
		return fmt.Sprintf(`sum(increase(istio_requests_total{source_workload_namespace="%s", request_protocol="http" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload, response_code) %s`, namespace, sourceLabel, interval, operator)
	case models.MetricHTTPRequestDuration:
		return fmt.Sprintf(`histogram_quantile(0.99, sum(increase(istio_request_duration_milliseconds_bucket{source_workload_namespace="%s", request_protocol="http" %s}[%ds])) by (le, destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload)) %s`, namespace, sourceLabel, interval, operator)
	case models.MetricTCPSentBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_sent_bytes_total{source_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, sourceLabel, interval, operator)
	case models.MetricTCPReceivedBytes:
		return fmt.Sprintf(`sum(increase(istio_tcp_received_bytes_total{source_workload_namespace="%s" %s}[%ds])) by (destination_service, destination_service_namespace, destination_service_name, destination_workload_namespace, destination_workload, destination_version, source_workload_namespace, source_workload) %s`, namespace, sourceLabel, interval, operator)
	default:
		return ""
	}
}

// depuplicateMetrics removes duplicate metrics from the given slice of
// Prometheus metrics. Two metrics are considered duplicates if they have the
// same labels.
func (d *Datasource) deduplicateMetrics(metrics []prometheus.Metric) []prometheus.Metric {
	var result []prometheus.Metric

	for _, m := range metrics {
		isDuplicate := false
		for _, r := range result {
			if reflect.DeepEqual(m.Labels, r.Labels) {
				isDuplicate = true
				break
			}
		}
		if !isDuplicate {
			result = append(result, m)
		}
	}
	return result
}

// Generate the edges from the given Prometheus metrics. The edges are filtered
// based on the given source and destination filters. If a source workload or
// destination workload matches any of the filters, the edge is skipped.
func (d *Datasource) metricsToEdges(metrics []prometheus.Metric, sourceFilters, destinationFilters []string) map[string]models.Edge {
	edges := make(map[string]models.Edge)

	for _, m := range metrics {
		if slices.Contains(sourceFilters, fmt.Sprintf("%s/%s", m.Labels["source_workload_namespace"], m.Labels["source_workload"])) || slices.Contains(destinationFilters, fmt.Sprintf("%s/%s", m.Labels["destination_workload_namespace"], m.Labels["destination_workload"])) {
			continue
		}

		var tmpEdges []models.Edge

		// If the source or destination workload is a waypoint, create a direct
		// edge between the source and destination workloads. Otherwise, create
		// one edge from the source wrokload to the destination service and from
		// the destination service to the destination workload.
		if m.Labels["source_workload"] == "waypoint" || m.Labels["destination_workload"] == "waypoint" {
			tmpEdges = []models.Edge{{
				ID:                   fmt.Sprintf("workload-%s-%s-workload-%s-%s", m.Labels["source_workload"], m.Labels["source_workload_namespace"], m.Labels["destination_service_name"], m.Labels["destination_service_namespace"]),
				Source:               fmt.Sprintf("Workload: %s (%s)", m.Labels["source_workload"], m.Labels["source_workload_namespace"]),
				SourceType:           "Workload",
				SourceName:           m.Labels["source_workload"],
				SourceNamespace:      m.Labels["source_workload_namespace"],
				Destination:          fmt.Sprintf("Workload: %s (%s)", m.Labels["destination_workload"], m.Labels["destination_workload_namespace"]),
				DestinationType:      "Workload",
				DestinationName:      m.Labels["destination_workload"],
				DestinationNamespace: m.Labels["destination_workload_namespace"],
				DestinationService:   m.Labels["destination_service"],
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
			}}
		} else {
			tmpEdges = []models.Edge{{
				ID:                   fmt.Sprintf("workload-%s-%s-service-%s-%s", m.Labels["source_workload"], m.Labels["source_workload_namespace"], m.Labels["destination_service_name"], m.Labels["destination_service_namespace"]),
				Source:               fmt.Sprintf("Workload: %s (%s)", m.Labels["source_workload"], m.Labels["source_workload_namespace"]),
				SourceType:           "Workload",
				SourceName:           m.Labels["source_workload"],
				SourceNamespace:      m.Labels["source_workload_namespace"],
				Destination:          fmt.Sprintf("Service: %s (%s)", m.Labels["destination_service_name"], m.Labels["destination_service_namespace"]),
				DestinationType:      "Service",
				DestinationName:      m.Labels["destination_service_name"],
				DestinationNamespace: m.Labels["destination_service_namespace"],
				DestinationService:   m.Labels["destination_service"],
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
			}, {
				ID:                   fmt.Sprintf("service-%s-%s-workload-%s-%s", m.Labels["destination_service_name"], m.Labels["destination_service_namespace"], m.Labels["destination_workload"], m.Labels["destination_workload_namespace"]),
				Source:               fmt.Sprintf("Service: %s (%s)", m.Labels["destination_service_name"], m.Labels["destination_service_namespace"]),
				SourceType:           "Service",
				SourceName:           m.Labels["destination_service_name"],
				SourceNamespace:      m.Labels["destination_service_namespace"],
				Destination:          fmt.Sprintf("Workload: %s (%s)", m.Labels["destination_workload"], m.Labels["destination_workload_namespace"]),
				DestinationType:      "Workload",
				DestinationName:      m.Labels["destination_workload"],
				DestinationNamespace: m.Labels["destination_workload_namespace"],
				DestinationService:   m.Labels["destination_service"],
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
			}}
		}

		// Go though all the temporary edges and aggregate the metrics into the
		// final edges map. Each edge is identified by its id. If an edge
		// with the same id already exists, we aggregate the metrics into the
		// existing edge.
		//
		// Notes:
		// - For request counts (gRPC and HTTP) we aggregate the counts based
		//   on the response codes. We also keep track of the total success
		//   and error counts.
		// - A gRPC error is considered to be any response where the
		//   "grpc_response_status" label is 2, 4, 12, 14, 14 or 15. This should
		//   correlate to the HTTP status codes 5xx (see
		//   https://gist.github.com/hamakn/708b9802ca845eb59f3975dbb3ae2a01).
		// - A HTTP error is considered to be any response where the response
		//   code starts with 5 (i.e., 5xx).
		// - For durations we take the latest value and only set it for edges
		//   where the destination type is "Service", because for the edges from
		//   services to workloads the duration depends on the source workload
		//   and I think it doesn't make sens to aggregate them.
		for _, edge := range tmpEdges {
			if _, ok := edges[edge.ID]; !ok {
				edges[edge.ID] = edge
			}

			if existingEdge, ok := edges[edge.ID]; ok {
				switch m.Labels["metric"] {
				case models.MetricGRPCRequests:
					code := m.Labels["grpc_response_status"]
					value := m.Value
					existingEdge.GRPCResponseCodes[code] += value
					if code == "2" || code == "4" || code == "12" || code == "13" || code == "14" || code == "15" {
						existingEdge.GRPCRequestsError += value
					} else {
						existingEdge.GRPCRequestsSuccess += value
					}
				case models.MetricGRPCRequestDuration:
					if existingEdge.DestinationType == "Service" && m.Value > 0 {
						existingEdge.GRPCRequestDuration = m.Value
					}
				case models.MetricGRPCSentMessages:
					existingEdge.GRPCSentMessages += m.Value
				case models.MetricGRPCReceivedMessages:
					existingEdge.GRPCReceivedMessages += m.Value
				case models.MetricHTTPRequests:
					code := m.Labels["response_code"]
					value := m.Value
					existingEdge.HTTPResponseCodes[code] += value
					if code[0] == '5' {
						existingEdge.HTTPRequestsError += value
					} else {
						existingEdge.HTTPRequestsSuccess += value
					}
				case models.MetricHTTPRequestDuration:
					if existingEdge.DestinationType == "Service" && m.Value > 0 {
						existingEdge.HTTPRequestDuration = m.Value
					}
				case models.MetricTCPSentBytes:
					existingEdge.TCPSentBytes += m.Value
				case models.MetricTCPReceivedBytes:
					existingEdge.TCPReceivedBytes += m.Value
				}

				edges[edge.ID] = existingEdge
			}
		}
	}

	return edges
}

// Generate the nodes from the given edges. The nodes are generated by going
// through all the edges and aggregating the metrics for each node.
func (d *Datasource) edgesToNodes(edges map[string]models.Edge) map[string]models.Node {
	nodes := make(map[string]models.Node)

	// Go through all the edges and generate a two nodes for each edge: one for
	// the source and one for the destination.
	//
	// Notes:
	// - If the node is a source, the edge metrics are added as client metrics.
	//   If the node is a destination, the edge metrics are added as server
	//   metrics.
	// - We ignore the gRPC and HTTP request durations for the nodes, because
	//   aggregating them doesn't make much sense.
	for _, edge := range edges {
		tmpNodes := []models.Node{{
			ID:                         edge.Source,
			Type:                       edge.SourceType,
			Name:                       edge.SourceName,
			Namespace:                  edge.SourceNamespace,
			Service:                    "",
			ClientGRPCResponseCodes:    edge.GRPCResponseCodes,
			ClientGRPCRequestsSuccess:  edge.GRPCRequestsSuccess,
			ClientGRPCRequestsError:    edge.GRPCRequestsError,
			ClientGRPCSentMessages:     edge.GRPCSentMessages,
			ClientGRPCReceivedMessages: edge.GRPCReceivedMessages,
			ClientHTTPResponseCodes:    edge.HTTPResponseCodes,
			ClientHTTPRequestsSuccess:  edge.HTTPRequestsSuccess,
			ClientHTTPRequestsError:    edge.HTTPRequestsError,
			ClientTCPSentBytes:         edge.TCPSentBytes,
			ClientTCPReceivedBytes:     edge.TCPReceivedBytes,
			ServerGRPCResponseCodes:    make(map[string]float64),
			ServerGRPCRequestsSuccess:  0,
			ServerGRPCRequestsError:    0,
			ServerGRPCSentMessages:     0,
			ServerGRPCReceivedMessages: 0,
			ServerHTTPResponseCodes:    make(map[string]float64),
			ServerHTTPRequestsSuccess:  0,
			ServerHTTPRequestsError:    0,
			ServerTCPSentBytes:         0,
			ServerTCPReceivedBytes:     0,
		}, {
			ID:                         edge.Destination,
			Type:                       edge.DestinationType,
			Name:                       edge.DestinationName,
			Namespace:                  edge.DestinationNamespace,
			Service:                    edge.DestinationService,
			ClientGRPCResponseCodes:    make(map[string]float64),
			ClientGRPCRequestsSuccess:  0,
			ClientGRPCRequestsError:    0,
			ClientGRPCSentMessages:     0,
			ClientGRPCReceivedMessages: 0,
			ClientHTTPResponseCodes:    make(map[string]float64),
			ClientHTTPRequestsSuccess:  0,
			ClientHTTPRequestsError:    0,
			ClientTCPSentBytes:         0,
			ClientTCPReceivedBytes:     0,
			ServerGRPCResponseCodes:    edge.GRPCResponseCodes,
			ServerGRPCRequestsSuccess:  edge.GRPCRequestsSuccess,
			ServerGRPCRequestsError:    edge.GRPCRequestsError,
			ServerGRPCSentMessages:     edge.GRPCSentMessages,
			ServerGRPCReceivedMessages: edge.GRPCReceivedMessages,
			ServerHTTPResponseCodes:    edge.HTTPResponseCodes,
			ServerHTTPRequestsSuccess:  edge.HTTPRequestsSuccess,
			ServerHTTPRequestsError:    edge.HTTPRequestsError,
			ServerTCPSentBytes:         edge.TCPSentBytes,
			ServerTCPReceivedBytes:     edge.TCPReceivedBytes,
		}}

		for _, node := range tmpNodes {
			// If the node doesn't exist yet, add it to the nodes map and do not
			// continue, since we do not have to aggregate the values with an
			// existing node.
			if _, ok := nodes[node.ID]; !ok {
				nodes[node.ID] = node
				continue
			}

			if existingNode, ok := nodes[node.ID]; ok {
				for code, count := range node.ClientGRPCResponseCodes {
					existingNode.ClientGRPCResponseCodes[code] += count
				}
				existingNode.ClientGRPCRequestsSuccess += node.ClientGRPCRequestsSuccess
				existingNode.ClientGRPCRequestsError += node.ClientGRPCRequestsError
				existingNode.ClientGRPCSentMessages += node.ClientGRPCSentMessages
				existingNode.ClientGRPCReceivedMessages += node.ClientGRPCReceivedMessages
				for code, count := range node.ClientHTTPResponseCodes {
					existingNode.ClientHTTPResponseCodes[code] += count
				}
				existingNode.ClientHTTPRequestsSuccess += node.ClientHTTPRequestsSuccess
				existingNode.ClientHTTPRequestsError += node.ClientHTTPRequestsError
				existingNode.ClientTCPSentBytes += node.ClientTCPSentBytes
				existingNode.ClientTCPReceivedBytes += node.ClientTCPReceivedBytes

				for code, count := range node.ServerGRPCResponseCodes {
					existingNode.ServerGRPCResponseCodes[code] += count
				}
				existingNode.ServerGRPCRequestsSuccess += node.ServerGRPCRequestsSuccess
				existingNode.ServerGRPCRequestsError += node.ServerGRPCRequestsError
				existingNode.ServerGRPCSentMessages += node.ServerGRPCSentMessages
				existingNode.ServerGRPCReceivedMessages += node.ServerGRPCReceivedMessages
				for code, count := range node.ServerHTTPResponseCodes {
					existingNode.ServerHTTPResponseCodes[code] += count
				}
				existingNode.ServerHTTPRequestsSuccess += node.ServerHTTPRequestsSuccess
				existingNode.ServerHTTPRequestsError += node.ServerHTTPRequestsError
				existingNode.ServerTCPSentBytes += node.ServerTCPSentBytes
				existingNode.ServerTCPReceivedBytes += node.ServerTCPReceivedBytes

				nodes[node.ID] = existingNode
			}
		}
	}

	return nodes
}

// generateEdgeField generates the data frame fields for the give edge. This
// also includes setting the color, main stat and secondary stat.
func (d *Datasource) getEdgeField(edge models.Edge, interval float64) models.Field {
	field := models.Field{}
	field.ID = edge.ID
	field.Source = edge.Source
	field.Destination = edge.Destination

	var grpcErrRate float64
	var httpErrRate float64

	// Set the details metrics for gRPC traffic and save the gRPC error rate
	// for later to use them for setting the color. All metrics are set also
	// when they are zero, except the gRPC request duration, where we use "-",
	// because only edges from a source workload to a destination service have
	// a duration.
	field.DetailsGRPCRate = []string{fmt.Sprintf("%.2frps", (edge.GRPCRequestsSuccess+edge.GRPCRequestsError)/interval)}
	if edge.GRPCRequestsError > 0 {
		grpcErrRate = (edge.GRPCRequestsError / (edge.GRPCRequestsSuccess + edge.GRPCRequestsError)) * 100
		field.DetailsGRPCErr = []string{fmt.Sprintf("%.2f%%", grpcErrRate)}
	} else {
		grpcErrRate = 0
		field.DetailsGRPCErr = []string{fmt.Sprintf("%.2f%%", grpcErrRate)}
	}
	if edge.GRPCRequestDuration > 0 {
		field.DetailsGRPCDuration = []string{fmt.Sprintf("%.2fms", edge.GRPCRequestDuration)}
	} else {
		field.DetailsGRPCDuration = []string{"-"}
	}
	field.DetailsGRPCSentMessages = []string{fmt.Sprintf("%.2fmps", edge.GRPCSentMessages/interval)}
	field.DetailsGRPCReceivedMessages = []string{fmt.Sprintf("%.2fmps", edge.GRPCReceivedMessages/interval)}

	// Set the details metrics for HTTP traffic and save the HTTP error rate
	// for later to use them for setting the color. All metrics are set also
	// when they are zero, except the HTTP request duration, where we use "-",
	// because only edges from a source workload to a destination service have
	// a duration.
	field.DetailsHTTPRate = []string{fmt.Sprintf("%.2frps", (edge.HTTPRequestsSuccess+edge.HTTPRequestsError)/interval)}
	if edge.HTTPRequestsError > 0 {
		httpErrRate = (edge.HTTPRequestsError / (edge.HTTPRequestsSuccess + edge.HTTPRequestsError)) * 100
		field.DetailsHTTPErr = []string{fmt.Sprintf("%.2f%%", httpErrRate)}
	} else {
		httpErrRate = 0
		field.DetailsHTTPErr = []string{fmt.Sprintf("%.2f%%", httpErrRate)}
	}
	if edge.HTTPRequestDuration > 0 {
		field.DetailsHTTPDuration = []string{fmt.Sprintf("%.2fms", edge.HTTPRequestDuration)}
	} else {
		field.DetailsHTTPDuration = []string{"-"}
	}

	// Set the details metrics for TCP traffic.
	field.DetailsTCPSentBytes = []string{fmt.Sprintf("%.2fbps", edge.TCPSentBytes/interval)}
	field.DetailsTCPReceivedBytes = []string{fmt.Sprintf("%.2fbps", edge.TCPReceivedBytes/interval)}

	// Set the color, main stat and secondary stat based on the traffic type:
	// - If there is more HTTP traffic than gRPC traffic, show the HTTP request
	//   rate and error percentage as main stat. The secondary stat is the HTTP
	//   request duration and the TCP traffic rate.
	// - If there is gRPC traffic, show the gRPC request rate and error
	//   percentage as main stat. The secondary stat is the gRPC request
	//   duration and the TCP traffic rate.
	// - If there is only TCP traffic, show the TCP traffic rate as main stat.
	//
	// The color is set as follows:
	// - For HTTP and gRPC traffic, if the error rate is above the error
	//   threshold, the color is red. If the error rate is above the warning
	//   threshold, the color is yellow. Otherwise, the color is green.
	// - For TCP traffic, the color is blue.
	// - If there is no traffic, the color is gray.
	if edge.HTTPRequestsSuccess+edge.HTTPRequestsError > edge.GRPCRequestsSuccess+edge.GRPCRequestsError {
		field.MainStat = append(field.MainStat, field.DetailsHTTPRate[0])
		if httpErrRate > 0 {
			field.MainStat = append(field.MainStat, field.DetailsHTTPErr[0])
		}

		if httpErrRate >= d.istioErrorThreshold {
			field.Color = "#f2495c"
		} else if httpErrRate > d.istioWarningThreshold {
			field.Color = "#fade2a"
		} else {
			field.Color = "#73bf69"
		}

		if edge.HTTPRequestDuration > 0 {
			field.SecondaryStat = append(field.SecondaryStat, field.DetailsHTTPDuration[0])
		}
		if edge.TCPSentBytes+edge.TCPReceivedBytes > 0 {
			field.SecondaryStat = append(field.SecondaryStat, fmt.Sprintf("%.2fbps", (edge.TCPSentBytes+edge.TCPReceivedBytes)/interval))
		}
	} else if edge.GRPCRequestsSuccess+edge.GRPCRequestsError > 0 {
		field.MainStat = append(field.MainStat, field.DetailsGRPCRate[0])
		if grpcErrRate > 0 {
			field.MainStat = append(field.MainStat, field.DetailsGRPCErr[0])
		}

		if grpcErrRate >= d.istioErrorThreshold {
			field.Color = "#f2495c"
		} else if grpcErrRate > d.istioWarningThreshold {
			field.Color = "#fade2a"
		} else {
			field.Color = "#73bf69"
		}

		if edge.GRPCRequestDuration > 0 {
			field.SecondaryStat = append(field.SecondaryStat, field.DetailsGRPCDuration[0])
		}
		if edge.TCPSentBytes+edge.TCPReceivedBytes > 0 {
			field.SecondaryStat = append(field.SecondaryStat, fmt.Sprintf("%.2fbps", (edge.TCPSentBytes+edge.TCPReceivedBytes)/interval))
		}
	} else if edge.TCPSentBytes+edge.TCPReceivedBytes > 0 {
		field.MainStat = append(field.MainStat, fmt.Sprintf("%.2fbps", (edge.TCPSentBytes+edge.TCPReceivedBytes)/interval))
		field.Color = "#5794f2"
	} else {
		field.Color = "#ccccdc"
	}

	return field
}

// generateNodeField generate the data frame fields for the given node. This
// also includes setting the color, main stat and secondary stat.
func (d *Datasource) getNodeField(node models.Node, interval float64) models.Field {
	field := models.Field{}
	field.ID = node.ID

	// If the node is a service, we generate the same stats as we generate for
	// edges, with the traffic were the node acting as a server.
	if node.Type == "Service" {
		return d.getEdgeField(models.Edge{
			ID:                   node.ID,
			Source:               node.ID,
			Destination:          node.ID,
			GRPCRequestsSuccess:  node.ServerGRPCRequestsSuccess,
			GRPCRequestsError:    node.ServerGRPCRequestsError,
			GRPCSentMessages:     node.ServerGRPCSentMessages,
			GRPCReceivedMessages: node.ServerGRPCReceivedMessages,
			HTTPRequestsSuccess:  node.ServerHTTPRequestsSuccess,
			HTTPRequestsError:    node.ServerHTTPRequestsError,
			TCPSentBytes:         node.ServerTCPSentBytes,
			TCPReceivedBytes:     node.ServerTCPReceivedBytes,
		}, interval)
	}

	var grpcServerErrRate float64
	var grpcClientErrRate float64
	var httpServerErrRate float64
	var httpClientErrRate float64

	// Set the details metrics for gRPC traffic. We always display the server
	// traffic first and afterwards the client traffic. All metrics are set also
	// when they are zero.
	field.DetailsGRPCRate = []string{fmt.Sprintf("%.2frps", (node.ServerGRPCRequestsSuccess+node.ServerGRPCRequestsError)/interval), fmt.Sprintf("%.2frps", (node.ClientGRPCRequestsSuccess+node.ClientGRPCRequestsError)/interval)}
	if node.ServerGRPCRequestsError > 0 && node.ClientGRPCRequestsError > 0 {
		grpcServerErrRate = (node.ServerGRPCRequestsError / (node.ServerGRPCRequestsSuccess + node.ServerGRPCRequestsError)) * 100
		grpcClientErrRate = (node.ClientGRPCRequestsError / (node.ClientGRPCRequestsSuccess + node.ClientGRPCRequestsError)) * 100
		field.DetailsGRPCErr = []string{fmt.Sprintf("%.2f%%", grpcServerErrRate), fmt.Sprintf("%.2f%%", grpcClientErrRate)}
	} else if node.ServerGRPCRequestsError > 0 && node.ClientGRPCRequestsError == 0 {
		grpcServerErrRate = (node.ServerGRPCRequestsError / (node.ServerGRPCRequestsSuccess + node.ServerGRPCRequestsError)) * 100
		grpcClientErrRate = 0
		field.DetailsGRPCErr = []string{fmt.Sprintf("%.2f%%", grpcServerErrRate), "0.00%"}
	} else if node.ServerGRPCRequestsError == 0 && node.ClientGRPCRequestsError > 0 {
		grpcServerErrRate = 0
		grpcClientErrRate = (node.ClientGRPCRequestsError / (node.ClientGRPCRequestsSuccess + node.ClientGRPCRequestsError)) * 100
		field.DetailsGRPCErr = []string{"0.00%", fmt.Sprintf("%.2f%%", grpcClientErrRate)}
	} else {
		grpcServerErrRate = 0
		grpcClientErrRate = 0
		field.DetailsGRPCErr = []string{"0.00%", "0.00%"}
	}
	field.DetailsGRPCSentMessages = []string{fmt.Sprintf("%.2fmps", node.ServerGRPCSentMessages/interval), fmt.Sprintf("%.2fmps", node.ClientGRPCSentMessages/interval)}
	field.DetailsGRPCReceivedMessages = []string{fmt.Sprintf("%.2fmps", node.ServerGRPCReceivedMessages/interval), fmt.Sprintf("%.2fmps", node.ClientGRPCReceivedMessages/interval)}

	// Set the details metrics for HTTP traffic. We always display the server
	// traffic first and afterwards the client traffic. All metrics are set also
	// when they are zero.
	field.DetailsHTTPRate = []string{fmt.Sprintf("%.2frps", (node.ServerHTTPRequestsSuccess+node.ServerHTTPRequestsError)/interval), fmt.Sprintf("%.2frps", (node.ClientHTTPRequestsSuccess+node.ClientHTTPRequestsError)/interval)}
	if node.ServerHTTPRequestsError > 0 && node.ClientHTTPRequestsError > 0 {
		httpServerErrRate = (node.ServerHTTPRequestsError / (node.ServerHTTPRequestsSuccess + node.ServerHTTPRequestsError)) * 100
		httpClientErrRate = (node.ClientHTTPRequestsError / (node.ClientHTTPRequestsSuccess + node.ClientHTTPRequestsError)) * 100
		field.DetailsHTTPErr = []string{fmt.Sprintf("%.2f%%", httpServerErrRate), fmt.Sprintf("%.2f%%", httpClientErrRate)}
	} else if node.ServerHTTPRequestsError > 0 && node.ClientHTTPRequestsError == 0 {
		httpServerErrRate = (node.ServerHTTPRequestsError / (node.ServerHTTPRequestsSuccess + node.ServerHTTPRequestsError)) * 100
		httpClientErrRate = 0
		field.DetailsHTTPErr = []string{fmt.Sprintf("%.2f%%", httpServerErrRate), "0.00%"}
	} else if node.ServerHTTPRequestsError == 0 && node.ClientHTTPRequestsError > 0 {
		httpServerErrRate = 0
		httpClientErrRate = (node.ClientHTTPRequestsError / (node.ClientHTTPRequestsSuccess + node.ClientHTTPRequestsError)) * 100
		field.DetailsHTTPErr = []string{"0.00%", fmt.Sprintf("%.2f%%", httpClientErrRate)}
	} else {
		httpServerErrRate = 0
		httpClientErrRate = 0
		field.DetailsHTTPErr = []string{"0.00%", "0.00%"}
	}

	// Set the details metrics for TCP traffic.
	field.DetailsTCPSentBytes = []string{fmt.Sprintf("%.2fbps", node.ServerTCPSentBytes/interval), fmt.Sprintf("%.2fbps", node.ClientTCPSentBytes/interval)}
	field.DetailsTCPReceivedBytes = []string{fmt.Sprintf("%.2fbps", node.ServerTCPReceivedBytes/interval), fmt.Sprintf("%.2fbps", node.ClientTCPReceivedBytes/interval)}

	// Set the color, main stat and secondary stat based on the traffic type:
	// - We always prefer server traffic over the client traffic.
	// - We prefer the traffic type with more requests. This means if we have
	//   more HTTP traffic then gRPC traffic we use the HTTP metrics in the
	//   same way as we do it for edges, otherwise we use the gRPC metrics in a
	//   similar way.
	if node.ServerHTTPRequestsSuccess+node.ServerHTTPRequestsError > node.ServerGRPCRequestsSuccess+node.ServerGRPCRequestsError {
		field.MainStat = append(field.MainStat, field.DetailsHTTPRate[0])
		if httpServerErrRate > 0 {
			field.MainStat = append(field.MainStat, field.DetailsHTTPErr[0])
		}

		if httpServerErrRate >= d.istioErrorThreshold {
			field.Color = "#f2495c"
		} else if httpServerErrRate > d.istioWarningThreshold {
			field.Color = "#fade2a"
		} else {
			field.Color = "#73bf69"
		}

		if node.ServerTCPSentBytes+node.ServerTCPReceivedBytes > 0 {
			field.SecondaryStat = append(field.SecondaryStat, fmt.Sprintf("%.2fbps", (node.ServerTCPSentBytes+node.ServerTCPReceivedBytes)/interval))
		}
	} else if node.ServerGRPCRequestsSuccess+node.ServerGRPCRequestsError > 0 {
		field.MainStat = append(field.MainStat, field.DetailsGRPCRate[0])
		if grpcServerErrRate > 0 {
			field.MainStat = append(field.MainStat, field.DetailsGRPCErr[0])
		}

		if grpcServerErrRate >= d.istioErrorThreshold {
			field.Color = "#f2495c"
		} else if grpcServerErrRate > d.istioWarningThreshold {
			field.Color = "#fade2a"
		} else {
			field.Color = "#73bf69"
		}

		if node.ServerTCPSentBytes+node.ServerTCPReceivedBytes > 0 {
			field.SecondaryStat = append(field.SecondaryStat, fmt.Sprintf("%.2fbps", (node.ServerTCPSentBytes+node.ServerTCPReceivedBytes)/interval))
		}
	} else if node.ClientHTTPRequestsSuccess+node.ClientHTTPRequestsError > node.ClientGRPCRequestsSuccess+node.ClientGRPCRequestsError {
		field.MainStat = append(field.MainStat, field.DetailsHTTPRate[1])
		if httpClientErrRate > 0 {
			field.MainStat = append(field.MainStat, field.DetailsHTTPErr[1])
		}

		if httpClientErrRate >= d.istioErrorThreshold {
			field.Color = "#f2495c"
		} else if httpClientErrRate > d.istioWarningThreshold {
			field.Color = "#fade2a"
		} else {
			field.Color = "#73bf69"
		}

		if node.ClientTCPSentBytes+node.ClientTCPReceivedBytes > 0 {
			field.SecondaryStat = append(field.SecondaryStat, fmt.Sprintf("%.2fbps", (node.ClientTCPSentBytes+node.ClientTCPReceivedBytes)/interval))
		}
	} else if node.ClientGRPCRequestsSuccess+node.ClientGRPCRequestsError > 0 {
		field.MainStat = append(field.MainStat, field.DetailsGRPCRate[1])
		if grpcClientErrRate > 0 {
			field.MainStat = append(field.MainStat, field.DetailsGRPCErr[1])
		}

		if grpcClientErrRate >= d.istioErrorThreshold {
			field.Color = "#f2495c"
		} else if grpcClientErrRate > d.istioWarningThreshold {
			field.Color = "#fade2a"
		} else {
			field.Color = "#73bf69"
		}

		if node.ClientTCPSentBytes+node.ClientTCPReceivedBytes > 0 {
			field.SecondaryStat = append(field.SecondaryStat, fmt.Sprintf("%.2fbps", (node.ClientTCPSentBytes+node.ClientTCPReceivedBytes)/interval))
		}
	} else if node.ServerTCPSentBytes+node.ServerTCPReceivedBytes > 0 {
		field.MainStat = append(field.MainStat, fmt.Sprintf("%.2fbps", (node.ServerTCPSentBytes+node.ServerTCPReceivedBytes)/interval))
		field.Color = "#5794f2"
	} else if node.ClientTCPSentBytes+node.ClientTCPReceivedBytes > 0 {
		field.MainStat = append(field.MainStat, fmt.Sprintf("%.2fbps", (node.ClientTCPSentBytes+node.ClientTCPReceivedBytes)/interval))
		field.Color = "#5794f2"
	} else {
		field.Color = "#ccccdc"
	}

	return field
}
