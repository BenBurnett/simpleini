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
