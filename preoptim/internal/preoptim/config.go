package preoptim

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var cfg Config

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	cfg = InitConfigs()
}

func InitConfigs() Config {

	clipUpBuc, exists := os.LookupEnv("CLIP_UPLOAD_BUCKET")
	if !exists {
		log.Fatalln("CLIP_UPLOAD_BUCKET not found in env vars")
	}

	redisPubChan, exists := os.LookupEnv("REDIS_PUBLISH_CHAN")
	if !exists {
		log.Fatalln("REDIS_PUBLISH_CHAN not found in env vars")
	}

	return Config{
		Manager: ConfigFeedManager{
			Interval: 30000,
			RestTime: 15000,
		},
		Worker: ConfigFeedWorker{
			BaseFolder: "./frames",

			CaptureFrameRate: 1.00,
			CaptureDuration:  30,
			CaptureFramesExt: ".jpg",

			ClipFrameRate: 12,

			ClipUploadBucket:    clipUpBuc,
			ClipPreSignValidFor: 900,

			RedisPublishChan: redisPubChan,
		},
	}
}

type ConfigFeedManager struct {
	Interval int64
	RestTime int64
}

type ConfigFeedWorker struct {
	BaseFolder string

	CaptureFrameRate float32
	CaptureDuration  int64
	CaptureFramesExt string

	ClipFrameRate float32

	ClipUploadBucket    string
	ClipPreSignValidFor int64

	RedisPublishChan string
}

type Config struct {
	Manager ConfigFeedManager
	Worker  ConfigFeedWorker
}
