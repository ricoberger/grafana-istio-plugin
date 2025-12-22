package prometheus

type LabelValuesQuery struct {
	Label   string
	Matches []string
}

type Metric struct {
	Value  float64
	Labels map[string]string
}
