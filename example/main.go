package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/BenBurnett/simpleini"
)

type DatabaseConfig struct {
	Host     string
	Port     uint
	Username string
	Password *string
	MaxConns int
}

type LoggingConfig struct {
	Level string
	File  *string
}

type ServerConfig struct {
	Host     string
	Port     uint
	Username *string
	Password string
	Timeout  float64
	Enabled  *bool
	Logging  *LoggingConfig
	IP       *net.IP `ini:"ip_address"` // Use a struct tag name that doesn't match the field name
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

func main() {
	file, err := os.Open("config.ini")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	var config Config
	err = simpleini.Parse(file, &config)
	if err != nil {
		fmt.Println("Error parsing INI:", err)
		return
	}

	fmt.Printf("Parsed Config: %+v\n", config)
}
