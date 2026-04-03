package config

// MQConfig RabbitMQ配置
type MQConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	VHost    string
}

// GetMQConfig 获取MQ配置
func GetMQConfig() MQConfig {
	return MQConfig{
		Host:     "rabbitmq", // Docker Compose服务名
		Port:     "5672",
		Username: "guest",
		Password: "guest",
		VHost:    "/",
	}
}