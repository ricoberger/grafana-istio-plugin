package prometheus

import (
	"context"

	"github.com/ricoberger/grafana-istio-plugin/pkg/models"
	"github.com/ricoberger/grafana-istio-plugin/pkg/roundtripper"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Client interface {
	CheckHealth(ctx context.Context) error
	GetLabelValues(ctx context.Context, query LabelValuesQuery, timeRange backend.TimeRange) ([]string, error)
	GetMetrics(ctx context.Context, metric, query string, timeRange backend.TimeRange) ([]Metric, error)
}

type client struct {
	api v1.API
}

func (c *client) CheckHealth(ctx context.Context) error {
	_, err := c.api.Buildinfo(ctx)
	return err
}

func (c *client) GetLabelValues(ctx context.Context, query LabelValuesQuery, timeRange backend.TimeRange) ([]string, error) {
	labelValues, _, err := c.api.LabelValues(ctx, query.Label, query.Matches, timeRange.From, timeRange.To)
	if err != nil {
		return nil, err
	}

	var values []string

	for _, labelValue := range labelValues {
		values = append(values, string(labelValue))
	}

	return values, nil
}

func (c *client) GetMetrics(ctx context.Context, metric, query string, timeRange backend.TimeRange) ([]Metric, error) {
	result, _, err := c.api.Query(ctx, query, timeRange.To)
	if err != nil {
		return nil, err
	}

	streams, ok := result.(model.Vector)
	if !ok {
		return nil, err
	}

	var metrics []Metric

	for _, stream := range streams {
		labels := make(map[string]string)
		labels["metric"] = metric

		for key, value := range stream.Metric {
			labels[string(key)] = string(value)
		}

		metrics = append(metrics, Metric{
			Value:  float64(stream.Value),
			Labels: labels,
		})
	}

	return metrics, nil
}

func NewClient(settings *models.PluginSettings) (Client, error) {
	roundTripper := roundtripper.DefaultRoundTripper

	if settings.PrometheusAuthMethod == models.PrometheusAuthMethodBasic {
		roundTripper = roundtripper.BasicAuthTransport{
			Transport: roundTripper,
			Username:  settings.PrometheusUsername,
			Password:  settings.Secrets.PrometheusPassword,
		}
	}

	if settings.PrometheusAuthMethod == models.PrometheusAuthMethodToken {
		roundTripper = roundtripper.TokenAuthTransporter{
			Transport: roundTripper,
			Token:     settings.Secrets.PrometheusToken,
		}
	}

	apiClient, err := api.NewClient(api.Config{
		Address:      settings.PrometheusUrl,
		RoundTripper: roundTripper,
	})
	if err != nil {
		return nil, err
	}

	return &client{
		api: v1.NewAPI(apiClient),
	}, nil
}
