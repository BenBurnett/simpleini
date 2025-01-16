package simpleini

import (
	"net"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

type DatabaseConfig struct {
	Host     string
	Port     uint
	Username string
	Password *string
	MaxConns int
}

type FileConfig struct {
	Path string
	Size int
}

type LoggingConfig struct {
	Level      string
	File       *string
	FileConfig FileConfig
}

type ServerConfig struct {
	Host        string
	Port        uint
	Username    *string
	Password    string
	Timeout     float64
	Enabled     *bool
	Logging     *LoggingConfig
	IP          *net.IP `ini:"ip_address"` // Use a struct tag name that doesn't match the field name
	Description string
	Notes       string
}

type Config struct {
	AppName  string
	Version  *string
	Server   ServerConfig
	Database DatabaseConfig
	Duration CustomDuration
}

// CustomDuration is a custom type that implements encoding.TextUnmarshaler
type CustomDuration time.Duration

func (d *CustomDuration) UnmarshalText(text []byte) error {
	duration, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}
	*d = CustomDuration(duration)
	return nil
}

// CustomStringSlice is a custom type that implements encoding.TextUnmarshaler
type CustomStringSlice []string

func (s *CustomStringSlice) UnmarshalText(text []byte) error {
	*s = strings.Split(string(text), "\n")
	return nil
}

type CustomSliceConfig struct {
	Values CustomStringSlice `ini:"values"`
}

type DefaultConfig struct {
	Name    string  `ini:"name" default:"default_name"`
	Age     *uint   `ini:"age" default:"25"`
	Score   float64 `ini:"score" default:"75.5"`
	Active  *bool   `ini:"active" default:"true"`
	Comment string  `ini:"comment" default:"default_comment"`
}

type InvalidDefaultIntConfig struct {
	Age int `ini:"age" default:"invalid_int"`
}

type InvalidDefaultUintConfig struct {
	Count uint `ini:"count" default:"invalid_uint"`
}

type InvalidDefaultFloatConfig struct {
	Score float64 `ini:"score" default:"invalid_float"`
}

type InvalidDefaultBoolConfig struct {
	Active bool `ini:"active" default:"invalid_bool"`
}

type InvalidDefaultIntSubConfig struct {
	Server struct {
		Age int `ini:"age" default:"invalid_int"`
	} `ini:"server"`
}

type InvalidDefaultUintSubConfig struct {
	Server struct {
		Count uint `ini:"count" default:"invalid_uint"`
	} `ini:"server"`
}

type InvalidDefaultFloatSubConfig struct {
	Server struct {
		Score float64 `ini:"score" default:"invalid_float"`
	} `ini:"server"`
}

type InvalidDefaultBoolSubConfig struct {
	Server struct {
		Active bool `ini:"active" default:"invalid_bool"`
	} `ini:"server"`
}

