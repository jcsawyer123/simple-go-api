package config

import "os"

type Config struct {
	Port           string
	AuthServiceURL string
	SNSTopicARN    string
	SQSQueueURL    string
	AWSRegion      string
	ProfilingPort  string

	// Metrics configuration
	Metrics MetricsConfig
}

type MetricsConfig struct {
	// Enable or disable metrics collection
	Enabled bool

	// Prometheus configuration
	Prometheus PrometheusConfig

	// Datadog configuration
	Datadog DatadogConfig
}

type PrometheusConfig struct {
	// Enable Prometheus metrics
	Enabled bool

	// Namespace (prefix) for metrics
	Namespace string

	// Subsystem (secondary prefix) for metrics
	Subsystem string

	// HTTP address to expose Prometheus metrics
	HTTPAddr string
}

type DatadogConfig struct {
	// Enable Datadog metrics
	Enabled bool

	// Namespace (prefix) for metrics
	Namespace string

	// Address of the DogStatsD server
	Address string

	// Default tags to add to all metrics
	DefaultTags map[string]string
}

func Load() (*Config, error) {
	// Load environment variables for metrics
	datadogEnabled := getEnvOrDefault("METRICS_DATADOG_ENABLED", "false") == "true"
	prometheusEnabled := getEnvOrDefault("METRICS_PROMETHEUS_ENABLED", "false") == "true"

	// Default tags for Datadog
	defaultTags := map[string]string{
		"service": "simple-go-api",
		"env":     getEnvOrDefault("ENV", "development"),
	}

	return &Config{
		Port:           getEnvOrDefault("PORT", "8080"),
		AuthServiceURL: getEnvOrDefault("AUTH_SERVICE_URL", "https://api.product.dev.alertlogic.com"),
		SNSTopicARN:    getEnvOrDefault("SNS_TOPIC_ARN", ""),
		SQSQueueURL:    getEnvOrDefault("SQS_QUEUE_URL", ""),
		AWSRegion:      getEnvOrDefault("AWS_REGION", "us-west-2"),
		ProfilingPort:  getEnvOrDefault("PROFILING_PORT", "6060"),

		Metrics: MetricsConfig{
			Enabled: datadogEnabled || prometheusEnabled,

			Prometheus: PrometheusConfig{
				Enabled:   prometheusEnabled,
				Namespace: "simple_go_api",
				Subsystem: "server",
				HTTPAddr:  getEnvOrDefault("METRICS_PROMETHEUS_ADDR", ":9090"),
			},

			Datadog: DatadogConfig{
				Enabled:     datadogEnabled,
				Namespace:   "simple_go_api",
				Address:     getEnvOrDefault("METRICS_DATADOG_ADDR", "localhost:8125"),
				DefaultTags: defaultTags,
			},
		},
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
