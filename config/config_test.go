package config

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-github/v72/github"
)

// MockGitHubClient is a mock implementation of the GitHubClient interface
type MockGitHubClient struct {
	repositories map[string]*github.Repository
	updateError  error
}

func NewMockGitHubClient() *MockGitHubClient {
	return &MockGitHubClient{
		repositories: make(map[string]*github.Repository),
	}
}

func (m *MockGitHubClient) GetRepository(ctx context.Context, org, repo string) *github.Repository {
	key := fmt.Sprintf("%s/%s", org, repo)
	return m.repositories[key]
}

func (m *MockGitHubClient) UpdateCustomProperties(ctx context.Context, org, repo string, properties map[string]string) error {
	return m.updateError
}

func (m *MockGitHubClient) AddRepository(org, repo string, customProperties map[string]interface{}) {
	owner := &github.User{Login: github.Ptr(org)}
	repository := &github.Repository{
		Name:             github.Ptr(repo),
		Owner:            owner,
		CustomProperties: customProperties,
	}
	key := fmt.Sprintf("%s/%s", org, repo)
	m.repositories[key] = repository
}

func (m *MockGitHubClient) SetUpdateError(err error) {
	m.updateError = err
}

func TestNewConfig(t *testing.T) {
	mockClient := NewMockGitHubClient()
	config := NewConfig(mockClient)

	if config == nil {
		t.Fatal("NewConfig should return a non-nil Config")
	}

	if config.githubClient != mockClient {
		t.Error("githubClient should be set correctly")
	}

	if config.repositories != nil {
		t.Error("repositories should be nil initially")
	}

	if config.configurationFiles != nil {
		t.Error("configurationFiles should be nil initially")
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name          string
		yamlContent   string
		expectError   bool
		errorContains string
	}{
		{
			name: "valid config",
			yamlContent: `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"
      - name: "org1/repo2"
  - value: "frontend"
    repositories:
      - name: "org2/repo1"`,
			expectError: false,
		},
		{
			name:          "invalid yaml",
			yamlContent:   "invalid: yaml: content: [",
			expectError:   true,
			errorContains: "failed to unmarshal config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfig(NewMockGitHubClient())
			reader := strings.NewReader(tt.yamlContent)

			err := config.LoadConfig(reader)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(config.configurationFiles) != 1 {
					t.Errorf("expected 1 config file, got %d", len(config.configurationFiles))
				}
				if config.configurationFiles[0].PropertyName != "team" {
					t.Errorf("expected property name 'team', got %q", config.configurationFiles[0].PropertyName)
				}
			}
		})
	}
}

func TestIsRepositoryExists(t *testing.T) {
	config := NewConfig(NewMockGitHubClient())

	// Add some test repositories
	owner1 := &github.User{Login: github.Ptr("org1")}
	owner2 := &github.User{Login: github.Ptr("org2")}
	repo1 := &github.Repository{Name: github.Ptr("repo1"), Owner: owner1}
	repo2 := &github.Repository{Name: github.Ptr("repo2"), Owner: owner2}

	config.repositories = []*github.Repository{repo1, repo2}

	tests := []struct {
		org    string
		repo   string
		exists bool
	}{
		{"org1", "repo1", true},
		{"org2", "repo2", true},
		{"org1", "repo2", false},
		{"org3", "repo1", false},
		{"nonexistent", "repo", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s/%s", tt.org, tt.repo), func(t *testing.T) {
			result := config.isRepositoryExists(tt.org, tt.repo)
			if result != tt.exists {
				t.Errorf("expected %v, got %v", tt.exists, result)
			}
		})
	}
}

