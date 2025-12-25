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

type EdgeField struct {
	ID                          string
	Source                      string
	Destination                 string
	MainStat                    []string
	SecondaryStat               []string
	Color                       string
	DetailsGRPCRate             string
	DetailsGRPCErr              string
	DetailsGRPCDuration         string
	DetailsGRPCSentMessages     string
	DetailsGRPCReceivedMessages string
	DetailsHTTPRate             string
	DetailsHTTPErr              string
	DetailsHTTPDuration         string
	DetailsTCPSentBytes         string
	DetailsTCPReceivedBytes     string
}
