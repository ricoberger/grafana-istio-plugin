package models

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

const (
	PrometheusAuthMethodNone  = "none"
	PrometheusAuthMethodBasic = "basic"
	PrometheusAuthMethodToken = "token"
)

type PluginSettings struct {
	PrometheusUrl          string                `json:"prometheusUrl"`
	PrometheusAuthMethod   string                `json:"prometheusAuthMethod"`
	PrometheusUsername     string                `json:"prometheusUsername"`
	IstioWarningThreshold  float64               `json:"istioWarningThreshold"`
	IstioErrorThreshold    float64               `json:"istioErrorThreshold"`
	IstioWorkloadDashboard string                `json:"istioWorkloadDashboard"`
	IstioServiceDashboard  string                `json:"istioServiceDashboard"`
	Secrets                *SecretPluginSettings `json:"-"`
}

type SecretPluginSettings struct {
	PrometheusPassword string `json:"prometheusPassword"`
	PrometheusToken    string `json:"prometheusToken"`
}

func LoadPluginSettings(source backend.DataSourceInstanceSettings) (*PluginSettings, error) {
	settings := PluginSettings{}
	err := json.Unmarshal(source.JSONData, &settings)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal PluginSettings json: %w", err)
	}

	settings.Secrets = loadSecretPluginSettings(source.DecryptedSecureJSONData)

	return &settings, nil
}

func loadSecretPluginSettings(source map[string]string) *SecretPluginSettings {
	return &SecretPluginSettings{
		PrometheusPassword: source["prometheusPassword"],
		PrometheusToken:    source["prometheusToken"],
	}
}
