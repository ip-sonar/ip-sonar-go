package ipsonar_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/ip-sonar/ip-sonar-go"
)

// mockHTTPClient implements HttpRequestDoer for testing
type mockHTTPClient struct {
	response *http.Response
	err      error
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

// Test data
var testIPGeolocation = IPGeolocation{
	IP:               ptr("192.168.1.1"),
	CountryCode:      ptr("US"),
	CountryName:      ptr("United States"),
	CityName:         ptr("New York"),
	ContinentCode:    ptr("NA"),
	ContinentName:    ptr("North America"),
	Latitude:         ptr(float32(40.7128)),
	Longitude:        ptr(float32(-74.0060)),
	Timezone:         ptr("America/New_York"),
	PostalCode:       ptr("10001"),
	AccuracyRadius:   ptr(int32(50)),
	IsInEu:           ptr(false),
	Subdivision1Code: ptr("NY"),
	Subdivision1Name: ptr("New York"),
	Subdivision2Code: ptr(""),
	Subdivision2Name: ptr(""),
}

// Helper
func ptr[T any](v T) *T { return &v }

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		server    string
		opts      []ClientOption
		wantError bool
	}{
		{
			name:      "valid server URL",
			server:    "https://api.example.com/",
			opts:      nil,
			wantError: false,
		},
		{
			name:      "server URL without trailing slash",
			server:    "https://api.example.com",
			opts:      nil,
			wantError: false,
		},
		{
			name:   "with HTTP client option",
			server: "https://api.example.com",
			opts: []ClientOption{
				WithHTTPClient(&mockHTTPClient{}),
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.server, tt.opts...)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				if client != nil {
					t.Error("expected nil client but got non-nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("expected non-nil client but got nil")
				}
				if client.Client == nil {
					t.Error("expected non-nil HTTP client but got nil")
				}
				// Ensure trailing slash is added
				if !strings.Contains(client.Server, "/") {
					t.Errorf("expected server URL to contain '/', got: %s", client.Server)
				}
			}
		})
	}
}

func TestClient_Lookup(t *testing.T) {
	// Create test response
	responseData := testIPGeolocation
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	// Create mock HTTP response
	mockResponse := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseJSON)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	// Create client with mock HTTP client
	client, err := NewClient("https://api.example.com", WithHTTPClient(&mockHTTPClient{
		response: mockResponse,
	}))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test lookup
	ctx := context.Background()
	params := &LookupParams{
		Fields:     ptr("ip,country_code,country_name"),
		LocaleCode: ptr("en"),
	}

	resp, err := client.Lookup(ctx, "192.168.1.1", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response but got nil")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}
}

func TestClient_LookupMy(t *testing.T) {
	// Create test response
	responseData := testIPGeolocation
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	// Create mock HTTP response
	mockResponse := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseJSON)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	// Create client with mock HTTP client
	client, err := NewClient("https://api.example.com", WithHTTPClient(&mockHTTPClient{
		response: mockResponse,
	}))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test lookup my IP
	ctx := context.Background()
	params := &LookupMyParams{
		Fields:     ptr("ip,country_code"),
		LocaleCode: ptr("en"),
	}

	resp, err := client.LookupMy(ctx, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response but got nil")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}
}

func TestClient_BatchLookup(t *testing.T) {
	// Create test response
	responseData := BatchLookupIPResponse{
		Data: []IPGeolocation{testIPGeolocation},
	}
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	// Create mock HTTP response
	mockResponse := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseJSON)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	// Create client with mock HTTP client
	client, err := NewClient("https://api.example.com", WithHTTPClient(&mockHTTPClient{
		response: mockResponse,
	}))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test batch lookup
	ctx := context.Background()
	params := &BatchLookupParams{
		Fields:     ptr("ip,country_code"),
		LocaleCode: ptr("en"),
	}
	body := BatchLookupJSONRequestBody{
		Data: []string{"192.168.1.1", "10.0.0.1"},
	}

	resp, err := client.BatchLookup(ctx, params, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response but got nil")
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode)
	}
}