func TestParse(t *testing.T) {
	iniContent := `
; This is a comment
# This is another comment

app_name = MyApp
version = 1.0.0
duration = 1h30m

[server]
host = localhost
port = 8080
username = admin
password = secret
timeout = 30.5
enabled = true
ip_address = 192.168.1.1

[server.logging]
level = debug
file = /var/log/myapp.log

[server.logging.file_config]
path = /var/log/myapp.log
size = 1024

[database]
host = db.local
port = 5432
username = dbadmin
password = dbsecret
max_conns = 100
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI: %v", errors)
	}

	checkConfig(t, &config)
}

func TestParse_CustomDelimiter(t *testing.T) {
	iniContent := `
; This is a comment
# This is another comment

app_name: MyApp
version: 1.0.0
duration: 1h30m

[server]
host: localhost
port: 8080
username: admin
password: secret
timeout: 30.5
enabled: true
ip_address: 192.168.1.1

[server.logging]
level: debug
file: /var/log/myapp.log

[server.logging.file_config]
path: /var/log/myapp.log
size: 1024

[database]
host: db.local
port: 5432
username: dbadmin
password: dbsecret
max_conns: 100
`

	config := Config{}
	errors := ParseWithDelimiter(strings.NewReader(iniContent), &config, ":")
	if errors != nil {
		t.Fatalf("Failed to parse INI with custom delimiter: %v", errors)
	}

	checkConfig(t, &config)
}

func checkConfig(t *testing.T, config *Config) {
	t.Helper()

	if config.AppName != "MyApp" {
		t.Errorf("Expected app_name to be 'MyApp', got '%s'", config.AppName)
	}
	if *config.Version != "1.0.0" {
		t.Errorf("Expected version to be '1.0.0', got '%s'", *config.Version)
	}
	if time.Duration(config.Duration) != time.Hour+30*time.Minute {
		t.Errorf("Expected duration to be '1h30m', got '%s'", time.Duration(config.Duration))
	}
	checkServerConfig(t, &config.Server)
	checkDatabaseConfig(t, &config.Database)
}

func checkServerConfig(t *testing.T, server *ServerConfig) {
	t.Helper()

	if server.Host != "localhost" {
		t.Errorf("Expected server host to be 'localhost', got '%s'", server.Host)
	}
	if server.Port != 8080 {
		t.Errorf("Expected server port to be 8080, got %d", server.Port)
	}
	if *server.Username != "admin" {
		t.Errorf("Expected server username to be 'admin', got '%s'", *server.Username)
	}
	if server.Password != "secret" {
		t.Errorf("Expected server password to be 'secret', got '%s'", server.Password)
	}
	if server.Timeout != 30.5 {
		t.Errorf("Expected server timeout to be 30.5, got %f", server.Timeout)
	}
	if !*server.Enabled {
		t.Errorf("Expected server enabled to be true, got %v", *server.Enabled)
	}
	if server.Logging.Level != "debug" {
		t.Errorf("Expected server logging level to be 'debug', got '%s'", server.Logging.Level)
	}
	if *server.Logging.File != "/var/log/myapp.log" {
		t.Errorf("Expected server logging file to be '/var/log/myapp.log', got '%s'", *server.Logging.File)
	}
	if server.IP == nil || server.IP.String() != "192.168.1.1" {
		t.Errorf("Expected server IP to be '192.168.1.1', got '%v'", server.IP)
	}
	if server.Logging.FileConfig.Path != "/var/log/myapp.log" {
		t.Errorf("Expected server logging file path to be '/var/log/myapp.log', got '%s'", server.Logging.FileConfig.Path)
	}
	if server.Logging.FileConfig.Size != 1024 {
		t.Errorf("Expected server logging file size to be 1024, got %d", server.Logging.FileConfig.Size)
	}
}

func checkDatabaseConfig(t *testing.T, database *DatabaseConfig) {
	t.Helper()

	if database.Host != "db.local" {
		t.Errorf("Expected database host to be 'db.local', got '%s'", database.Host)
	}
	if database.Port != 5432 {
		t.Errorf("Expected database port to be 5432, got %d", database.Port)
	}
	if database.Username != "dbadmin" {
		t.Errorf("Expected database username to be 'dbadmin', got '%s'", database.Username)
	}
	if *database.Password != "dbsecret" {
		t.Errorf("Expected database password to be 'dbsecret', got '%s'", *database.Password)
	}
	if database.MaxConns != 100 {
		t.Errorf("Expected database max_conns to be 100, got %d", database.MaxConns)
	}
}

func TestParse_Subsections(t *testing.T) {
	iniContent := `
[server]
host = localhost
port = 8080

[server.logging]
level = info
file = /var/log/server.log
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI: %v", errors)
	}

	if config.Server.Host != "localhost" {
		t.Errorf("Expected server host to be 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected server port to be 8080, got %d", config.Server.Port)
	}
	if config.Server.Logging.Level != "info" {
		t.Errorf("Expected server logging level to be 'info', got '%s'", config.Server.Logging.Level)
	}
	if *config.Server.Logging.File != "/var/log/server.log" {
		t.Errorf("Expected server logging file to be '/var/log/server.log', got '%s'", *config.Server.Logging.File)
	}
}