func TestGenerateRepositories(t *testing.T) {
	tests := []struct {
		name              string
		configContent     string
		mockRepositories  map[string]*github.Repository
		expectError       bool
		errorContains     string
		expectedRepoCount int
	}{
		{
			name:          "no config files",
			expectError:   true,
			errorContains: "no config files loaded",
		},
		{
			name: "invalid repository format",
			configContent: `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "invalidformat"`,
			expectError:   true,
			errorContains: "is not in the format 'org/repo'",
		},
		{
			name: "repository not found",
			configContent: `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"`,
			mockRepositories: map[string]*github.Repository{},
			expectError:      true,
			errorContains:    "repository repo1 not found in organization org1",
		},
		{
			name: "successful repository generation",
			configContent: `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"
      - name: "org1/repo2"`,
			mockRepositories: map[string]*github.Repository{
				"org1/repo1": {
					Name:  github.Ptr("repo1"),
					Owner: &github.User{Login: github.Ptr("org1")},
				},
				"org1/repo2": {
					Name:  github.Ptr("repo2"),
					Owner: &github.User{Login: github.Ptr("org1")},
				},
			},
			expectError:       false,
			expectedRepoCount: 2,
		},
		{
			name: "duplicate repositories",
			configContent: `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"
  - value: "frontend"
    repositories:
      - name: "org1/repo1"`,
			mockRepositories: map[string]*github.Repository{
				"org1/repo1": {
					Name:  github.Ptr("repo1"),
					Owner: &github.User{Login: github.Ptr("org1")},
				},
			},
			expectError:       false,
			expectedRepoCount: 1, // Should not duplicate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockGitHubClient()
			for key := range tt.mockRepositories {
				parts := strings.Split(key, "/")
				mockClient.AddRepository(parts[0], parts[1], nil)
			}

			config := NewConfig(mockClient)

			if tt.configContent != "" {
				reader := strings.NewReader(tt.configContent)
				if err := config.LoadConfig(reader); err != nil {
					t.Fatalf("Failed to load config: %v", err)
				}
			}

			err := config.GenerateRepositories(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(config.repositories) != tt.expectedRepoCount {
					t.Errorf("expected %d repositories, got %d", tt.expectedRepoCount, len(config.repositories))
				}
			}
		})
	}
}