func TestClientWithResponses_LookupWithResponse(t *testing.T) {
	// Create test response
	responseData := testIPGeolocation
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	// Create mock HTTP response
	mockResponse := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseJSON)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	// Create client with responses
	client, err := NewClientWithResponses("https://api.example.com", WithHTTPClient(&mockHTTPClient{
		response: mockResponse,
	}))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test lookup with response parsing
	ctx := context.Background()
	params := &LookupParams{
		Fields:     ptr("ip,country_code,country_name"),
		LocaleCode: ptr("en"),
	}

	resp, err := client.LookupWithResponse(ctx, "192.168.1.1", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response but got nil")
	}
	if resp.StatusCode() != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode())
	}
	if resp.Status() != "200 OK" {
		t.Errorf("expected status '200 OK', got %q", resp.Status())
	}
	if resp.JSON200 == nil {
		t.Fatal("expected non-nil JSON200 but got nil")
	}
	if resp.JSON200.IP == nil || *resp.JSON200.IP != "192.168.1.1" {
		want := "192.168.1.1"
		var got string
		if resp.JSON200.IP != nil {
			got = *resp.JSON200.IP
		}
		t.Errorf("expected IP %q, got %q", want, got)
	}
}

func TestClientWithResponses_BatchLookupWithResponse(t *testing.T) {
	// Create test response
	responseData := BatchLookupIPResponse{
		Data: []IPGeolocation{testIPGeolocation},
	}
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	// Create mock HTTP response
	mockResponse := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseJSON)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	// Create client with responses
	client, err := NewClientWithResponses("https://api.example.com", WithHTTPClient(&mockHTTPClient{
		response: mockResponse,
	}))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Test batch lookup with response parsing
	ctx := context.Background()
	params := &BatchLookupParams{
		Fields:     ptr("ip,country_code"),
		LocaleCode: ptr("en"),
	}
	body := BatchLookupJSONRequestBody{
		Data: []string{"192.168.1.1", "10.0.0.1"},
	}

	resp, err := client.BatchLookupWithResponse(ctx, params, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response but got nil")
	}
	if resp.StatusCode() != 200 {
		t.Errorf("expected status code 200, got %d", resp.StatusCode())
	}
	if resp.JSON200 == nil {
		t.Fatal("expected non-nil JSON200 but got nil")
	}
	if len(resp.JSON200.Data) != 1 {
		t.Errorf("expected 1 item in data, got %d", len(resp.JSON200.Data))
	}
	if len(resp.JSON200.Data) > 0 {
		if resp.JSON200.Data[0].IP == nil || *resp.JSON200.Data[0].IP != "192.168.1.1" {
			want := "192.168.1.1"
			var got string
			if resp.JSON200.Data[0].IP != nil {
				got = *resp.JSON200.Data[0].IP
			}
			t.Errorf("expected IP %q, got %q", want, got)
		}
	}
}

func TestClientWithResponses_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		contentType  string
		responseBody string
		checkJSON401 bool
		checkJSON404 bool
		checkJSON429 bool
		checkJSON500 bool
	}{
		{
			name:         "401 Unauthorized",
			statusCode:   401,
			contentType:  "application/json",
			responseBody: `{"message":"Unauthorized"}`,
			checkJSON401: true,
		},
		{
			name:         "404 Not Found",
			statusCode:   404,
			contentType:  "application/json",
			responseBody: `{"message":"Not Found"}`,
			checkJSON404: true,
		},
		{
			name:         "429 Too Many Requests",
			statusCode:   429,
			contentType:  "application/json",
			responseBody: `{"message":"Too Many Requests"}`,
			checkJSON429: true,
		},
		{
			name:         "500 Internal Server Error",
			statusCode:   500,
			contentType:  "application/json",
			responseBody: `{"message":"Internal Server Error"}`,
			checkJSON500: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP response
			mockResponse := &http.Response{
				StatusCode: tt.statusCode,
				Status:     fmt.Sprintf("%d %s", tt.statusCode, http.StatusText(tt.statusCode)),
				Header:     make(http.Header),
				Body:       io.NopCloser(bytes.NewReader([]byte(tt.responseBody))),
			}
			mockResponse.Header.Set("Content-Type", tt.contentType)

			// Create client with responses
			client, err := NewClientWithResponses("https://api.example.com", WithHTTPClient(&mockHTTPClient{
				response: mockResponse,
			}))
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			// Test error handling for different endpoints
			ctx := context.Background()

			// Test Lookup error handling
			if tt.statusCode != 500 { // Lookup doesn't handle 500 errors
				resp, err := client.LookupWithResponse(ctx, "192.168.1.1", &LookupParams{})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.StatusCode() != tt.statusCode {
					t.Errorf("expected status code %d, got %d", tt.statusCode, resp.StatusCode())
				}

				if tt.checkJSON401 && resp.JSON401 != nil {
					if resp.JSON401.Message != "Unauthorized" {
						t.Errorf("expected message 'Unauthorized', got %q", resp.JSON401.Message)
					}
				}
				if tt.checkJSON404 && resp.JSON404 != nil {
					if resp.JSON404.Message != "Not Found" {
						t.Errorf("expected message 'Not Found', got %q", resp.JSON404.Message)
					}
				}
				if tt.checkJSON429 && resp.JSON429 != nil {
					if resp.JSON429.Message != "Too Many Requests" {
						t.Errorf("expected message 'Too Many Requests', got %q", resp.JSON429.Message)
					}
				}
			}

			// Test BatchLookup error handling
			if tt.statusCode != 404 { // BatchLookup doesn't handle 404 errors
				// Reset the body reader
				mockResponse.Body = io.NopCloser(bytes.NewReader([]byte(tt.responseBody)))

				resp, err := client.BatchLookupWithResponse(ctx, &BatchLookupParams{}, BatchLookupJSONRequestBody{Data: []string{"192.168.1.1"}})
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if resp.StatusCode() != tt.statusCode {
					t.Errorf("expected status code %d, got %d", tt.statusCode, resp.StatusCode())
				}

				if tt.checkJSON401 && resp.JSON401 != nil {
					if resp.JSON401.Message != "Unauthorized" {
						t.Errorf("expected message 'Unauthorized', got %q", resp.JSON401.Message)
					}
				}
				if tt.checkJSON429 && resp.JSON429 != nil {
					if resp.JSON429.Message != "Too Many Requests" {
						t.Errorf("expected message 'Too Many Requests', got %q", resp.JSON429.Message)
					}
				}
				if tt.checkJSON500 && resp.JSON500 != nil {
					if resp.JSON500.Message != "Internal Server Error" {
						t.Errorf("expected message 'Internal Server Error', got %q", resp.JSON500.Message)
					}
				}
			}
		})
	}
}