func TestParse_InvalidLine(t *testing.T) {
	iniContent := `
[server]
host = localhost
port
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil {
		t.Fatal("Expected error for invalid line, got nil")
	}
	for _, err := range errors {
		if strings.Contains(err.Error(), "invalid line format at line 4") {
			return
		}
	}
	t.Fatalf("Expected error for invalid line format at line 4, got %v", errors)
}

func TestParse_EmptyFile(t *testing.T) {
	iniContent := ``

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse empty INI: %v", errors)
	}
}

func TestParse_InvalidLineFormat(t *testing.T) {
	iniContent := `
app_name MyApp
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil {
		t.Fatal("Expected error for invalid line format, got nil")
	}
	for _, err := range errors {
		if strings.Contains(err.Error(), "invalid line format at line 2") {
			return
		}
	}
	t.Fatalf("Expected error for invalid line format at line 2, got %v", errors)
}

func TestParse_InvalidIntValue(t *testing.T) {
	iniContent := `
[database]
max_conns = not_an_int
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil {
		t.Fatal("Expected error for invalid integer value, got nil")
	}
	for _, err := range errors {
		if strings.Contains(err.Error(), "error at line 3: invalid value for field type int") {
			return
		}
	}
	t.Fatalf("Expected error for invalid integer value at line 3, got %v", errors)
}

func TestParse_InvalidUintValue(t *testing.T) {
	iniContent := `
[server]
port = not_a_uint
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil {
		t.Fatal("Expected error for invalid unsigned integer value, got nil")
	}
	for _, err := range errors {
		if strings.Contains(err.Error(), "error at line 3: invalid value for field type uint") {
			return
		}
	}
	t.Fatalf("Expected error for invalid unsigned integer value at line 3, got %v", errors)
}

func TestParse_InvalidFloatValue(t *testing.T) {
	iniContent := `
[server]
timeout = not_a_float
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil {
		t.Fatal("Expected error for invalid float value, got nil")
	}
	for _, err := range errors {
		if strings.Contains(err.Error(), "error at line 3: invalid value for field type float64") {
			return
		}
	}
	t.Fatalf("Expected error for invalid float value at line 3, got %v", errors)
}

func TestParse_InvalidBoolValue(t *testing.T) {
	iniContent := `
[server]
enabled = not_a_bool
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil {
		t.Fatal("Expected error for invalid boolean value, got nil")
	}
	for _, err := range errors {
		if strings.Contains(err.Error(), "error at line 3: invalid value for field type bool") {
			return
		}
	}
	t.Fatalf("Expected error for invalid boolean value at line 3, got %v", errors)
}

func TestParse_UnsupportedFieldType(t *testing.T) {
	type UnsupportedConfig struct {
		Data map[string]string `ini:"data"`
	}

	iniContent := `
data = value
`

	config := UnsupportedConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "unsupported field type") {
		t.Fatalf("Expected error for unsupported field type, got %v", errors)
	}
}

