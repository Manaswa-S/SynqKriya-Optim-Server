package postoptim

import (
	"context"
	sqlc "optim/internal/sqlc/generate"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PostOptim struct {
	Queries *sqlc.Queries
	Redis   *redis.Client
	DB      *pgxpool.Pool

	Manager *Manager
}

func NewPostOptim(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *PostOptim {
	return &PostOptim{
		Queries: queries,
		Redis:   redis,
		DB:      db,
	}
}

type Manager struct {
	pullPolicyChan chan *MLReq

	ctx    context.Context
	ctxCnl context.CancelFunc
	wg     sync.WaitGroup
	done   chan struct{}

	redisChanGetInp string
	redisChanSetOut string
}
