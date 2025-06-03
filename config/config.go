package config

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/google/go-github/v72/github"
)

// GitHubClient defines the interface for GitHub operations
type GitHubClient interface {
	GetRepository(ctx context.Context, org, repo string) *github.Repository
	UpdateCustomProperties(ctx context.Context, org, repo string, properties map[string]string) error
}

type Config struct {
	githubClient       GitHubClient
	repositories       []*github.Repository
	configurationFiles []*ConfigFile
}

type ConfigFile struct {
	PropertyName string `yaml:"property_name"`
	Values       []struct {
		Value        string `yaml:"value"`
		Repositories []struct {
			Name string `yaml:"name"`
		} `yaml:"repositories"`
	} `yaml:"values"`
}

type PropertyDiff struct {
	Organization string
	Repository   string
	PropertyName string
	OldValue     string
	NewValue     string
}

func NewConfig(githubClient GitHubClient) *Config {
	return &Config{
		githubClient: githubClient,
	}
}

func (c *Config) LoadConfig(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("failed to read config data: %w", err)
	}

	var configFile ConfigFile
	if err := yaml.Unmarshal(data, &configFile); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	c.configurationFiles = append(c.configurationFiles, &configFile)

	return nil
}

func (c *Config) isRepositoryExists(organizationName, repositoryName string) bool {
	for _, repository := range c.repositories {
		if repository.GetOwner().GetLogin() == organizationName && repository.GetName() == repositoryName {
			return true
		}
	}
	return false
}

func (c *Config) GenerateRepositories(ctx context.Context) error {
	if len(c.configurationFiles) == 0 {
		return fmt.Errorf("no config files loaded")
	}

	for _, configFile := range c.configurationFiles {
		for _, value := range configFile.Values {
			for _, repositoryConfig := range value.Repositories {
				repositoryParts := strings.Split(repositoryConfig.Name, "/")
				if len(strings.Split(repositoryConfig.Name, "/")) != 2 {
					return fmt.Errorf("repository name %s is not in the format 'org/repo'", repositoryConfig.Name)
				}
				organizationName := repositoryParts[0]
				repositoryName := repositoryParts[1]

				if c.isRepositoryExists(organizationName, repositoryName) {
					continue
				}

				repository := c.githubClient.GetRepository(ctx, organizationName, repositoryName)
				if repository == nil {
					return fmt.Errorf("repository %s not found in organization %s", repositoryName, organizationName)
				}
				c.repositories = append(c.repositories, repository)
			}
		}
	}

	return nil
}

func (c *Config) parseCustomPropertyValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func (c *Config) GenerateDiffs(ctx context.Context) ([]*PropertyDiff, error) {
	if len(c.repositories) == 0 {
		return nil, fmt.Errorf("no repositories found")
	}

	var propertyDiffs []*PropertyDiff

	for _, repository := range c.repositories {
		for _, configFile := range c.configurationFiles {
			for _, value := range configFile.Values {
				for _, repositoryConfig := range value.Repositories {
					if repositoryConfig.Name != fmt.Sprintf("%s/%s", repository.GetOwner().GetLogin(), repository.GetName()) {
						continue
					}

					propertyName := configFile.PropertyName
					oldValue := c.parseCustomPropertyValue(repository.CustomProperties[propertyName])
					newValue := value.Value

					if oldValue != newValue {
						propertyDiffs = append(propertyDiffs, &PropertyDiff{
							Organization: repository.GetOwner().GetLogin(),
							Repository:   repository.GetName(),
							PropertyName: propertyName,
							OldValue:     oldValue,
							NewValue:     newValue,
						})
					}
				}
			}
		}
	}

	sort.Slice(propertyDiffs, func(i, j int) bool {
		if propertyDiffs[i].Organization != propertyDiffs[j].Organization {
			return propertyDiffs[i].Organization < propertyDiffs[j].Organization
		}
		if propertyDiffs[i].Repository != propertyDiffs[j].Repository {
			return propertyDiffs[i].Repository < propertyDiffs[j].Repository
		}
		return propertyDiffs[i].PropertyName < propertyDiffs[j].PropertyName
	})

	return propertyDiffs, nil
}

func (c *Config) ApplyChange(ctx context.Context, propertyDiff *PropertyDiff) error {
	if propertyDiff == nil {
		return fmt.Errorf("property diff is nil")
	}

	propertyUpdates := map[string]string{
		propertyDiff.PropertyName: propertyDiff.NewValue,
	}
	if err := c.githubClient.UpdateCustomProperties(ctx, propertyDiff.Organization, propertyDiff.Repository, propertyUpdates); err != nil {
		return fmt.Errorf("failed to update property %s for repository %s/%s: %w", propertyDiff.PropertyName, propertyDiff.Organization, propertyDiff.Repository, err)
	}

	return nil
}
