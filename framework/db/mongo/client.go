package mongo

import (
	"context"
	"errors"
	"fmt"

	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client executes queries and validations against MongoDB.
type Client struct {
	client *mongodriver.Client
}

func New(ctx context.Context, uri string) (*Client, error) {
	if uri == "" {
		return nil, errors.New("mongo uri is required")
	}
	cli, err := mongodriver.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return &Client{client: cli}, nil
}

func (c *Client) Close(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	return c.client.Disconnect(ctx)
}

func (c *Client) Collection(database, collection string) (*mongodriver.Collection, error) {
	if c.client == nil {
		return nil, errors.New("mongo client is nil")
	}
	if database == "" || collection == "" {
		return nil, errors.New("database and collection required")
	}
	return c.client.Database(database).Collection(collection), nil
}

func (c *Client) ValidateCount(ctx context.Context, database, collection string, query any, expected int64) error {
	col, err := c.Collection(database, collection)
	if err != nil {
		return err
	}
	count, err := col.CountDocuments(ctx, query)
	if err != nil {
		return err
	}
	if count != expected {
		return fmt.Errorf("expected %d documents, got %d", expected, count)
	}
	return nil
}

func (c *Client) ValidateContains(ctx context.Context, database, collection string, query any, contains map[string]any) error {
	col, err := c.Collection(database, collection)
	if err != nil {
		return err
	}
	var result map[string]any
	if err := col.FindOne(ctx, query).Decode(&result); err != nil {
		return err
	}
	for key, expected := range contains {
		if val, ok := result[key]; !ok || fmt.Sprint(val) != fmt.Sprint(expected) {
			return fmt.Errorf("field %s expected %v got %v", key, expected, val)
		}
	}
	return nil
}
