package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client is a lightweight wrapper around pgx connection pool with helpers used
// by the declarative executor.
type Client struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Client, error) {
	if dsn == "" {
		return nil, errors.New("postgres dsn is required")
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &Client{pool: pool}, nil
}

func (c *Client) Close() {
	if c == nil || c.pool == nil {
		return
	}
	c.pool.Close()
}

func (c *Client) Query(ctx context.Context, schema, table string, filters map[string]any) ([]map[string]any, error) {
	if c.pool == nil {
		return nil, errors.New("postgres client not configured")
	}
	qb := BuildSelect(schema, table, filters)
	rows, err := c.pool.Query(ctx, qb.SQL, qb.Args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols := rows.FieldDescriptions()
	var result []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := map[string]any{}
		for i, col := range cols {
			row[string(col.Name)] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (c *Client) ValidateCount(ctx context.Context, schema, table string, filters map[string]any, expected int) error {
	records, err := c.Query(ctx, schema, table, filters)
	if err != nil {
		return err
	}
	if len(records) != expected {
		return fmt.Errorf("expected %d rows, got %d", expected, len(records))
	}
	return nil
}

func (c *Client) ValidateContains(ctx context.Context, schema, table string, filters map[string]any, expected map[string]any) error {
	records, err := c.Query(ctx, schema, table, filters)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return errors.New("no rows returned for contains validation")
	}
	for key, expectedVal := range expected {
		found := false
		for _, row := range records {
			if v, ok := row[key]; ok && fmt.Sprint(v) == fmt.Sprint(expectedVal) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("expected field %s=%v not found", key, expectedVal)
		}
	}
	return nil
}

// QueryBuilder is a trivial SELECT query generator.
type QueryBuilder struct {
	SQL  string
	Args []any
}

func BuildSelect(schema, table string, filters map[string]any) QueryBuilder {
	fullTable := table
	if schema != "" {
		fullTable = fmt.Sprintf("%s.%s", schema, table)
	}
	qb := QueryBuilder{SQL: fmt.Sprintf("SELECT * FROM %s", fullTable)}
	if len(filters) == 0 {
		return qb
	}
	clauses := make([]string, 0, len(filters))
	args := make([]any, 0, len(filters))
	idx := 1
	for k, v := range filters {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", k, idx))
		args = append(args, v)
		idx++
	}
	qb.SQL = fmt.Sprintf("%s WHERE %s", qb.SQL, strings.Join(clauses, " AND "))
	qb.Args = args
	return qb
}
