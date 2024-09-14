# go-flareio

[![Go Reference](https://pkg.go.dev/badge/github.com/Flared/go-flareio.svg)](https://pkg.go.dev/github.com/Flared/go-flareio)

`flareio` is a light [Flare API](https://api.docs.flare.io/) SDK that wraps around [net/http.Client](https://pkg.go.dev/net/http#Client) and automatically manages authentication.

It exposes methods that are similar to `net/http.Client` with the exception that they accept paths as parameters instead of full URLs.

Usage examples and use cases are documented in the [Flare API documentation](https://api.docs.flare.io/sdk/go).

## Contributing

- `make test` will run tests
- `make format` format will format the code
- `make lint` will run typechecking + linting


## Basic Usage

```go
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Flared/go-flareio"
)

func main() {
	client := flareio.NewApiClient(
		os.Getenv("FLARE_API_KEY"),
	)
	resp, err := client.Get(
		"/leaksdb/v2/sources", nil,
	)
	if err != nil {
		fmt.Printf("failed to get sources: %s\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		fmt.Printf("failed to print sources: %s\n", err)
		os.Exit(1)
	}
}
```