func TestParse_NoMatchingSection(t *testing.T) {
	iniContent := `
[unknown]
key = value
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "no matching field found for section") {
		t.Fatalf("Expected error for no matching field found for section, got %v", errors)
	}
}

func TestParse_FieldForSectionNotStruct(t *testing.T) {
	type InvalidConfig struct {
		Server string `ini:"server"`
	}

	iniContent := `
[server]
host = localhost
`

	config := InvalidConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "field for section 'server' is not a struct") {
		t.Fatalf("Expected error for field for section 'server' not being a struct, got %v", errors)
	}
}

func TestSetConfigValue(t *testing.T) {
	type TestConfig struct {
		Name   *string  `ini:"name"`
		Age    *uint    `ini:"age"`
		Score  *float64 `ini:"score"`
		Active *bool    `ini:"active"`
	}

	config := &TestConfig{}
	err := setConfigValue(config, "", "name", "John Doe")
	if err != nil {
		t.Fatalf("Failed to set name: %v", err)
	}
	if *config.Name != "John Doe" {
		t.Errorf("Expected name to be 'John Doe', got '%s'", *config.Name)
	}

	err = setConfigValue(config, "", "age", "30")
	if err != nil {
		t.Fatalf("Failed to set age: %v", err)
	}
	if *config.Age != 30 {
		t.Errorf("Expected age to be 30, got %d", *config.Age)
	}

	err = setConfigValue(config, "", "score", "95.5")
	if err != nil {
		t.Fatalf("Failed to set score: %v", err)
	}
	if *config.Score != 95.5 {
		t.Errorf("Expected score to be 95.5, got %f", *config.Score)
	}

	err = setConfigValue(config, "", "active", "true")
	if err != nil {
		t.Fatalf("Failed to set active: %v", err)
	}
	if !*config.Active {
		t.Errorf("Expected active to be true, got %v", *config.Active)
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
		Name *string `ini:"name"`
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
	if err == nil || !strings.Contains(err.Error(), "invalid value for field type int") {
		t.Fatalf("Expected error for invalid integer value, got %v", err)
	}
}

func TestSetFieldValue_InvalidUintValue(t *testing.T) {
	var uintValue uint
	err := setFieldValue(reflect.ValueOf(&uintValue).Elem(), "not_a_uint")
	if err == nil || !strings.Contains(err.Error(), "invalid value for field type uint") {
		t.Fatalf("Expected error for invalid unsigned integer value, got %v", err)
	}
}

func TestSetFieldValue_InvalidFloatValue(t *testing.T) {
	var floatValue float64
	err := setFieldValue(reflect.ValueOf(&floatValue).Elem(), "not_a_float")
	if err == nil || !strings.Contains(err.Error(), "invalid value for field type float64") {
		t.Fatalf("Expected error for invalid float value, got %v", err)
	}
}

func TestSetFieldValue_InvalidBoolValue(t *testing.T) {
	var boolValue bool
	err := setFieldValue(reflect.ValueOf(&boolValue).Elem(), "not_a_bool")
	if err == nil || !strings.Contains(err.Error(), "invalid value for field type bool") {
		t.Fatalf("Expected error for invalid boolean value, got %v", err)
	}
}

func TestSetFieldValue_UnsupportedFieldType(t *testing.T) {
	var unsupportedValue map[string]string
	err := setFieldValue(reflect.ValueOf(&unsupportedValue).Elem(), "value")
	if err == nil || !strings.Contains(err.Error(), "unsupported field type") {
		t.Fatalf("Expected error for unsupported field type, got %v", err)
	}
}

func TestParse_EmptySection(t *testing.T) {
	iniContent := `
[server]
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse empty section: %v", errors)
	}
}

func TestParse_EmptyKey(t *testing.T) {
	iniContent := `
[server]
host =
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse empty key: %v", errors)
	}
	if config.Server.Host != "" {
		t.Errorf("Expected server host to be empty, got '%s'", config.Server.Host)
	}
}

func TestParse_EmptyValue(t *testing.T) {
	iniContent := `
[server]
host = 
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse empty value: %v", errors)
	}
	if config.Server.Host != "" {
		t.Errorf("Expected server host to be empty, got '%s'", config.Server.Host)
	}
}

func TestParse_MissingSectionHeader(t *testing.T) {
	iniContent := `
host = localhost
port = 8080
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "no matching field found for key") {
		t.Fatalf("Expected error for no matching field found for key, got %v", errors)
	}
}

func TestParse_CommentOnlyFile(t *testing.T) {
	iniContent := `
; This is a comment
# This is another comment
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse comment-only INI: %v", errors)
	}
}

func TestParse_MultilineString(t *testing.T) {
	iniContent := `
[server]
description = This is a
    multiline
    description
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse multiline string: %v", errors)
	}
	expectedDescription := "This is a\nmultiline\ndescription"
	if config.Server.Description != expectedDescription {
		t.Errorf("Expected server description to be '%s', got '%s'", expectedDescription, config.Server.Description)
	}
}

func TestParse_MultilineInt(t *testing.T) {
	iniContent := `
[database]
max_conns = 100
    200
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type int") {
		t.Fatalf("Expected error for multiline integer value, got %v", errors)
	}
}

func TestParse_MultilineFloat(t *testing.T) {
	iniContent := `
[server]
timeout = 30.5
    40.5
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type float64") {
		t.Fatalf("Expected error for multiline float value, got %v", errors)
	}
}

func TestParse_MultilineBool(t *testing.T) {
	iniContent := `
[server]
enabled = true
    false
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type bool") {
		t.Fatalf("Expected error for multiline boolean value, got %v", errors)
	}
}

