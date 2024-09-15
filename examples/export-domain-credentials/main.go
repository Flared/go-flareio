//go:build go1.23

package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Flared/go-flareio"
)

type CredentialsResponse struct {
	Items []Credential `json:"items"`
}

type CredentialSource struct {
	Id string `json:"id"`
}

type Credential struct {
	Id           int               `json:"id"`
	Source       *CredentialSource `json:"source"`
	IdentityName string            `json:"identity_name"`
	Hash         string            `json:"hash"`
}

func exportDomainCredentials(
	client *flareio.ApiClient,
	domain string,
) error {
	csvWriter := csv.NewWriter(os.Stdout)

	for result, err := range client.IterGet(
		"/leaksdb/v2/credentials/by_domain/"+url.QueryEscape(domain),
		nil,
	) {
		// Rate Limiting
		time.Sleep(time.Second * 1)

		if err != nil {
			return fmt.Errorf("failed to fetch page: %w", err)
		}

		var credentialsResponse CredentialsResponse
		if err := json.NewDecoder(result.Response.Body).Decode(&credentialsResponse); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		for _, credential := range credentialsResponse.Items {
			if err := csvWriter.Write(
				[]string{
					strconv.Itoa(credential.Id),
					credential.Source.Id,
					credential.IdentityName,
					credential.Hash,
				},
			); err != nil {
				return fmt.Errorf("failed to output record: %w", err)
			}
		}

		csvWriter.Flush()
		if err := csvWriter.Error(); err != nil {
			return fmt.Errorf("failed to flush writer: %w", err)
		}

		if err := result.Response.Body.Close(); err != nil {
			return fmt.Errorf("failed to close response: %w", err)
		}
	}

	return nil
}

func main() {
	client := flareio.NewApiClient(
		os.Getenv("FLARE_API_KEY"),
	)
	if err := exportDomainCredentials(client, "scatterholt.com"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
