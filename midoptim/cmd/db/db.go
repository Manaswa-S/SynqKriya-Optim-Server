package db

import (
	"context"
	"errors"
	"fmt"
	sqlc "midoptim/internal/sqlc/generate"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type DataStore struct {
	PgPool  *pgxpool.Pool
	Queries *sqlc.Queries
	Redis   *redis.Client
}

func NewDataStore() (*DataStore, error) {

	ds := new(DataStore)
	var err error
	ds.PgPool, ds.Queries, err = InitDB()
	if err != nil {
		return nil, err
	}

	ds.Redis, err = InitRedis()
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func InitDB() (*pgxpool.Pool, *sqlc.Queries, error) {
	fmt.Println("Connecting to Databases and Cache...")

	ctx := context.Background()

	dbConnStr, exists := os.LookupEnv("PG_DB_CONN_STR")
	if !exists {
		return nil, nil, errors.New("pg db conn str not found in env")
	}

	pool, err := pgxpool.New(ctx, dbConnStr)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating database pool: %s", err)
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, nil, errors.New("pgx Pool connection failed : " + err.Error())
	}
	defer conn.Release()

	err = conn.Ping(ctx)
	if err != nil {
		return nil, nil, errors.New("database connection failed : " + err.Error())
	} else {
		fmt.Println("Database connection is alive!")
	}

	queries := sqlc.New(pool)

	return pool, queries, nil
}

func InitRedis() (*redis.Client, error) {

	ctx := context.Background()

	connStr, exists := os.LookupEnv("REDIS_DB_CONN_STR")
	if !exists {
		return nil, errors.New("redis pass not found in env")
	}

	redisOpts, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, errors.New("failed to parse redis conn url : " + err.Error())
	}

	redCl := redis.NewClient(redisOpts)

	if _, err := redCl.Ping(ctx).Result(); err != nil {
		return nil, errors.New("Redis basic access failed (PING failed):" + err.Error())
	}
	fmt.Println("Redis connection is alive!")

	return redCl, nil
}

// Close Data store connections
func Close(ds *DataStore) error {
	fmt.Println("Closing connections of Data stores...")

	if ds.PgPool != nil {
		ds.PgPool.Close()
	}

	if ds.Redis != nil {
		return ds.Redis.Close()
	}

	return nil
}
