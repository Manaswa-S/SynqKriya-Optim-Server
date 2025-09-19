package midoptim

import (
	"context"
	sqlc "midoptim/internal/sqlc/generate"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type MidOptim struct {
	Queries *sqlc.Queries
	Redis   *redis.Client
	DB      *pgxpool.Pool

	Manager *Manager
}

func NewMidOptim(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *MidOptim {
	return &MidOptim{
		Queries: queries,
		Redis:   redis,
		DB:      db,
	}
}

type Manager struct {
	pullCalcChan chan *YoloDSReq

	ctx    context.Context
	ctxCnl context.CancelFunc
	wg     sync.WaitGroup
	done   chan struct{}

	redisChanGetInp string
	redisChanSetOut string

	occupancyMap map[int64]int64
}
