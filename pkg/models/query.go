package models

type QueryType string

const (
	QueryTypeNamespaces   = "namespaces"
	QueryTypeApplications = "applications"
	QueryTypeGraph        = "graph"

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

type QueryModelGraph struct {
	Namespace   string   `json:"namespace"`
	Application string   `json:"application"`
	Metrics     []string `json:"metrics"`
}
