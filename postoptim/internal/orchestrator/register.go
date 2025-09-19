package optim

import (
	"optim/internal/postoptim"
	sqlc "optim/internal/sqlc/generate"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Optim struct {
	Queries *sqlc.Queries
	Redis   *redis.Client
	DB      *pgxpool.Pool

	PostOptim *postoptim.PostOptim
}

func NewOptim(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *Optim {
	return &Optim{
		Queries: queries,
		Redis:   redis,
		DB:      db,
	}
}

func (s *Optim) InitOptim() error {

	s.PostOptim = postoptim.NewPostOptim(s.Queries, s.Redis, s.DB)
	err := s.PostOptim.InitPostOptim()
	if err != nil {
		s.PostOptim.StopPostOptim()
		return err
	}

	return nil
}

func (s *Optim) StopOptim() error {

	s.PostOptim.StopPostOptim()

	return nil
}
