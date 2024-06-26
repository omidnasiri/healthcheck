package config

type Config struct {
	DB         DBConfig
	WebhookURL string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}
