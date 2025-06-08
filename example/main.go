package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	sdk "github.com/ip-sonar/ip-sonar-go"
)

func main() {
	// Initialize client
	client, err := sdk.NewClient(sdk.API_SERVER, sdk.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		// Set your API key here if you have one
		//req.Header.Set(sdk.API_KEY_HEADER, "{your_api_key_here}")
		return nil
	}))
	if err != nil {
		slog.Error("create client", slog.Any("err", err))
		return
	}

	// Lookup my IP
	resp, err := client.LookupMy(context.Background(), nil)
	if err != nil {
		slog.Error("lookup my IP", slog.Any("err", err))
		return
	}

	// Check if the response status code is OK (200)
	if resp.StatusCode != 200 {
		slog.Error("lookup my IP not ok", slog.Int("status", resp.StatusCode))
		return
	}

	// Parse the response
	parsed, err := sdk.ParseLookupResponse(resp)
	if err != nil {
		slog.Error("parse lookup response", slog.Any("err", err))
		return
	}

	// Every response field is optional, so check if they are not nil before using them
	if parsed.JSON200.IP != nil {
		slog.Info("IP: " + *parsed.JSON200.IP)
	}
	if parsed.JSON200.CountryCode != nil {
		slog.Info("Country code: " + *parsed.JSON200.CountryCode)
	}
	if parsed.JSON200.CountryName != nil {
		slog.Info("Country name: " + *parsed.JSON200.CountryName)
	}
	if parsed.JSON200.CityName != nil {
		slog.Info("City name: " + *parsed.JSON200.CityName)
	}
	if parsed.JSON200.Latitude != nil {
		slog.Info(fmt.Sprintf("Latitude: %f", *parsed.JSON200.Latitude))
	}
	if parsed.JSON200.Longitude != nil {
		slog.Info(fmt.Sprintf("Longitude: %f", *parsed.JSON200.Longitude))
	}
}
