package plugin

import (
	"context"

	"github.com/ricoberger/grafana-istio-plugin/pkg/models"
	"github.com/ricoberger/grafana-istio-plugin/pkg/prometheus"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin
// in runtime. In this example datasource instance implements
// backend.QueryDataHandler, backend.CheckHealthHandler interfaces. Plugin
// should not implement all these interfaces - only those which are required for
// a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(_ context.Context, pCtx backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger := backend.Logger.With("datasource", pCtx.Name).With("datasourceId", pCtx.ID).With("datasourceUid", pCtx.UID)
	logger.Debug("Creating new datasource instance")

	settings, err := models.LoadPluginSettings(pCtx)
	if err != nil {
		logger.Error("Failed to load plugin settings", "error", err.Error())
		return nil, err
	}

	prometheusClient, err := prometheus.NewClient(settings)
	if err != nil {
		logger.Error("Failed to create Prometheus client", "error", err.Error())
		return nil, err
	}

	istioWarningThreshold := settings.IstioWarningThreshold
	if istioWarningThreshold == 0 {
		istioWarningThreshold = 0
	}

	istioErrorThreshold := settings.IstioErrorThreshold
	if istioErrorThreshold == 0 {
		istioErrorThreshold = 5
	}

	ds := &Datasource{
		prometheusClient:       prometheusClient,
		istioWarningThreshold:  istioWarningThreshold,
		istioErrorThreshold:    istioErrorThreshold,
		istioWorkloadDashboard: settings.IstioWorkloadDashboard,
		istioServiceDashboard:  settings.IstioServiceDashboard,
		logger:                 logger,
	}

	queryTypeMux := datasource.NewQueryTypeMux()
	queryTypeMux.HandleFunc(models.QueryTypeNamespaces, ds.handleNamespacesQueries)
	queryTypeMux.HandleFunc(models.QueryTypeApplications, ds.handleApplicationsQueries)
	queryTypeMux.HandleFunc(models.QueryTypeWorkloads, ds.handleWorkloadsQueries)
	queryTypeMux.HandleFunc(models.QueryTypeFilters, ds.handleFiltersQueries)
	queryTypeMux.HandleFunc(models.QueryTypeApplicationGraph, ds.handleApplicationGraphQueries)
	queryTypeMux.HandleFunc(models.QueryTypeWorkloadGraph, ds.handleWorkloadGraphQueries)
	queryTypeMux.HandleFunc(models.QueryTypeNamespaceGraph, ds.handleNamespaceGraphQueries)
	ds.queryHandler = queryTypeMux

	return ds, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	queryHandler           backend.QueryDataHandler
	prometheusClient       prometheus.Client
	istioWarningThreshold  float64
	istioErrorThreshold    float64
	istioWorkloadDashboard string
	istioServiceDashboard  string
	logger                 log.Logger
}

// QueryData handles multiple queries and returns multiple responses. The
// queries are matched by their QueryType field against a handler function. See
// the NewDatasource function where the query type multiplexer is created and
// handlers are registered.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ctx, span := tracing.DefaultTracer().Start(ctx, "QueryData")
	defer span.End()

	return d.queryHandler.QueryData(ctx, req)
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a
// new instance created. As soon as datasource settings change detected by SDK
// old datasource instance will be disposed and a new one will be created using
// NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// CheckHealth handles health checks sent from Grafana to the plugin. The main
// use case for these health checks is the test button on the datasource
// configuration page which allows users to verify that a datasource is working
// as expected.
func (d *Datasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}

	err := d.prometheusClient.CheckHealth(ctx)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Prometheus health check failed: " + err.Error()
		return res, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}
