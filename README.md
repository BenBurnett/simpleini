# Simple INI

Simple INI is a Go library for parsing INI files using reflection. It provides a simple and flexible way to map INI file keys to Go struct fields.

[![GoDoc](https://godoc.org/github.com/BenBurnett/simpleini?status.svg)](https://godoc.org/github.com/BenBurnett/simpleini)
[![Go Report Card](https://goreportcard.com/badge/github.com/BenBurnett/simpleini)](https://goreportcard.com/report/github.com/BenBurnett/simpleini)
[![Build Status](https://github.com/BenBurnett/simpleini/workflows/CI/badge.svg)](https://github.com/BenBurnett/simpleini/actions)
[![Codecov](https://codecov.io/gh/BenBurnett/simpleini/branch/main/graph/badge.svg)](https://codecov.io/gh/BenBurnett/simpleini)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/BenBurnett/simpleini/blob/main/LICENSE)

## ⚠️ Warning: Project in Development

**This project is currently in active development and may undergo significant changes. Use at your own risk.**

---

## Table of Contents

- [Installation](#installation)
- [Example](#example)
- [Features](#features)
  - [Implicit Key Name Mapping](#implicit-key-name-mapping)
  - [Overriding Implicit Name Mapping](#overriding-implicit-name-mapping)
  - [Default Values](#default-values)
  - [Comments](#comments)
  - [Custom Delimiter](#custom-delimiter)
  - [Sections and Subsections](#sections-and-subsections)
  - [Custom Types](#custom-types)
  - [Multiline](#multiline)
  - [Slices](#slices)
  - [Environment Variable Expansion](#environment-variable-expansion)
  - [Include Directive](#include-directive)
- [Usage](#usage)
- [Error Handling](#error-handling)
- [Contributing](#contributing)
- [License](#license)

## Installation

To install Simple INI, use `go get`:

```sh
go get github.com/BenBurnett/simpleini
```

## Example

An example usage of Simple INI can be found in the `example` folder. This example demonstrates how to define a struct with `ini` tags and parse an INI file into that struct.

## Features

### Implicit Key Name Mapping

By default, Simple INI maps snake_case keys in the INI file to PascalCase field names in the Go struct. For example, the key `app_name` in the INI file will map to the field `AppName` in the Go struct.

```ini
app_name = MyApp
app_version = 1.0.0
```

```go
type AppConfig struct {
	AppName    string
	AppVersion string
}
```

### Overriding Implicit Name Mapping

You can override the implicit name mapping by using struct tags. For example, if you want the key `ip_address` in the INI file to map to the field `IP` in the Go struct, you can use the following struct tag:

```ini
ip_address = 192.168.1.1
server_name = MyServer
maximum_connections = 100
```

```go
type ServerConfig struct {
	IP             *net.IP `ini:"ip_address"`
	ServerName     string
	MaxConnections int     `ini:"maximum_connections"`
}
```

### Default Values

You can specify default values for fields using the `default` struct tag. These values will be used if the corresponding key is not present in the INI file.

```ini
app_name = MyApp
```

```go
type AppConfig struct {
	AppName    string `default:"DefaultApp"`
	AppVersion string `default:"1.0.0"`
}
```

### Comments

Simple INI supports comments. Lines starting with `;` or `#` are treated as comments and ignored.

```ini
; This is a comment
# This is also a comment
app_name = MyApp
```

### Custom Delimiter

You can specify a custom delimiter for key-value pairs in the INI file. By default, the delimiter is `=`.

```ini
app_name: MyApp
app_version: 1.0.0
```

```go
type AppConfig struct {
	AppName    string
	AppVersion string
}

var config AppConfig
simpleini.SetDelimiter(":")
err := simpleini.Parse(strings.NewReader(iniData), &config)
if err != nil {
	log.Fatal(err)
}
fmt.Println(config.AppName)    // Output: MyApp
fmt.Println(config.AppVersion) // Output: 1.0.0
```

### Sections and Subsections

Simple INI supports sections and subsections in the INI file. Sections are defined using square brackets, and subsections can be defined using dot notation.

```ini
app_name = MyApp
app_version = 1.0.0

[database]
host = localhost
port = 5432

[server]
ip_address = 192.168.1.1

[server.logging]
level = debug

[server.logging.file]
path = /var/log/myapp.log
```

```go
type DatabaseConfig struct {
	Host string
	Port int
}

type FileLoggingConfig struct {
	Path string
}

type LoggingConfig struct {
	Level string
	File  FileLoggingConfig
}

type ServerConfig struct {
	IPAddress string
	Logging   LoggingConfig
}

type Config struct {
	AppName    string
	AppVersion string
	Database   DatabaseConfig
	Server     ServerConfig
}
```

### Custom Types

Simple INI supports custom types that implement the `encoding.TextUnmarshaler` interface. This allows you to define custom parsing logic for specific fields.

```ini
date = 2023-10-01
```

```go
type CustomDate struct {
	time.Time
}

func (d *CustomDate) UnmarshalText(text []byte) error {
	parsedTime, err := time.Parse("2006-01-02", string(text))
	if err != nil {
		return err
	}
	d.Time = parsedTime
	return nil
}

type Config struct {
	Date CustomDate
}
```

### Multiline

Simple INI supports multiline values for strings. A multiline value continues when the next line starts with a space or tab.

```ini
description = This is a long description
              that spans multiple lines.
```

```go
type Config struct {
	Description string
}
```

### Slices

Simple INI supports parsing slices from multiline values in the INI file. A multiline value continues when the next line starts with a space or tab.

```ini
servers = server1
          server2
          server3
```

```go
type Config struct {
	Servers []string `ini:"servers"`
}
```

**Note:** A slice of custom types will call the `encoding.TextUnmarshaler` for each value. A single custom type will call it with the entire multiline value and can parse it in any way.

### Environment Variable Expansion

Simple INI supports expanding environment variables in values. Environment variables are referenced using the `${VAR_NAME}` syntax.

```ini
path = ${HOME}/myapp/config
```

```go
type Config struct {
	Path string
}
```

### Include Directive

Simple INI supports including other INI files using the `!include` directive. The included file's content will be parsed as if it were part of the original file.

```ini
other_field = i was included
```
```ini
!include other_config.ini

app_name = MyApp
```

```go
type AppConfig struct {
	AppName string
	OtherField string
}
```

## Usage

This example demonstrates how to use several features of Simple INI, including implicit key name mapping, overriding implicit name mapping, default values, custom types, and environment variable expansion.

```ini
app_name = MyApp
app_version = 1.0.0
date = 2023-10-01

[database]
host = localhost
port = 5432

[server]
ip_address = 192.168.1.1
server_name = MyServer
maximum_connections = 100

[server.logging]
level = debug

[server.logging.file]
path = ${HOME}/myapp/logs/app.log
```

```go
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BenBurnett/simpleini"
)

type CustomDate struct {
	time.Time
}

func (d *CustomDate) UnmarshalText(text []byte) error {
	parsedTime, err := time.Parse("2006-01-02", string(text))
	if err != nil {
		return err
	}
	d.Time = parsedTime
	return nil
}

type DatabaseConfig struct {
	Host string
	Port int
}

type FileLoggingConfig struct {
	Path string
}

type LoggingConfig struct {
	Level string
	File  FileLoggingConfig
}

type ServerConfig struct {
	IPAddress       string `ini:"ip_address"`
	ServerName      string `ini:"server_name"`
	MaxConnections  int    `ini:"maximum_connections"`
	Logging         LoggingConfig
}

type AppConfig struct {
	AppName    string
	AppVersion string
	Date       CustomDate
	Database   DatabaseConfig
	Server     ServerConfig
}

func main() {
	file, err := os.Open("config.ini")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	os.Setenv("HOME", "/home/user")

	var config AppConfig
	err = simpleini.Parse(file, &config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("AppName:", config.AppName)                   // Output: MyApp
	fmt.Println("AppVersion:", config.AppVersion)             // Output: 1.0.0
	fmt.Println("Database Host:", config.Database.Host)       // Output: localhost
	fmt.Println("Database Port:", config.Database.Port)       // Output: 5432
	fmt.Println("Server IP Address:", config.Server.IPAddress) // Output: 192.168.1.1
	fmt.Println("Server Name:", config.Server.ServerName)     // Output: MyServer
	fmt.Println("Max Connections:", config.Server.MaxConnections) // Output: 100
	fmt.Println("Logging Level:", config.Server.Logging.Level) // Output: debug
	fmt.Println("Log File Path:", config.Server.Logging.File.Path) // Output: /home/user/myapp/logs/app.log
	fmt.Println("Date:", config.Date.Format("2006-01-02"))    // Output: 2023-10-01
}
```

## Error Handling

Simple INI returns a slice of errors if any issues are encountered during parsing. You can handle these errors in your code to provide meaningful feedback to the user.

```go
errors := simpleini.Parse(file, &config)
if len(errors) > 0 {
	for _, err := range errors {
		fmt.Println("Error:", err)
	}
}
```

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
