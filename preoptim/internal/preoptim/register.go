package preoptim

import (
	"context"
	sqlc "optim/internal/sqlc/generate"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type PreOptim struct {
	Queries      *sqlc.Queries
	Redis        *redis.Client
	DB           *pgxpool.Pool
	AWSClient    *s3.Client        // main aws base client
	AWSUploader  *manager.Uploader // main aws uploader client
	AWSPreSigner *s3.PresignClient // main aws presigner client, derieved from main base client

	Manager *FeedManager
}

func NewPreOptim(queries *sqlc.Queries, redis *redis.Client, db *pgxpool.Pool, awsClient *s3.Client) *PreOptim {
	return &PreOptim{
		Queries:      queries,
		Redis:        redis,
		DB:           db,
		AWSClient:    awsClient,
		AWSUploader:  manager.NewUploader(awsClient),
		AWSPreSigner: s3.NewPresignClient(awsClient),
	}
}

type CameraInfo struct {
	cameraID   int64
	junctionID int64
	rtspUrl    string
	angle      float64
	resolution string
	status     string
	createdAt  time.Time
	juncName   string
	lattitude  float64
	longitude  float64
}

type CameraConfig struct {
	baseFolder string // the base folder to save everything to

	captureFrameRate    float32 // the framerate of the capture
	captureDuration     int64   // seconds, the duration that the frames will be captured for
	captureFramesNaming string  // the frames naming format
	captureFrameExt     string  // the output extension of the captured frames

	clipFrameRate float32 // the framerate of the output clip
	clipPath      string

	uploadBucket    string // the bucket to upload the clip
	preSignValidFor int64  // seconds, the time for which the pre-signed aws url stays valid

	redisPubChan string // the channel to publish the result
}

type FeedWorker struct {
	info CameraInfo

	ctx    context.Context // context for each routine, derived from rootCtx
	ctxCnl context.CancelFunc
	work   chan struct{} // channel to signal the worker to start processing

	config CameraConfig
}

type FeedManager struct {
	cameras  []*FeedWorker // holds FeedWorkers'
	mu       sync.Mutex    // protects cameras
	Interval int           // milliseconds, the total time to distribute between all feeds
	RestTime int           // milliseconds, the rest time between two intervals, helps all routines finish before moving forward
	stagger  int           // milliseconds, the internal sleep time between two captures are started, offloads equally instead of a burst

	rootCtx    context.Context // the root context
	rootCtxCnl context.CancelFunc
	wg         sync.WaitGroup // used to manage all feed workers
	done       chan struct{}  // main chan to close anything and everything
}
