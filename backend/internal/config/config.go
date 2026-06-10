package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	MinIO    MinIOConfig    `mapstructure:"minio"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Mode         string        `mapstructure:"mode"` // debug, release
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"timezone"`
}

func (d *DatabaseConfig) DSN() string {
	return "host=" + d.Host + " port=" + strconv.Itoa(d.Port) + " user=" + d.User +
		" password=" + d.Password + " dbname=" + d.DBName + " sslmode=" + d.SSLMode +
		" TimeZone=" + d.TimeZone
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MinIOConfig struct {
	Endpoint       string `mapstructure:"endpoint"`
	PublicEndpoint string `mapstructure:"public_endpoint"`
	AccessKey      string `mapstructure:"access_key"`
	SecretKey      string `mapstructure:"secret_key"`
	SSL            bool   `mapstructure:"ssl"`
}

type JWTConfig struct {
	Secret      string        `mapstructure:"secret"`
	AccessTTL   time.Duration `mapstructure:"access_ttl"`
	RefreshTTL  time.Duration `mapstructure:"refresh_ttl"`
	Issuer      string        `mapstructure:"issuer"`
}

func Load(path string) *Config {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setEnvIfSet("DB_HOST", "database.host")
	setEnvIfSet("DB_PORT", "database.port")
	setEnvIfSet("DB_USER", "database.user")
	setEnvIfSet("DB_PASSWORD", "database.password")
	setEnvIfSet("DB_NAME", "database.dbname")
	setEnvIfSet("DB_SSLMODE", "database.sslmode")
	setEnvIfSet("DB_TIMEZONE", "database.timezone")
	setEnvIfSet("REDIS_ADDR", "redis.addr")
	setEnvIfSet("REDIS_PASSWORD", "redis.password")
	setEnvIfSet("REDIS_DB", "redis.db")
	setEnvIfSet("MINIO_ENDPOINT", "minio.endpoint")
	setEnvIfSet("MINIO_PUBLIC_ENDPOINT", "minio.public_endpoint")
	setEnvIfSet("MINIO_ACCESS_KEY", "minio.access_key")
	setEnvIfSet("MINIO_SECRET_KEY", "minio.secret_key")
	setEnvIfSet("MINIO_SSL", "minio.ssl")
	setEnvIfSet("JWT_SECRET", "jwt.secret")
	setEnvIfSet("SERVER_PORT", "server.port")
	setEnvIfSet("SERVER_MODE", "server.mode")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}
	return &cfg
}

func setEnvIfSet(envKey, viperKey string) {
	if v, ok := os.LookupEnv(envKey); ok {
		viper.Set(viperKey, v)
	}
}
