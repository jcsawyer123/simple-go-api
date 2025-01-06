package config

import "os"

type Config struct {
    Port           string
    AuthServiceURL string
    SNSTopicARN    string
    SQSQueueURL    string
    AWSRegion      string
}

func Load() (*Config, error) {
    return &Config{
        Port:           getEnvOrDefault("PORT", "8080"),
        AuthServiceURL: getEnvOrDefault("AUTH_SERVICE_URL", "http://auth-service"),
        SNSTopicARN:    getEnvOrDefault("SNS_TOPIC_ARN", ""),
        SQSQueueURL:    getEnvOrDefault("SQS_QUEUE_URL", ""),
        AWSRegion:      getEnvOrDefault("AWS_REGION", "us-west-2"),
    }, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value, exists := os.LookupEnv(key); exists {
        return value
    }
    return defaultValue
}
