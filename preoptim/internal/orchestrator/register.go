package optim

import (
	"context"
	"fmt"
	"optim/internal/preoptim"
	rpcserver "optim/internal/preoptim/rpcServer"
	sqlc "optim/internal/sqlc/generate"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type Optim struct {
	Queries *sqlc.Queries
	Redis   *redis.Client
	DB      *pgxpool.Pool

	PreOptim    *preoptim.PreOptim
	RPCPreOptim *grpc.Server
}

func NewOptim(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool) *Optim {
	return &Optim{
		Queries: queries,
		Redis:   redis,
		DB:      db,
	}
}

func (s *Optim) InitOptim() error {

	client, err := getS3Client()
	if err != nil {
		return err
	}

	s.PreOptim = preoptim.NewPreOptim(s.Queries, s.Redis, s.DB, client)
	err = s.PreOptim.InitPreOptim()
	if err != nil {
		s.PreOptim.StopPreOptim()
		return err
	}

	s.RPCPreOptim, err = rpcserver.InitPreOptimRPCServer(s.PreOptim)
	if err != nil {
		s.RPCPreOptim.GracefulStop()
		return err
	}

	return nil
}

func (s *Optim) StopOptim() error {

	s.RPCPreOptim.GracefulStop()
	s.PreOptim.StopPreOptim()

	return nil
}

func getS3Client() (*s3.Client, error) {

	region := os.Getenv("AWS_REGION")
	if region == "" {
		return nil, fmt.Errorf("aws region not found")
	}

	keyId := os.Getenv("AWS_ACCESS_KEY_ID")
	if keyId == "" {
		return nil, fmt.Errorf("aws key id not found")
	}

	key := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if key == "" {
		return nil, fmt.Errorf("aws key not found")
	}

	ctx := context.Background()

	config, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(keyId, key, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load a default config : %v", err)
	}

	client := s3.NewFromConfig(config, func(o *s3.Options) {
		o.ResponseChecksumValidation = aws.ResponseChecksumValidation(0)
	})

	return client, nil
}