func TestParse_MultipleMultilineStrings(t *testing.T) {
	iniContent := `
[server]
description = This is a
    multiline
    description
notes = These are
    additional
    notes

[database]
host = db.local
port = 5432
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse multiple multiline strings: %v", errors)
	}

	expectedDescription := "This is a\nmultiline\ndescription"
	if config.Server.Description != expectedDescription {
		t.Errorf("Expected server description to be '%s', got '%s'", expectedDescription, config.Server.Description)
	}

	expectedNotes := "These are\nadditional\nnotes"
	if config.Server.Notes != expectedNotes {
		t.Errorf("Expected server notes to be '%s', got '%s'", expectedNotes, config.Server.Notes)
	}

	if config.Database.Host != "db.local" {
		t.Errorf("Expected database host to be 'db.local', got '%s'", config.Database.Host)
	}
	if config.Database.Port != 5432 {
		t.Errorf("Expected database port to be 5432, got %d", config.Database.Port)
	}
}

func TestParse_MultilineFollowedByError(t *testing.T) {
	iniContent := `
[server]
description = This is a
    multiline
    description
invalid_field = value
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "no matching field found for key 'invalid_field'") {
		t.Fatalf("Expected error for invalid field after multiline field, got %v", errors)
	}
}

func TestParse_ErroneousMultilineFollowedByValidField(t *testing.T) {
	iniContent := `
[database]
max_conns = 100
    200
invalid_field
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type int") {
		t.Fatalf("Expected error for multiline integer value, got %v", errors)
	}
}

func TestParse_DefaultValues(t *testing.T) {
	iniContent := `
name = custom_name
`

	config := DefaultConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI with default values: %v", errors)
	}

	if config.Name != "custom_name" {
		t.Errorf("Expected name to be 'custom_name', got '%s'", config.Name)
	}
	if *config.Age != 25 {
		t.Errorf("Expected age to be 25, got %d", config.Age)
	}
	if config.Score != 75.5 {
		t.Errorf("Expected score to be 75.5, got %f", config.Score)
	}
	if !*config.Active {
		t.Errorf("Expected active to be true, got %v", config.Active)
	}
	if config.Comment != "default_comment" {
		t.Errorf("Expected comment to be 'default_comment', got '%s'", config.Comment)
	}
}

func TestParse_InvalidDefaultIntValue(t *testing.T) {
	config := InvalidDefaultIntConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type int") {
		t.Fatalf("Expected error for invalid default int value, got %v", errors)
	}
}

func TestParse_InvalidDefaultUintValue(t *testing.T) {
	config := InvalidDefaultUintConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type uint") {
		t.Fatalf("Expected error for invalid default uint value, got %v", errors)
	}
}

func TestParse_InvalidDefaultFloatValue(t *testing.T) {
	config := InvalidDefaultFloatConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type float64") {
		t.Fatalf("Expected error for invalid default float value, got %v", errors)
	}
}

func TestParse_InvalidDefaultBoolValue(t *testing.T) {
	config := InvalidDefaultBoolConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type bool") {
		t.Fatalf("Expected error for invalid default bool value, got %v", errors)
	}
}

func TestParse_InvalidDefaultIntSubValue(t *testing.T) {
	config := InvalidDefaultIntSubConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type int") {
		t.Fatalf("Expected error for invalid default int value in subsection, got %v", errors)
	}
}

func TestParse_InvalidDefaultUintSubValue(t *testing.T) {
	config := InvalidDefaultUintSubConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type uint") {
		t.Fatalf("Expected error for invalid default uint value in subsection, got %v", errors)
	}
}

func TestParse_InvalidDefaultFloatSubValue(t *testing.T) {
	config := InvalidDefaultFloatSubConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type float64") {
		t.Fatalf("Expected error for invalid default float value in subsection, got %v", errors)
	}
}

func TestParse_InvalidDefaultBoolSubValue(t *testing.T) {
	config := InvalidDefaultBoolSubConfig{}
	errors := Parse(strings.NewReader(""), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type bool") {
		t.Fatalf("Expected error for invalid default bool value in subsection, got %v", errors)
	}
}

func TestParse_EnvVarSubstitution(t *testing.T) {
	os.Setenv("APP_NAME", "EnvApp")
	os.Setenv("DB_HOST", "env.db.local")
	defer os.Unsetenv("APP_NAME")
	defer os.Unsetenv("DB_HOST")

	iniContent := `
app_name = ${APP_NAME}

[database]
host = $DB_HOST
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI with env var substitution: %v", errors)
	}

	if config.AppName != "EnvApp" {
		t.Errorf("Expected app_name to be 'EnvApp', got '%s'", config.AppName)
	}
	if config.Database.Host != "env.db.local" {
		t.Errorf("Expected database host to be 'env.db.local', got '%s'", config.Database.Host)
	}
}

