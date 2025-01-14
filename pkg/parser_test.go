package simpleini

import (
	"strings"
	"testing"
)

type DatabaseConfig struct {
	Host     string `ini:"host"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}

type LoggingConfig struct {
	Level string `ini:"level"`
	File  string `ini:"file"`
}

type ServerConfig struct {
	Host     string        `ini:"host"`
	Port     int           `ini:"port"`
	Username string        `ini:"username"`
	Password string        `ini:"password"`
	Timeout  float64       `ini:"timeout"`
	Enabled  bool          `ini:"enabled"`
	Logging  LoggingConfig `ini:"logging"`
}

type Config struct {
	AppName  string         `ini:"app_name"`
	Version  string         `ini:"version"`
	Server   ServerConfig   `ini:"server"`
	Database DatabaseConfig `ini:"database"`
}

func TestParseINI(t *testing.T) {
	iniContent := `
		; This is a comment
		# This is another comment

		app_name = MyApp
		version = 1.0.0

		[server]
		host = localhost
		port = 8080
		username = admin
		password = secret
		timeout = 30.5
		enabled = true

		[server.logging]
		level = debug
		file = /var/log/myapp.log

		[database]
		host = db.local
		port = 5432
		username = dbadmin
		password = dbsecret
	`

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err != nil {
		t.Fatalf("Failed to parse INI: %v", err)
	}

	if config.AppName != "MyApp" {
		t.Errorf("Expected app_name to be 'MyApp', got '%s'", config.AppName)
	}
	if config.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got '%s'", config.Version)
	}
	if config.Server.Host != "localhost" {
		t.Errorf("Expected server host to be 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected server port to be 8080, got %d", config.Server.Port)
	}
	if config.Server.Username != "admin" {
		t.Errorf("Expected server username to be 'admin', got '%s'", config.Server.Username)
	}
	if config.Server.Password != "secret" {
		t.Errorf("Expected server password to be 'secret', got '%s'", config.Server.Password)
	}
	if config.Server.Timeout != 30.5 {
		t.Errorf("Expected server timeout to be 30.5, got %f", config.Server.Timeout)
	}
	if !config.Server.Enabled {
		t.Errorf("Expected server enabled to be true, got %v", config.Server.Enabled)
	}
	if config.Server.Logging.Level != "debug" {
		t.Errorf("Expected server logging level to be 'debug', got '%s'", config.Server.Logging.Level)
	}
	if config.Server.Logging.File != "/var/log/myapp.log" {
		t.Errorf("Expected server logging file to be '/var/log/myapp.log', got '%s'", config.Server.Logging.File)
	}
	if config.Database.Host != "db.local" {
		t.Errorf("Expected database host to be 'db.local', got '%s'", config.Database.Host)
	}
	if config.Database.Port != 5432 {
		t.Errorf("Expected database port to be 5432, got %d", config.Database.Port)
	}
	if config.Database.Username != "dbadmin" {
		t.Errorf("Expected database username to be 'dbadmin', got '%s'", config.Database.Username)
	}
	if config.Database.Password != "dbsecret" {
		t.Errorf("Expected database password to be 'dbsecret', got '%s'", config.Database.Password)
	}
}

func TestParseINI_InvalidLine(t *testing.T) {
	iniContent := `
		[server]
		host = localhost
		port
	`

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err == nil {
		t.Fatal("Expected error for invalid line, got nil")
	}
}

func TestParseINI_EmptyFile(t *testing.T) {
	iniContent := ``

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err != nil {
		t.Fatalf("Failed to parse empty INI: %v", err)
	}
}
