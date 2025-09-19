package preoptim

import (
	"context"
	"fmt"
	"strconv"
)

// CRUD

// Create
func (s *PreOptim) AddCamera(ctx context.Context, cameraid string) error {

	cameraID, err := strconv.ParseInt(cameraid, 10, 64)
	if err != nil {
		return err
	}

	cameraInfo, err := s.Queries.GetCameraInfo(ctx, cameraID)
	if err != nil {
		return err
	}

	newCtx, newCancel := context.WithCancel(s.Manager.rootCtx)
	clipPath := fmt.Sprintf("./frames/%d.mp4", cameraInfo.CameraID)

	worker := &FeedWorker{
		info: CameraInfo{
			cameraID:   cameraInfo.CameraID,
			junctionID: cameraInfo.JunctionID,
			rtspUrl:    cameraInfo.RtspUrl,
			angle:      cameraInfo.Angle,
			resolution: cameraInfo.Resolution,
			status:     cameraInfo.Status,
			createdAt:  cameraInfo.CreatedAt.Time,
			juncName:   cameraInfo.Juncname,
			lattitude:  cameraInfo.Latitude,
			longitude:  cameraInfo.Longitude,
		},

		ctx:    newCtx,
		ctxCnl: newCancel,
		work:   make(chan struct{}),

		config: CameraConfig{
			baseFolder: cfg.Worker.BaseFolder,

			captureFrameRate:    cfg.Worker.CaptureFrameRate,
			captureDuration:     cfg.Worker.CaptureDuration,
			captureFramesNaming: "",
			captureFrameExt:     cfg.Worker.CaptureFramesExt,

			clipFrameRate: cfg.Worker.ClipFrameRate,
			clipPath:      clipPath,

			uploadBucket:    cfg.Worker.ClipUploadBucket,
			preSignValidFor: cfg.Worker.ClipPreSignValidFor,

			redisPubChan: cfg.Worker.RedisPublishChan,
		},
	}

	s.Manager.mu.Lock()
	s.Manager.cameras = append(s.Manager.cameras, worker)
	s.Manager.mu.Unlock()

	s.Manager.wg.Add(1)

	go func(camera *FeedWorker) {
		defer s.Manager.wg.Done()

		err := s.processFeed(camera)
		if err != nil {
			newErr(err)
			return
		}

	}(worker)

	return nil
}

// Delete
func (s *PreOptim) DeleteCamera(ctx context.Context, cameraid string) error {

	return nil
}
