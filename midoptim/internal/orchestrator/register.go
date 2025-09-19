package optim

import (
	"midoptim/internal/midoptim"
	sqlc "midoptim/internal/sqlc/generate"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Optim struct {
	Queries *sqlc.Queries
	Redis   *redis.Client
	DB      *pgxpool.Pool

	MidOptim *midoptim.MidOptim
}

func NewOptim(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *Optim {
	return &Optim{
		Queries: queries,
		Redis:   redis,
		DB:      db,
	}
}

func (s *Optim) InitOptim() error {

	s.MidOptim = midoptim.NewMidOptim(s.Queries, s.Redis, s.DB)
	err := s.MidOptim.InitMidOptim()
	if err != nil {
		s.MidOptim.StopMidOptim()
		return err
	}

	return nil
}

func (s *Optim) StopOptim() error {

	s.MidOptim.StopMidOptim()

	return nil
}