func TestParse_CaseInsensitiveKeys(t *testing.T) {
	iniContent := `
App_Name = MyApp
VERSION = 1.0.0
DURATION = 1h30m

[SERVER]
HOST = localhost
PORT = 8080
USERNAME = admin
PASSWORD = secret
TIMEOUT = 30.5
ENABLED = true
IP_ADDRESS = 192.168.1.1

[SERVER.LOGGING]
LEVEL = debug
FILE = /var/log/myapp.log

[SERVER.LOGGING.FILE_CONFIG]
PATH = /var/log/myapp.log
SIZE = 1024

[DATABASE]
HOST = db.local
PORT = 5432
USERNAME = dbadmin
PASSWORD = dbsecret
MAX_CONNS = 100
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI: %v", errors)
	}

	checkConfig(t, &config)
}

func TestParse_CaseInsensitiveSections(t *testing.T) {
	iniContent := `
app_name = MyApp
version = 1.0.0
duration = 1h30m

[Server]
host = localhost
port = 8080
username = admin
password = secret
timeout = 30.5
enabled = true
ip_address = 192.168.1.1

[Server.Logging]
level = debug
file = /var/log/myapp.log

[Server.Logging.File_Config]
path = /var/log/myapp.log
size = 1024

[Database]
host = db.local
port = 5432
username = dbadmin
password = dbsecret
max_conns = 100
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI: %v", errors)
	}

	checkConfig(t, &config)
}

type DefaultSectionConfig struct {
	Server *struct {
		Host    string `ini:"host" default:"localhost"`
		Port    uint   `ini:"port" default:"8080"`
		Enabled *bool  `ini:"enabled" default:"true"`
	} `ini:"server"`
}

func TestParse_DefaultValuesWithPointerSection(t *testing.T) {
	iniContent := ``

	config := DefaultSectionConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI with default values in pointer section: %v", errors)
	}

	if config.Server == nil {
		t.Fatalf("Expected server section to be initialized, got nil")
	}
	if config.Server.Host != "localhost" {
		t.Errorf("Expected server host to be 'localhost', got '%s'", config.Server.Host)
	}
	if config.Server.Port != 8080 {
		t.Errorf("Expected server port to be 8080, got %d", config.Server.Port)
	}
	if config.Server.Enabled == nil || !*config.Server.Enabled {
		t.Errorf("Expected server enabled to be true, got %v", config.Server.Enabled)
	}
}

func TestParse_CustomStringSlice(t *testing.T) {
	iniContent := `
values = value1
    value2
    value3
`

	config := CustomSliceConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI with custom string slice: %v", errors)
	}

	expectedValues := CustomStringSlice{"value1", "value2", "value3"}
	if !reflect.DeepEqual(config.Values, expectedValues) {
		t.Errorf("Expected values to be '%v', got '%v'", expectedValues, config.Values)
	}
}

type PrimitiveSliceConfig struct {
	Ints    []int     `ini:"ints"`
	Uints   []uint    `ini:"uints"`
	Floats  []float64 `ini:"floats"`
	Bools   []bool    `ini:"bools"`
	Strings []string  `ini:"strings"`
}