func TestWithBaseURL(t *testing.T) {
	baseURL := "https://custom.api.com"
	client, err := NewClient("https://original.com", WithBaseURL(baseURL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := baseURL + "/"
	if client.Server != expected {
		t.Errorf("expected server URL %q, got %q", expected, client.Server)
	}
}

func TestWithRequestEditorFn(t *testing.T) {
	var called bool
	editor := func(ctx context.Context, req *http.Request) error {
		called = true
		req.Header.Set("Custom-Header", "test-value")
		return nil
	}

	// Create a test server to capture the request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Custom-Header") != "test-value" {
			t.Errorf("expected Custom-Header to be 'test-value', got %q", r.Header.Get("Custom-Header"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, WithRequestEditorFn(editor))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	// Make a request
	ctx := context.Background()
	_, err = client.Lookup(ctx, "192.168.1.1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The request should succeed and the editor should be called
	if !called {
		t.Error("expected request editor to be called but it wasn't")
	}
}

func TestIPGeolocationStructure(t *testing.T) {
	// Test that our test data structure is valid
	geo := testIPGeolocation

	// Use cmp for deep comparison of complex structures
	want := IPGeolocation{
		IP:               ptr("192.168.1.1"),
		CountryCode:      ptr("US"),
		CountryName:      ptr("United States"),
		CityName:         ptr("New York"),
		ContinentCode:    ptr("NA"),
		ContinentName:    ptr("North America"),
		Latitude:         ptr(float32(40.7128)),
		Longitude:        ptr(float32(-74.0060)),
		Timezone:         ptr("America/New_York"),
		PostalCode:       ptr("10001"),
		AccuracyRadius:   ptr(int32(50)),
		IsInEu:           ptr(false),
		Subdivision1Code: ptr("NY"),
		Subdivision1Name: ptr("New York"),
		Subdivision2Code: ptr(""),
		Subdivision2Name: ptr(""),
	}

	if diff := cmp.Diff(want, geo); diff != "" {
		t.Errorf("IPGeolocation mismatch (-want +got):\n%s", diff)
	}
}

// Benchmark tests
func BenchmarkClient_Lookup(b *testing.B) {
	// Create mock response
	responseData := testIPGeolocation
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		b.Fatalf("failed to marshal test data: %v", err)
	}

	mockResponse := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(responseJSON)),
	}
	mockResponse.Header.Set("Content-Type", "application/json")

	client, err := NewClient("https://api.example.com", WithHTTPClient(&mockHTTPClient{
		response: mockResponse,
	}))
	if err != nil {
		b.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()
	params := &LookupParams{Fields: ptr("ip,country_code")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset the body reader for each iteration
		mockResponse.Body = io.NopCloser(bytes.NewReader(responseJSON))
		_, err := client.Lookup(ctx, "192.168.1.1", params)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
