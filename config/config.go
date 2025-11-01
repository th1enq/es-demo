package config

import (
	"time"

	"github.com/spf13/viper"
	"github.com/th1enq/es-demo/internal/delivery/http"
	"github.com/th1enq/es-demo/pkg/es"
	kafkaClient "github.com/th1enq/es-demo/pkg/kafka"
	"github.com/th1enq/es-demo/pkg/logger"
	"github.com/th1enq/es-demo/pkg/mongodb"
	"github.com/th1enq/es-demo/pkg/postgres"
)

type JWTConfig struct {
	SecretKey     string        `json:"secret_key"`
	TokenDuration time.Duration `json:"token_duration"`
	Issuer        string        `json:"issuer"`
}

type Config struct {
	Logger               logger.Config
	Postgres             postgres.Config
	PgStore              es.Config
	MongoDB              mongodb.Config
	Server               http.Config
	JWT                  JWTConfig
	KafkaPublisherConfig es.KafkaEventsBusConfig
	Kafka                *kafkaClient.Config
	Projections          Projections
}

type Projections struct {
	MongoGroup                string
	MongoSubscriptionPoolSize int
}

func Load() *Config {
	viper := viper.New()

	// Enable reading from environment variables
	viper.AutomaticEnv()

	viper.SetDefault("LOGGER_LEVEL", "info")
	viper.SetDefault("LOGGER_FILE_PATH", "./logs/app.log")
	viper.SetDefault("LOGGER_MAX_SIZE", 100)
	viper.SetDefault("LOGGER_MAX_BACKUPS", 10)
	viper.SetDefault("LOGGER_MAX_AGE", 30)
	loggerEnv := logger.Config{
		Level:      viper.GetString("LOGGER_LEVEL"),
		FilePath:   viper.GetString("LOGGER_FILE_PATH"),
		MaxSize:    viper.GetInt("LOGGER_MAX_SIZE"),
		MaxBackups: viper.GetInt("LOGGER_MAX_BACKUPS"),
		MaxAge:     viper.GetInt("LOGGER_MAX_AGE"),
	}

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "password")
	viper.SetDefault("DB_NAME", "es_demo")
	postgresEnv := postgres.Config{
		Host:     viper.GetString("DB_HOST"),
		Port:     viper.GetInt("DB_PORT"),
		User:     viper.GetString("DB_USER"),
		Password: viper.GetString("DB_PASSWORD"),
		DBName:   viper.GetString("DB_NAME"),
	}

	viper.SetDefault("SNAPSHOT_FREQUENCY", 5)
	pgStoreEnv := es.Config{
		SnapshotFrequency: viper.GetUint64("SNAPSHOT_FREQUENCY"),
	}

	viper.SetDefault("MONGODB_URI", "mongodb://localhost:27017")
	viper.SetDefault("MONGODB_DATABASE", "es_demo")
	viper.SetDefault("MONGODB_USERNAME", "appuser")
	viper.SetDefault("MONGODB_PASSWORD", "apppassword")
	mongoDBEnv := mongodb.Config{
		URI:      viper.GetString("MONGODB_URI"),
		Db:       viper.GetString("MONGODB_DATABASE"),
		User:     viper.GetString("MONGODB_USERNAME"),
		Password: viper.GetString("MONGODB_PASSWORD"),
	}

	viper.SetDefault("SERVER_HOST", "localhost")
	viper.SetDefault("SERVER_PORT", 8080)
	serverEnv := http.Config{
		Host: viper.GetString("SERVER_HOST"),
		Port: viper.GetInt("SERVER_PORT"),
	}

	viper.SetDefault("JWT_SECRET_KEY", "your-super-secret-jwt-key-change-in-production")
	viper.SetDefault("JWT_TOKEN_DURATION", "24h")
	viper.SetDefault("JWT_ISSUER", "es-demo-banking")
	tokenDuration, _ := time.ParseDuration(viper.GetString("JWT_TOKEN_DURATION"))
	jwtEnv := JWTConfig{
		SecretKey:     viper.GetString("JWT_SECRET_KEY"),
		TokenDuration: tokenDuration,
		Issuer:        viper.GetString("JWT_ISSUER"),
	}

	// Kafka Configuration
	viper.SetDefault("KAFKA_BROKERS", "localhost:9093")
	viper.SetDefault("KAFKA_GROUP_ID", "bank_account_microservice_consumer")
	viper.SetDefault("KAFKA_INIT_TOPICS", true)
	kafkaEnv := &kafkaClient.Config{
		Brokers: []string{viper.GetString("KAFKA_BROKERS")},
		GroupID: viper.GetString("KAFKA_GROUP_ID"),
	}

	// Kafka Publisher (Event Bus) Configuration
	viper.SetDefault("KAFKA_TOPIC_PREFIX", "bank_account")
	viper.SetDefault("KAFKA_PARTITIONS", 10)
	viper.SetDefault("KAFKA_REPLICATION_FACTOR", 1)
	kafkaPublisherEnv := es.KafkaEventsBusConfig{
		TopicPrefix:       viper.GetString("KAFKA_TOPIC_PREFIX"),
		Partitions:        viper.GetInt("KAFKA_PARTITIONS"),
		ReplicationFactor: viper.GetInt("KAFKA_REPLICATION_FACTOR"),
	}

	// Projections Configuration
	viper.SetDefault("PROJECTION_MONGO_GROUP", "mongoGroup")
	viper.SetDefault("PROJECTION_MONGO_POOL_SIZE", 10)
	projectionsEnv := Projections{
		MongoGroup:                viper.GetString("PROJECTION_MONGO_GROUP"),
		MongoSubscriptionPoolSize: viper.GetInt("PROJECTION_MONGO_POOL_SIZE"),
	}

	return &Config{
		Logger:               loggerEnv,
		Postgres:             postgresEnv,
		PgStore:              pgStoreEnv,
		MongoDB:              mongoDBEnv,
		Server:               serverEnv,
		JWT:                  jwtEnv,
		Kafka:                kafkaEnv,
		KafkaPublisherConfig: kafkaPublisherEnv,
		Projections:          projectionsEnv,
	}
}