func TestParse_PrimitiveSlices(t *testing.T) {
	iniContent := `
ints = 1
    2
    3
uints = 4
    5
    6
floats = 1.1
    2.2
    3.3
bools = true
    false
    true
strings = one
    two
    three
`

	config := PrimitiveSliceConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI with primitive slices: %v", errors)
	}

	expectedInts := []int{1, 2, 3}
	if !reflect.DeepEqual(config.Ints, expectedInts) {
		t.Errorf("Expected ints to be '%v', got '%v'", expectedInts, config.Ints)
	}

	expectedUints := []uint{4, 5, 6}
	if !reflect.DeepEqual(config.Uints, expectedUints) {
		t.Errorf("Expected uints to be '%v', got '%v'", expectedUints, config.Uints)
	}

	expectedFloats := []float64{1.1, 2.2, 3.3}
	if !reflect.DeepEqual(config.Floats, expectedFloats) {
		t.Errorf("Expected floats to be '%v', got '%v'", expectedFloats, config.Floats)
	}

	expectedBools := []bool{true, false, true}
	if !reflect.DeepEqual(config.Bools, expectedBools) {
		t.Errorf("Expected bools to be '%v', got '%v'", expectedBools, config.Bools)
	}

	expectedStrings := []string{"one", "two", "three"}
	if !reflect.DeepEqual(config.Strings, expectedStrings) {
		t.Errorf("Expected strings to be '%v', got '%v'", expectedStrings, config.Strings)
	}
}

type CustomTypeSliceConfig struct {
	IPs       []net.IP         `ini:"ips"`
	Durations []CustomDuration `ini:"durations"`
}

func TestParse_CustomTypeSlices(t *testing.T) {
	iniContent := `
ips = 192.168.1.1
    10.0.0.1
    172.16.0.1
durations = 1h
    30m
    15s
`

	config := CustomTypeSliceConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors != nil {
		t.Fatalf("Failed to parse INI with custom type slices: %v", errors)
	}

	expectedIPs := []net.IP{
		net.ParseIP("192.168.1.1"),
		net.ParseIP("10.0.0.1"),
		net.ParseIP("172.16.0.1"),
	}
	if !reflect.DeepEqual(config.IPs, expectedIPs) {
		t.Errorf("Expected IPs to be '%v', got '%v'", expectedIPs, config.IPs)
	}

	expectedDurations := []CustomDuration{
		CustomDuration(time.Hour),
		CustomDuration(30 * time.Minute),
		CustomDuration(15 * time.Second),
	}
	if !reflect.DeepEqual(config.Durations, expectedDurations) {
		t.Errorf("Expected durations to be '%v', got '%v'", expectedDurations, config.Durations)
	}
}

func TestParse_InvalidPrimitiveSliceValues(t *testing.T) {
	iniContent := `
ints = 1
    not_an_int
    3
`

	config := PrimitiveSliceConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid value for field type int") {
		t.Fatalf("Expected error for invalid integer value in slice, got %v", errors)
	}
}

func TestParse_InvalidCustomTypeSliceValues(t *testing.T) {
	iniContent := `
ips = 192.168.1.1
    invalid_ip
    172.16.0.1
`

	config := CustomTypeSliceConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "invalid IP address: invalid_ip") {
		t.Fatalf("Expected error for invalid IP value in slice, got %v", errors)
	}
}

type DuplicateTagConfig struct {
	Field1 string `ini:"duplicate"`
	Field2 string `ini:"duplicate"`
}

type DuplicateNameConfig struct {
	Field1 string
	Field2 string `ini:"Field1"`
}

func TestParse_DuplicateTag(t *testing.T) {
	iniContent := `
duplicate = value
`

	config := DuplicateTagConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "duplicate tag name 'duplicate'") {
		t.Fatalf("Expected error for duplicate tag name, got %v", errors)
	}
}

func TestParse_DuplicateName(t *testing.T) {
	iniContent := `
field1 = value
`

	config := DuplicateNameConfig{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil || !strings.Contains(errors[0].Error(), "duplicate tag name 'Field1'") {
		t.Fatalf("Expected error for duplicate field name, got %v", errors)
	}
}

func TestParse_MultipleErrors(t *testing.T) {
	iniContent := `
[server]
port = not_a_uint
timeout = not_a_float
enabled = not_a_bool
`

	config := Config{}
	errors := Parse(strings.NewReader(iniContent), &config)
	if errors == nil {
		t.Fatal("Expected multiple parsing errors, got nil")
	}
	expectedErrors := []string{
		"error at line 3: invalid value for field type uint",
		"error at line 4: invalid value for field type float64",
		"error at line 5: invalid value for field type bool",
	}
	for _, expectedError := range expectedErrors {
		found := false
		for _, err := range errors {
			if strings.Contains(err.Error(), expectedError) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Expected error %q, but it was not found in %v", expectedError, errors)
		}
	}
}
