package api

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stephenafamo/bob"
)

type Server struct {
	db      bob.DB
	storage Storage
}

var _ ServerInterface = &Server{}

func NewServer(ctx context.Context, dsn string) (*Server, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return &Server{
		db: bob.NewDB(stdlib.OpenDBFromPool(pool)),
	}, nil
}

func (s *Server) Close() error {
	return s.db.Close()
}
