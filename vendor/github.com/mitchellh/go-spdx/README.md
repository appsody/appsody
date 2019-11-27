# go-spdx [![Godoc](https://godoc.org/github.com/mitchellh/go-spdx?status.svg)](https://godoc.org/github.com/mitchellh/go-spdx)

go-spdx is a Go library for listing and looking up licenses using
[SPDX IDs](https://spdx.org/licenses/). SPDX IDs are an unambiguous way
to reference a specific software license. The IDs are looked up using the
spdx.org website (or custom URLs may be specified). Offline lookup is not
currently supported.

This library does not implement the SPDX document format. SPDX document
parsing and printing are provided by other libraries, including a library
[in the official spdx organization](https://github.com/spdx/tools-go). This
library instead provides the ability to look up licenses via SPDX IDs.

## Usage

```go
// Get the list of all known licenses
list, err := spdx.List()

// Get a single license with more detail such as the license text
lic, err := spdx.License("MIT")

// Create a custom client so you can control the HTTP client or the URLs
// that are used to access licenses.
client := &spdx.Client{ /* ... */ }
client.List()
client.License("MIT")
```