func TestParseCustomPropertyValue(t *testing.T) {
	config := NewConfig(NewMockGitHubClient())

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string value", "test-value", "test-value"},
		{"empty string", "", ""},
		{"nil value", nil, ""},
		{"int value", 123, ""},
		{"bool value", true, ""},
		{"slice value", []string{"a", "b"}, ""},
		{"map value", map[string]string{"key": "value"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.parseCustomPropertyValue(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateDiffs(t *testing.T) {
	tests := []struct {
		name         string
		setupConfig  func() *Config
		expectError  bool
		errorMsg     string
		expectedDiff int
	}{
		{
			name: "no repositories",
			setupConfig: func() *Config {
				return NewConfig(NewMockGitHubClient())
			},
			expectError: true,
			errorMsg:    "no repositories found",
		},
		{
			name: "no diffs needed",
			setupConfig: func() *Config {
				config := NewConfig(NewMockGitHubClient())

				// Load config
				configContent := `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"`
				reader := strings.NewReader(configContent)
				if err := config.LoadConfig(reader); err != nil {
					panic(fmt.Sprintf("LoadConfig failed: %v", err))
				}

				// Add repository with same value
				owner := &github.User{Login: github.Ptr("org1")}
				repo := &github.Repository{
					Name:  github.Ptr("repo1"),
					Owner: owner,
					CustomProperties: map[string]interface{}{
						"team": "backend",
					},
				}
				config.repositories = []*github.Repository{repo}

				return config
			},
			expectError:  false,
			expectedDiff: 0,
		},
		{
			name: "diffs needed",
			setupConfig: func() *Config {
				config := NewConfig(NewMockGitHubClient())

				// Load config
				configContent := `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"
  - value: "frontend"
    repositories:
      - name: "org1/repo2"`
				reader := strings.NewReader(configContent)
				if err := config.LoadConfig(reader); err != nil {
					panic(fmt.Sprintf("LoadConfig failed: %v", err))
				}

				// Add repositories with different values
				owner := &github.User{Login: github.Ptr("org1")}
				repo1 := &github.Repository{
					Name:  github.Ptr("repo1"),
					Owner: owner,
					CustomProperties: map[string]interface{}{
						"team": "old-backend",
					},
				}
				repo2 := &github.Repository{
					Name:  github.Ptr("repo2"),
					Owner: owner,
					CustomProperties: map[string]interface{}{
						"team": "old-frontend",
					},
				}
				config.repositories = []*github.Repository{repo1, repo2}

				return config
			},
			expectError:  false,
			expectedDiff: 2,
		},
		{
			name: "repository name mismatch",
			setupConfig: func() *Config {
				config := NewConfig(NewMockGitHubClient())

				// Load config
				configContent := `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"`
				reader := strings.NewReader(configContent)
				if err := config.LoadConfig(reader); err != nil {
					panic(fmt.Sprintf("LoadConfig failed: %v", err))
				}

				// Add repository with different name
				owner := &github.User{Login: github.Ptr("org1")}
				repo := &github.Repository{
					Name:  github.Ptr("repo2"),
					Owner: owner,
					CustomProperties: map[string]interface{}{
						"team": "old-value",
					},
				}
				config.repositories = []*github.Repository{repo}

				return config
			},
			expectError:  false,
			expectedDiff: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setupConfig()

			diffs, err := config.GenerateDiffs(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(diffs) != tt.expectedDiff {
					t.Errorf("expected %d diffs, got %d", tt.expectedDiff, len(diffs))
				}
			}
		})
	}
}

func TestGenerateDiffsSorting(t *testing.T) {
	config := NewConfig(NewMockGitHubClient())

	// Load config
	configContent := `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org2/repo1"
      - name: "org1/repo2"
      - name: "org1/repo1"`
	reader := strings.NewReader(configContent)
	if err := config.LoadConfig(reader); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Add repositories in unsorted order
	repos := []*github.Repository{
		{
			Name:             github.Ptr("repo1"),
			Owner:            &github.User{Login: github.Ptr("org2")},
			CustomProperties: map[string]interface{}{"team": "old"},
		},
		{
			Name:             github.Ptr("repo2"),
			Owner:            &github.User{Login: github.Ptr("org1")},
			CustomProperties: map[string]interface{}{"team": "old"},
		},
		{
			Name:             github.Ptr("repo1"),
			Owner:            &github.User{Login: github.Ptr("org1")},
			CustomProperties: map[string]interface{}{"team": "old"},
		},
	}
	config.repositories = repos

	diffs, err := config.GenerateDiffs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diffs) != 3 {
		t.Fatalf("expected 3 diffs, got %d", len(diffs))
	}

	// Check sorting: should be org1/repo1, org1/repo2, org2/repo1
	expected := []struct {
		org  string
		repo string
	}{
		{"org1", "repo1"},
		{"org1", "repo2"},
		{"org2", "repo1"},
	}

	for i, expected := range expected {
		if diffs[i].Organization != expected.org || diffs[i].Repository != expected.repo {
			t.Errorf("expected diff[%d] to be %s/%s, got %s/%s",
				i, expected.org, expected.repo,
				diffs[i].Organization, diffs[i].Repository)
		}
	}
}

// TestGenerateDiffsSortingByPropertyName tests sorting by property name when org and repo are the same
func TestGenerateDiffsSortingByPropertyName(t *testing.T) {
	config := NewConfig(NewMockGitHubClient())

	// Load config with multiple properties for the same repo
	configContent1 := `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"`
	reader1 := strings.NewReader(configContent1)
	if err := config.LoadConfig(reader1); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	configContent2 := `property_name: "environment"
values:
  - value: "production"
    repositories:
      - name: "org1/repo1"`
	reader2 := strings.NewReader(configContent2)
	if err := config.LoadConfig(reader2); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Add repository with different values for both properties
	repo := &github.Repository{
		Name:  github.Ptr("repo1"),
		Owner: &github.User{Login: github.Ptr("org1")},
		CustomProperties: map[string]interface{}{
			"team":        "old-team",
			"environment": "old-env",
		},
	}
	config.repositories = []*github.Repository{repo}

	diffs, err := config.GenerateDiffs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diffs) != 2 {
		t.Fatalf("expected 2 diffs, got %d", len(diffs))
	}

	// Check sorting by property name: "environment" should come before "team"
	if diffs[0].PropertyName != "environment" {
		t.Errorf("expected first diff property to be 'environment', got %q", diffs[0].PropertyName)
	}
	if diffs[1].PropertyName != "team" {
		t.Errorf("expected second diff property to be 'team', got %q", diffs[1].PropertyName)
	}
}

// TestGenerateDiffsWithNilCustomProperties tests handling of repositories with nil CustomProperties
func TestGenerateDiffsWithNilCustomProperties(t *testing.T) {
	config := NewConfig(NewMockGitHubClient())

	// Load config
	configContent := `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"`
	reader := strings.NewReader(configContent)
	if err := config.LoadConfig(reader); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Add repository with nil CustomProperties
	repo := &github.Repository{
		Name:             github.Ptr("repo1"),
		Owner:            &github.User{Login: github.Ptr("org1")},
		CustomProperties: nil,
	}
	config.repositories = []*github.Repository{repo}

	diffs, err := config.GenerateDiffs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}

	// When CustomProperties is nil, the property should be treated as empty string
	if diffs[0].OldValue != "" {
		t.Errorf("expected old value to be empty string, got %q", diffs[0].OldValue)
	}
	if diffs[0].NewValue != "backend" {
		t.Errorf("expected new value to be 'backend', got %q", diffs[0].NewValue)
	}
}

