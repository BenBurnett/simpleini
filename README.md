# Simple INI

Simple INI is a Go library for parsing INI files.

## Features

- Support for sectionless blocks, section blocks, and sub-section blocks
- Automatic marshalling of data to and from user-defined structs
- Support for multiple data types: int, uint, float, bool, and string
- Handles pointers to structs and basic types

## Installation

To install Simple INI, use `go get`:

```sh
go get github.com/BenBurnett/simpleini
```

## Example

An example usage of Simple INI can be found in the `examples` folder. This example demonstrates how to define a struct with `ini` tags and parse an INI file into that struct.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
