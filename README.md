# Simple INI

Simple INI is a Go library for parsing INI files. It supports sectionless blocks, section blocks, sub-section blocks, and various data types including int, float, bool, and string. The library also allows automatic marshalling of data to and from user-defined structs using struct tags and reflection.

## Features

- Support for sectionless blocks, section blocks, and sub-section blocks
- Automatic marshalling of data to and from user-defined structs
- Support for int, float, bool, and string data types

## Installation

To install Simple INI, use `go get`:

```sh
go get github.com/BenBurnett/simpleini
```

## Usage

### Define Your Config Struct

Define a struct that represents your configuration. Use struct tags to specify the INI keys.

```go
package main

import (
    "fmt"
    "os"
    "github.com/BenBurnett/simpleini"
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

func main() {
    file, err := os.Open("config.ini")
    if (err != nil) {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    var config Config
    err = Simple INI.Parse(file, &config)
    if err != nil {
        fmt.Println("Error parsing INI:", err)
        return
    }

    fmt.Printf("Parsed Config: %+v\n", config)
}
```

### Example INI File

```ini
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
```

### Running the Example

Save the above code in a file named `main.go` and the INI content in a file named `config.ini`. Then run the following command:

```sh
go run main.go
```

You should see the parsed configuration printed to the console.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
