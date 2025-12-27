package models

import (
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type Fields []*data.Field

func (f *Fields) Add(name string, labels data.Labels, values any, config ...*data.FieldConfig) *data.Field {
	field := data.NewField(name, labels, values)
	if len(config) > 0 {
		field.SetConfig(config[0])
	}
	*f = append(*f, field)
	return field
}

type Edge struct {
	ID                   string
	Source               string
	SourceType           string
	SourceName           string
	SourceNamespace      string
	Destination          string
	DestinationType      string
	DestinationName      string
	DestinationNamespace string
	DestinationService   string
	GRPCResponseCodes    map[string]float64
	GRPCRequestsSuccess  float64
	GRPCRequestsError    float64
	GRPCRequestDuration  float64
	GRPCSentMessages     float64
	GRPCReceivedMessages float64
	HTTPResponseCodes    map[string]float64
	HTTPRequestsSuccess  float64
	HTTPRequestsError    float64
	HTTPRequestDuration  float64
	TCPSentBytes         float64
	TCPReceivedBytes     float64
}

type Node struct {
	ID                         string
	Type                       string
	Name                       string
	Namespace                  string
	Service                    string
	ClientGRPCResponseCodes    map[string]float64
	ClientGRPCRequestsSuccess  float64
	ClientGRPCRequestsError    float64
	ClientGRPCSentMessages     float64
	ClientGRPCReceivedMessages float64
	ClientHTTPResponseCodes    map[string]float64
	ClientHTTPRequestsSuccess  float64
	ClientHTTPRequestsError    float64
	ClientTCPSentBytes         float64
	ClientTCPReceivedBytes     float64
	ServerGRPCResponseCodes    map[string]float64
	ServerGRPCRequestsSuccess  float64
	ServerGRPCRequestsError    float64
	ServerGRPCSentMessages     float64
	ServerGRPCReceivedMessages float64
	ServerHTTPResponseCodes    map[string]float64
	ServerHTTPRequestsSuccess  float64
	ServerHTTPRequestsError    float64
	ServerTCPSentBytes         float64
	ServerTCPReceivedBytes     float64
}

type Field struct {
	ID                          string
	Source                      string
	Destination                 string
	MainStat                    []string
	SecondaryStat               []string
	Color                       string
	DetailsGRPCRate             []string
	DetailsGRPCErr              []string
	DetailsGRPCDuration         []string
	DetailsGRPCSentMessages     []string
	DetailsGRPCReceivedMessages []string
	DetailsHTTPRate             []string
	DetailsHTTPErr              []string
	DetailsHTTPDuration         []string
	DetailsTCPSentBytes         []string
	DetailsTCPReceivedBytes     []string
}
