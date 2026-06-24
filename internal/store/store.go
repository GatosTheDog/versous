package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type Comment struct {
	ID        string
	Product   string
	Source    string
	Body      string
	Url       string
	Embedding []float32
}

type Store interface {
	UpsertComment(ctx context.Context, c Comment) error
	SimilarComments(ctx context.Context, queryVec []float32, product string, limit int) ([]Comment, error)
	Close()
}

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context, connStr string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("New Postgres: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

func (p *Postgres) UpsertComment(ctx context.Context, c Comment) error {
	_, err := p.pool.Exec(ctx, `
        INSERT INTO comments (id, product, source, body, url, embedding)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE
            SET body = EXCLUDED.body,
                embedding = EXCLUDED.embedding
    `, c.ID, c.Product, c.Source, c.Body, c.Url, pgvector.NewVector(c.Embedding))
	if err != nil {
		return fmt.Errorf("upsert comment: %w", err)
	}
	return nil
}

func (p *Postgres) SimilarComments(ctx context.Context, queryVec []float32, product string, limit int) ([]Comment, error) {
	rows, err := p.pool.Query(ctx, `
        SELECT id, product, source, body, url
        FROM comments
        WHERE product = $2
        ORDER BY embedding <=> $1
        LIMIT $3
    `, pgvector.NewVector(queryVec), product, limit)
	if err != nil {
		return nil, fmt.Errorf("similar comments query: %w", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.Product, &c.Source, &c.Body, &c.Url); err != nil {
			return nil, fmt.Errorf("similar comments scan: %w", err)
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("similar comments rows: %w", err)
	}

	return comments, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}
