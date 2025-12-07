package env

// EnvironmentConfig represents the infrastructure needed for a test suite.
type EnvironmentConfig struct {
	Postgres *PostgresConfig `json:"postgres"`
	Mongo    *MongoConfig    `json:"mongodb"`
	Redis    *RedisConfig    `json:"redis"`
	RabbitMQ *RabbitConfig   `json:"rabbitmq"`
}

type PostgresConfig struct {
	Version   string            `json:"version"`
	Memory    string            `json:"memory"`
	Databases []PostgresDBEntry `json:"databases"`
}

type PostgresDBEntry struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type MongoConfig struct {
	Version   string         `json:"version"`
	Memory    string         `json:"memory"`
	Databases []MongoDBEntry `json:"databases"`
}

type MongoDBEntry struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RedisConfig struct {
	Version string `json:"version"`
	Memory  string `json:"memory"`
}

type RabbitConfig struct {
	Version string `json:"version"`
	Memory  string `json:"memory"`
}
