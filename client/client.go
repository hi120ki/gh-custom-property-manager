package client

import (
	"context"

	"github.com/google/go-github/v72/github"
	"golang.org/x/oauth2"
)

type Client struct {
	githubClient *github.Client
}

func NewClient(ctx context.Context, token string) *Client {
	githubClient := github.NewClient(
		oauth2.NewClient(
			ctx, oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		),
	)
	return &Client{githubClient: githubClient}
}

func (c *Client) GetRepository(ctx context.Context, org, repo string) *github.Repository {
	repository, _, err := c.githubClient.Repositories.Get(ctx, org, repo)
	if err != nil {
		return nil
	}
	return repository
}

func (c *Client) UpdateCustomProperties(ctx context.Context, org, repo string, properties map[string]string) error {
	customPropertyValues := make([]*github.CustomPropertyValue, 0, len(properties))
	for propertyName, propertyValue := range properties {
		customPropertyValues = append(customPropertyValues, &github.CustomPropertyValue{
			PropertyName: propertyName,
			Value:        propertyValue,
		})
	}
	_, err := c.githubClient.Repositories.CreateOrUpdateCustomProperties(ctx, org, repo, customPropertyValues)
	return err
}
