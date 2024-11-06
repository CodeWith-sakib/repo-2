package config

import (
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
}

type DatabaseConfig struct {
	Type         string `json:"type"` // memory or postgres
	DSN          string `json:"dsn"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
}

type WorkerConfig struct {
	WorkerID          string        `json:"worker_id"`
	Concurrency       int           `json:"concurrency"`
	PollInterval      time.Duration `json:"poll_interval"`
	LeaseDuration     time.Duration `json:"lease_duration"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
}

type Config struct {
	ConfigFile string         `json:"-"`
	Server     ServerConfig   `json:"server"`
	Database   DatabaseConfig `json:"database"`
	Worker     WorkerConfig   `json:"worker"`
}

func DefaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Host:         "127.0.0.1",
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
		Database: DatabaseConfig{
			Type:         "memory",
			DSN:          "",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
		},
		Worker: WorkerConfig{
			WorkerID:          "worker-default",
			Concurrency:       4,
			PollInterval:      50 * time.Millisecond,
			LeaseDuration:     10 * time.Second,
			HeartbeatInterval: 3 * time.Second,
		},
	}
}

type FlagOverrides struct {
	ConfigFile  *string
	Port        *int
	Host        *string
	DBType      *string
	DBDSN       *string
	Concurrency *int
}

func Load(args []string) (*Config, error) {
	// 1. Defaults
	cfg := DefaultConfig()

	// 2. Parse Flags
	fs := flag.NewFlagSet("kestrel", flag.ContinueOnError)
	configFile := fs.String("config", "", "Path to configuration file")
	port := fs.Int("port", 0, "Server HTTP port")
	host := fs.String("host", "", "Server HTTP host")
	dbType := fs.String("db-type", "", "Database backend (memory|postgres)")
	dbDSN := fs.String("db-dsn", "", "Database connection DSN")
	concurrency := fs.Int("concurrency", 0, "Worker concurrency limit")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Determine config file path (Flag > Env)
	filePath := *configFile
	if filePath == "" {
		filePath = os.Getenv("KESTREL_CONFIG")
	}

	// 3. Load from Config File (if specified)
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err == nil {
			var fileCfg Config
			if err := json.Unmarshal(data, &fileCfg); err == nil {
				if fileCfg.Server.Host != "" {
					cfg.Server.Host = fileCfg.Server.Host
				}
				if fileCfg.Server.Port != 0 {
					cfg.Server.Port = fileCfg.Server.Port
				}
				if fileCfg.Database.Type != "" {
					cfg.Database.Type = fileCfg.Database.Type
				}
				if fileCfg.Database.DSN != "" {
					cfg.Database.DSN = fileCfg.Database.DSN
				}
				if fileCfg.Worker.Concurrency != 0 {
					cfg.Worker.Concurrency = fileCfg.Worker.Concurrency
				}
				if fileCfg.Worker.WorkerID != "" {
					cfg.Worker.WorkerID = fileCfg.Worker.WorkerID
				}
			}
		}
	}

	// 4. Override with Environment Variables
	if envHost := os.Getenv("KESTREL_SERVER_HOST"); envHost != "" {
		cfg.Server.Host = envHost
	}
	if envPort := os.Getenv("KESTREL_SERVER_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			cfg.Server.Port = p
		}
	}
	if envDBType := os.Getenv("KESTREL_DB_TYPE"); envDBType != "" {
		cfg.Database.Type = envDBType
	}
	if envDBDSN := os.Getenv("KESTREL_DB_DSN"); envDBDSN != "" {
		cfg.Database.DSN = envDBDSN
	}
	if envConc := os.Getenv("KESTREL_WORKER_CONCURRENCY"); envConc != "" {
		if c, err := strconv.Atoi(envConc); err == nil && c > 0 {
			cfg.Worker.Concurrency = c
		}
	}

	// 5. Override with CLI Flags
	if *host != "" {
		cfg.Server.Host = *host
	}
	if *port != 0 {
		cfg.Server.Port = *port
	}
	if *dbType != "" {
		cfg.Database.Type = *dbType
	}
	if *dbDSN != "" {
		cfg.Database.DSN = *dbDSN
	}
	if *concurrency != 0 {
		cfg.Worker.Concurrency = *concurrency
	}

	cfg.ConfigFile = filePath
	return &cfg, nil
}
