# Simple INI

Simple INI is a Go library for parsing INI files.

[![GoDoc](https://godoc.org/github.com/BenBurnett/simpleini?status.svg)](https://godoc.org/github.com/BenBurnett/simpleini)
[![Go Report Card](https://goreportcard.com/badge/github.com/BenBurnett/simpleini)](https://goreportcard.com/report/github.com/BenBurnett/simpleini)
[![Build Status](https://github.com/BenBurnett/simpleini/workflows/CI/badge.svg)](https://github.com/BenBurnett/simpleini/actions)
[![Codecov](https://codecov.io/gh/BenBurnett/simpleini/branch/main/graph/badge.svg)](https://codecov.io/gh/BenBurnett/simpleini)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/BenBurnett/simpleini/blob/main/LICENSE)

## ⚠️ Warning: Project in Development

**This project is currently in active development and may undergo significant changes. Use at your own risk.**

---

## Features

- Support for sectionless blocks, section blocks, and sub-section blocks
- Support for multiple data types: int, uint, float, bool, and string
- Handles pointers to structs and basic types
- Supports custom types that implement the `encoding.TextUnmarshaler` interface
- Implicit name mapping from snake_case to PascalCase
- Ability to override implicit name mapping with struct tags

## Installation

To install Simple INI, use `go get`:

```sh
go get github.com/BenBurnett/simpleini
```

## Example

An example usage of Simple INI can be found in the `examples` folder. This example demonstrates how to define a struct with `ini` tags and parse an INI file into that struct.

### Implicit Name Mapping

By default, Simple INI maps snake_case keys in the INI file to PascalCase field names in the Go struct. For example, the key `app_name` in the INI file will map to the field `AppName` in the Go struct.

### Overriding Implicit Name Mapping

You can override the implicit name mapping by using struct tags. For example, if you want the key `ip_address` in the INI file to map to the field `IP` in the Go struct, you can use the following struct tag:

```go
type ServerConfig struct {
	IP *net.IP `ini:"ip_address"`
}
```

## License

This project is licensed under the MIT License. See the LICENSE file for details.
