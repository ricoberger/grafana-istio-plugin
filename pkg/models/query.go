package models

type QueryType string

const (
	QueryTypeNamespaces       = "namespaces"
	QueryTypeApplications     = "applications"
	QueryTypeWorkloads        = "workloads"
	QueryTypeFilters          = "filters"
	QueryTypeApplicationGraph = "applicationgraph"
	QueryTypeWorkloadGraph    = "workloadgraph"
	QueryTypeNamespaceGraph   = "namespacegraph"

	MetricGRPCRequests         = "grpcRequests"
	MetricGRPCRequestDuration  = "grpcRequestDuration"
	MetricGRPCSentMessages     = "grpcSentMessages"
	MetricGRPCReceivedMessages = "grpcReceivedMessages"
	MetricHTTPRequests         = "httpRequests"
	MetricHTTPRequestDuration  = "httpRequestDuration"
	MetricTCPSentBytes         = "tcpSentBytes"
	MetricTCPReceivedBytes     = "tcpReceivedBytes"
)

type QueryModelApplications struct {
	Namespace string `json:"namespace"`
}

type QueryModelWorkloads struct {
	Namespace string `json:"namespace"`
}

type QueryModelFilters struct {
	FilterType  string `json:"filterType"`
	Namespace   string `json:"namespace"`
	Application string `json:"application"`
	Workload    string `json:"workload"`
}

type QueryModelApplicationGraph struct {
	Namespace          string   `json:"namespace"`
	Application        string   `json:"application"`
	Metrics            []string `json:"metrics"`
	IdleEdges          bool     `json:"idleEdges"`
	SourceFilters      []string `json:"sourceFilters"`
	DestinationFilters []string `json:"destinationFilters"`
}

type QueryModelWorkloadGraph struct {
	Namespace          string   `json:"namespace"`
	Workload           string   `json:"workload"`
	Metrics            []string `json:"metrics"`
	IdleEdges          bool     `json:"idleEdges"`
	SourceFilters      []string `json:"sourceFilters"`
	DestinationFilters []string `json:"destinationFilters"`
}

type QueryModelNamespaceGraph struct {
	Namespace          string   `json:"namespace"`
	Metrics            []string `json:"metrics"`
	IdleEdges          bool     `json:"idleEdges"`
	SourceFilters      []string `json:"sourceFilters"`
	DestinationFilters []string `json:"destinationFilters"`
}
