package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/BenBurnett/simpleini"
)

type DatabaseConfig struct {
	Host     string  `ini:"host"`
	Port     uint    `ini:"port"`
	Username string  `ini:"username"`
	Password *string `ini:"password"`
	MaxConns int     `ini:"max_conns"`
}

type LoggingConfig struct {
	Level string  `ini:"level"`
	File  *string `ini:"file"`
}

type ServerConfig struct {
	Host     string         `ini:"host"`
	Port     uint           `ini:"port"`
	Username *string        `ini:"username"`
	Password string         `ini:"password"`
	Timeout  float64        `ini:"timeout"`
	Enabled  *bool          `ini:"enabled"`
	Logging  *LoggingConfig `ini:"logging"`
	IP       *net.IP        `ini:"ip"`
}

type Config struct {
	AppName  string         `ini:"app_name"`
	Version  *string        `ini:"version"`
	Server   ServerConfig   `ini:"server"`
	Database DatabaseConfig `ini:"database"`
	Duration CustomDuration `ini:"duration"`
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
