package simpleini

import (
	"reflect"
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

func TestParse(t *testing.T) {
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

func TestParse_InvalidLine(t *testing.T) {
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

func TestParse_EmptyFile(t *testing.T) {
	iniContent := ``

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err != nil {
		t.Fatalf("Failed to parse empty INI: %v", err)
	}
}

func TestParse_InvalidLineFormat(t *testing.T) {
	iniContent := `
		app_name MyApp
	`

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err == nil || !strings.Contains(err.Error(), "invalid line format") {
		t.Fatalf("Expected error for invalid line format, got %v", err)
	}
}

func TestParse_InvalidIntValue(t *testing.T) {
	iniContent := `
		[server]
		port = not_an_int
	`

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err == nil || !strings.Contains(err.Error(), "invalid integer value") {
		t.Fatalf("Expected error for invalid integer value, got %v", err)
	}
}

func TestParse_InvalidFloatValue(t *testing.T) {
	iniContent := `
		[server]
		timeout = not_a_float
	`

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err == nil || !strings.Contains(err.Error(), "invalid float value") {
		t.Fatalf("Expected error for invalid float value, got %v", err)
	}
}

func TestParse_InvalidBoolValue(t *testing.T) {
	iniContent := `
		[server]
		enabled = not_a_bool
	`

	config := Config{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err == nil || !strings.Contains(err.Error(), "invalid boolean value") {
		t.Fatalf("Expected error for invalid boolean value, got %v", err)
	}
}

func TestParse_UnsupportedFieldType(t *testing.T) {
	type UnsupportedConfig struct {
		Data []string `ini:"data"`
	}

	iniContent := `
		data = value
	`

	config := UnsupportedConfig{}
	err := Parse(strings.NewReader(iniContent), &config)
	if err == nil || !strings.Contains(err.Error(), "unsupported field type") {
		t.Fatalf("Expected error for unsupported field type, got %v", err)
	}
}
func TestSetConfigValue(t *testing.T) {
	type TestConfig struct {
		Name   string  `ini:"name"`
		Age    int     `ini:"age"`
		Score  float64 `ini:"score"`
		Active bool    `ini:"active"`
	}

	config := &TestConfig{}
	err := setConfigValue(config, "", "name", "John Doe")
	if err != nil {
		t.Fatalf("Failed to set name: %v", err)
	}
	if config.Name != "John Doe" {
		t.Errorf("Expected name to be 'John Doe', got '%s'", config.Name)
	}

	err = setConfigValue(config, "", "age", "30")
	if err != nil {
		t.Fatalf("Failed to set age: %v", err)
	}
	if config.Age != 30 {
		t.Errorf("Expected age to be 30, got %d", config.Age)
	}

	err = setConfigValue(config, "", "score", "95.5")
	if err != nil {
		t.Fatalf("Failed to set score: %v", err)
	}
	if config.Score != 95.5 {
		t.Errorf("Expected score to be 95.5, got %f", config.Score)
	}

	err = setConfigValue(config, "", "active", "true")
	if err != nil {
		t.Fatalf("Failed to set active: %v", err)
	}
	if !config.Active {
		t.Errorf("Expected active to be true, got %v", config.Active)
	}

	err = setConfigValue(config, "", "unknown", "value")
	if err == nil {
		t.Fatal("Expected error for unknown field, got nil")
	}
}

func TestSetConfigValue_InvalidConfigType(t *testing.T) {
	config := "invalid"
	err := setConfigValue(config, "", "name", "John Doe")
	if err == nil || !strings.Contains(err.Error(), "configuration must be a pointer to a struct") {
		t.Fatalf("Expected error for invalid config type, got %v", err)
	}
}

func TestSetStructValue_NoMatchingField(t *testing.T) {
	type TestConfig struct {
		Name string `ini:"name"`
	}

	config := &TestConfig{}
	err := setStructValue(reflect.ValueOf(config).Elem(), "unknown", "value")
	if err == nil || !strings.Contains(err.Error(), "no matching field found for key") {
		t.Fatalf("Expected error for no matching field, got %v", err)
	}
}

func TestSetFieldValue_InvalidIntValue(t *testing.T) {
	var intValue int
	err := setFieldValue(reflect.ValueOf(&intValue).Elem(), "not_an_int")
	if err == nil || !strings.Contains(err.Error(), "invalid integer value") {
		t.Fatalf("Expected error for invalid integer value, got %v", err)
	}
}

func TestSetFieldValue_InvalidFloatValue(t *testing.T) {
	var floatValue float64
	err := setFieldValue(reflect.ValueOf(&floatValue).Elem(), "not_a_float")
	if err == nil || !strings.Contains(err.Error(), "invalid float value") {
		t.Fatalf("Expected error for invalid float value, got %v", err)
	}
}

func TestSetFieldValue_InvalidBoolValue(t *testing.T) {
	var boolValue bool
	err := setFieldValue(reflect.ValueOf(&boolValue).Elem(), "not_a_bool")
	if err == nil || !strings.Contains(err.Error(), "invalid boolean value") {
		t.Fatalf("Expected error for invalid boolean value, got %v", err)
	}
}

func TestSetFieldValue_UnsupportedFieldType(t *testing.T) {
	var unsupportedValue []string
	err := setFieldValue(reflect.ValueOf(&unsupportedValue).Elem(), "value")
	if err == nil || !strings.Contains(err.Error(), "unsupported field type") {
		t.Fatalf("Expected error for unsupported field type, got %v", err)
	}
}
