package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/google/go-github/v72/github"
)

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	token := "test-token"

	client := NewClient(ctx, token)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.githubClient == nil {
		t.Fatal("githubClient is nil")
	}
}

func TestGetRepository_Success(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/test-org/test-repo" {
			t.Errorf("Expected path /repos/test-org/test-repo, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"id": 1,
			"name": "test-repo",
			"full_name": "test-org/test-repo",
			"owner": {
				"login": "test-org"
			}
		}`
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	repo := c.GetRepository(ctx, "test-org", "test-repo")

	if repo == nil {
		t.Fatal("GetRepository returned nil")
	}
	if repo.GetName() != "test-repo" {
		t.Errorf("Expected repo name 'test-repo', got %s", repo.GetName())
	}
}

func TestGetRepository_Error(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"message": "Not Found"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	repo := c.GetRepository(ctx, "test-org", "nonexistent-repo")

	if repo != nil {
		t.Error("GetRepository should return nil on error")
	}
}

func TestUpdateCustomProperties_Success(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/test-org/test-repo/properties/values" {
			t.Errorf("Expected path /repos/test-org/test-repo/properties/values, got %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	properties := map[string]string{
		"property1": "value1",
		"property2": "value2",
	}

	err := c.UpdateCustomProperties(ctx, "test-org", "test-repo", properties)

	if err != nil {
		t.Errorf("UpdateCustomProperties returned error: %v", err)
	}
}

func TestUpdateCustomProperties_Error(t *testing.T) {
	// Mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"message": "Internal Server Error"}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	properties := map[string]string{
		"property1": "value1",
	}

	err := c.UpdateCustomProperties(ctx, "test-org", "test-repo", properties)

	if err == nil {
		t.Error("UpdateCustomProperties should return error when server returns 500")
	}
}

func TestUpdateCustomProperties_EmptyProperties(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	properties := map[string]string{}

	err := c.UpdateCustomProperties(ctx, "test-org", "test-repo", properties)

	if err != nil {
		t.Errorf("UpdateCustomProperties with empty properties returned error: %v", err)
	}
}

func TestUpdateCustomProperties_NilProperties(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	var properties map[string]string = nil

	err := c.UpdateCustomProperties(ctx, "test-org", "test-repo", properties)

	if err != nil {
		t.Errorf("UpdateCustomProperties with nil properties returned error: %v", err)
	}
}

func TestUpdateCustomProperties_LargeProperties(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/test-org/test-repo/properties/values" {
			t.Errorf("Expected path /repos/test-org/test-repo/properties/values, got %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	// Create a larger set of properties
	properties := make(map[string]string)
	for i := range 10 {
		properties["property"+strconv.Itoa(i)] = "value" + strconv.Itoa(i)
	}

	err := c.UpdateCustomProperties(ctx, "test-org", "test-repo", properties)

	if err != nil {
		t.Errorf("UpdateCustomProperties with large properties returned error: %v", err)
	}
}

// Helper function to parse URL
func mustParseURL(s string) *url.URL {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

// Additional edge case tests

func TestNewClient_WithDifferentTokens(t *testing.T) {
	testCases := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"short token", "abc"},
		{"long token", "ghp_1234567890abcdefghijklmnopqrstuvwxyz1234567890"},
		{"special characters", "token_with-special.chars"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			client := NewClient(ctx, tc.token)

			if client == nil {
				t.Fatal("NewClient returned nil")
			}
			if client.githubClient == nil {
				t.Fatal("githubClient is nil")
			}
		})
	}
}

func TestGetRepository_WithSpecialCharacters(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/repos/test-org/repo-with-special.chars_123"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{
			"id": 1,
			"name": "repo-with-special.chars_123",
			"full_name": "test-org/repo-with-special.chars_123",
			"owner": {
				"login": "test-org"
			}
		}`
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	repo := c.GetRepository(ctx, "test-org", "repo-with-special.chars_123")

	if repo == nil {
		t.Fatal("GetRepository returned nil")
	}
	if repo.GetName() != "repo-with-special.chars_123" {
		t.Errorf("Expected repo name 'repo-with-special.chars_123', got %s", repo.GetName())
	}
}

func TestUpdateCustomProperties_WithSpecialValues(t *testing.T) {
	// Mock server setup
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/test-org/test-repo/properties/values" {
			t.Errorf("Expected path /repos/test-org/test-repo/properties/values, got %s", r.URL.Path)
		}
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH method, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	client := github.NewClient(nil)
	client.BaseURL = mustParseURL(server.URL + "/")
	c := &Client{githubClient: client}

	ctx := context.Background()
	properties := map[string]string{
		"property-with-dashes":      "value-with-dashes",
		"property_with_underscores": "value_with_underscores",
		"property.with.dots":        "value.with.dots",
		"property with spaces":      "value with spaces",
		"property123":               "value123",
		"PROPERTY_UPPERCASE":        "VALUE_UPPERCASE",
		"":                          "", // edge case: empty property name and value
	}

	err := c.UpdateCustomProperties(ctx, "test-org", "test-repo", properties)

	if err != nil {
		t.Errorf("UpdateCustomProperties returned error: %v", err)
	}
}
