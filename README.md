# Simple INI

Simple INI is a Go library for parsing INI files.

[![GoDoc](https://godoc.org/github.com/BenBurnett/simpleini?status.svg)](https://godoc.org/github.com/BenBurnett/simpleini)
[![Go Report Card](https://goreportcard.com/badge/github.com/BenBurnett/simpleini)](https://goreportcard.com/report/github.com/BenBurnett/simpleini)
[![Build Status](https://github.com/BenBurnett/simpleini/workflows/CI/badge.svg)](https://github.com/BenBurnett/simpleini/actions)
[![Codecov](https://codecov.io/gh/BenBurnett/simpleini/branch/main/graph/badge.svg)](https://codecov.io/gh/BenBurnett/simpleini)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/BenBurnett/simpleini/blob/main/LICENSE)

## Features

- Support for sectionless blocks, section blocks, and sub-section blocks
- Support for multiple data types: int, uint, float, bool, and string
- Handles pointers to structs and basic types
- Supports custom types that implement the `encoding.TextUnmarshaler` interface

## Installation

To install Simple INI, use `go get`:

```sh
go get github.com/BenBurnett/simpleini
```

## Example

An example usage of Simple INI can be found in the `examples` folder. This example demonstrates how to define a struct with `ini` tags and parse an INI file into that struct.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