// TestGenerateDiffsWithMissingProperty tests handling of repositories with missing specific property
func TestGenerateDiffsWithMissingProperty(t *testing.T) {
	config := NewConfig(NewMockGitHubClient())

	// Load config
	configContent := `property_name: "team"
values:
  - value: "backend"
    repositories:
      - name: "org1/repo1"`
	reader := strings.NewReader(configContent)
	if err := config.LoadConfig(reader); err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Add repository with CustomProperties but without the "team" property
	repo := &github.Repository{
		Name:  github.Ptr("repo1"),
		Owner: &github.User{Login: github.Ptr("org1")},
		CustomProperties: map[string]interface{}{
			"environment": "production",
			"other":       "value",
		},
	}
	config.repositories = []*github.Repository{repo}

	diffs, err := config.GenerateDiffs(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}

	// When the specific property is missing, it should be treated as empty string
	if diffs[0].OldValue != "" {
		t.Errorf("expected old value to be empty string, got %q", diffs[0].OldValue)
	}
	if diffs[0].NewValue != "backend" {
		t.Errorf("expected new value to be 'backend', got %q", diffs[0].NewValue)
	}
}

func TestApplyChange(t *testing.T) {
	tests := []struct {
		name          string
		diff          *PropertyDiff
		updateError   error
		expectError   bool
		errorContains string
	}{
		{
			name:          "nil diff",
			diff:          nil,
			expectError:   true,
			errorContains: "property diff is nil",
		},
		{
			name: "successful update",
			diff: &PropertyDiff{
				Organization: "org1",
				Repository:   "repo1",
				PropertyName: "team",
				OldValue:     "old",
				NewValue:     "new",
			},
			expectError: false,
		},
		{
			name: "update failure",
			diff: &PropertyDiff{
				Organization: "org1",
				Repository:   "repo1",
				PropertyName: "team",
				OldValue:     "old",
				NewValue:     "new",
			},
			updateError:   fmt.Errorf("API error"),
			expectError:   true,
			errorContains: "failed to update property team for repository org1/repo1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := NewMockGitHubClient()
			mockClient.SetUpdateError(tt.updateError)

			config := NewConfig(mockClient)

			err := config.ApplyChange(context.Background(), tt.diff)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestApplyChangeMultiple tests applying multiple changes using the new ApplyChange function
func TestApplyChangeMultiple(t *testing.T) {
	mockClient := NewMockGitHubClient()
	config := NewConfig(mockClient)

	diffs := []*PropertyDiff{
		{
			Organization: "org1",
			Repository:   "repo1",
			PropertyName: "team",
			OldValue:     "old1",
			NewValue:     "new1",
		},
		{
			Organization: "org2",
			Repository:   "repo2",
			PropertyName: "environment",
			OldValue:     "old2",
			NewValue:     "new2",
		},
	}

	// Apply each diff individually
	for _, diff := range diffs {
		err := config.ApplyChange(context.Background(), diff)
		if err != nil {
			t.Errorf("unexpected error applying diff: %v", err)
		}
	}
}

// Test to cover LoadConfig read error
func TestLoadConfigReadError(t *testing.T) {
	config := NewConfig(NewMockGitHubClient())

	// Mock io.Reader that returns an error
	errorReader := &errorReader{}

	err := config.LoadConfig(errorReader)
	if err == nil {
		t.Error("expected error but got none")
	}

	if !strings.Contains(err.Error(), "failed to read config data") {
		t.Errorf("expected error to contain 'failed to read config data', got %q", err.Error())
	}
}

// Mock implementation of io.Reader that returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}
