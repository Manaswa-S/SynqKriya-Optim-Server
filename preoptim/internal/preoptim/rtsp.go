package preoptim

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

func (s *PreOptim) InitPreOptim() error {

	ctx, cancel := context.WithCancel(context.Background())

	s.Manager = &FeedManager{
		cameras:    make([]*FeedWorker, 0),
		mu:         sync.Mutex{},
		Interval:   30000,
		RestTime:   15000,
		rootCtx:    ctx,
		rootCtxCnl: cancel,
		wg:         sync.WaitGroup{},
		done:       make(chan struct{}),
	}

	fmt.Printf("starting capture routines for all cameras ... ")
	err := s.initAllCameras()
	if err != nil {
		return err
	}
	fmt.Printf(" done \n")

	fmt.Printf("starting the capture scheduler for all cameras ... ")
	err = s.initScheduler()
	if err != nil {
		return err
	}
	fmt.Printf(" done \n")

	return nil
}

func (s *PreOptim) StopPreOptim() {

	fmt.Printf("stopping all capture routines ...")

	close(s.Manager.done)
	for _, camera := range s.Manager.cameras {
		camera.ctxCnl()
	}

	s.Manager.wg.Wait()

	fmt.Printf(" done \n")
}

type CameraStatus string

const (
	Active CameraStatus = "active"
)

func (s *PreOptim) initAllCameras() error {

	cameras, err := s.Queries.GetAllCameras(s.Manager.rootCtx)
	if err != nil {
		return err
	}

	for _, camera := range cameras {

		newCtx, newCancel := context.WithCancel(s.Manager.rootCtx)
		clipPath := fmt.Sprintf("./frames/%d.mp4", camera.CameraID)

		s.Manager.cameras = append(s.Manager.cameras, &FeedWorker{
			info: CameraInfo{
				cameraID:   camera.CameraID,
				junctionID: camera.JunctionID,
				rtspUrl:    camera.RtspUrl,
				angle:      camera.Angle,
				resolution: camera.Resolution,
				status:     camera.Status,
				createdAt:  camera.CreatedAt.Time,
				juncName:   camera.Juncname,
				lattitude:  camera.Latitude,
				longitude:  camera.Longitude,
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
		})
	}

	for _, camera := range s.Manager.cameras {
		s.Manager.wg.Add(1)
		go func(camera *FeedWorker) {
			defer s.Manager.wg.Done()

			err := s.processFeed(camera)
			if err != nil {
				newErr(err)
				return
			}

		}(camera)
	}

	return nil
}

func (s *PreOptim) initScheduler() error {

	s.Manager.wg.Add(1)

	go func() {
		defer s.Manager.wg.Done()

		for {
			select {
			case <-s.Manager.done:
				fmt.Println("shutting scheduler")
				return
			default:

				s.Manager.mu.Lock()
				camerasSnap := append([]*FeedWorker{}, s.Manager.cameras...)
				s.Manager.mu.Unlock()

				camsCnt := len(camerasSnap)
				if camsCnt == 0 {
					continue
				}

				s.Manager.stagger = (s.Manager.Interval / camsCnt)

				for _, cam := range camerasSnap {
					select {
					case <-s.Manager.done:
						fmt.Println("shutting scheduler")
						return
					case cam.work <- struct{}{}:
					default:
						fmt.Printf("dropping work ping for : %d\n", cam.info.cameraID)
					}

					time.Sleep(time.Duration(s.Manager.stagger) * time.Millisecond)
				}

				fmt.Println("scheduler resting...")
				time.Sleep(time.Duration(s.Manager.RestTime) * time.Millisecond)
			}
		}

	}()

	return nil
}

func (s *PreOptim) processFeed(camera *FeedWorker) error {

	camera.config.baseFolder = fmt.Sprintf("%s/%d", camera.config.baseFolder, camera.info.cameraID)
	camera.config.captureFramesNaming = camera.config.baseFolder + "/out_%03d" + camera.config.captureFrameExt

	defer func() {
		err := os.RemoveAll(camera.config.baseFolder)
		if err != nil {
			newErr(err)
			return
		}
	}()

	for {
		select {
		case <-s.Manager.done:
			return nil
		case <-camera.ctx.Done():
			return nil
		case <-camera.work:

			newLog(LogInfo, fmt.Sprintf("capturefeed : %d : %v : %s \n", camera.info.cameraID, time.Now(), camera.info.rtspUrl))

			metrics := &InternalMetrics{StartTime: time.Now()}
			err := s.manageProcess(camera, metrics)
			if err != nil {
				return err
			}
			metrics.EndTime = time.Now()

			newLog(LogMetric, fmt.Sprintf("donefeed : %d : start>%v : end>%v \n"+
				">>> recording : start> %v : end> %v \n"+
				">>> clipping : start> %v : end> %v \n"+
				">>> upload : start> %v : end> %v \n",
				camera.info.cameraID, metrics.StartTime, metrics.EndTime,
				metrics.RecordStart, metrics.RecordEnd,
				metrics.ClipStart, metrics.ClipEnd,
				metrics.UploadStart, metrics.UploadEnd,
			))
		}
	}
}

func (s *PreOptim) manageProcess(camera *FeedWorker, metrics *InternalMetrics) error {

	defer func() {
		err := os.RemoveAll(camera.config.baseFolder)
		if err != nil {
			fmt.Println(err)
			return
		}
	}()

	err := os.MkdirAll(camera.config.baseFolder, 0755)
	if err != nil {
		return err
	}

	metrics.RecordStart = time.Now()
	err = s.captureFrames(camera)
	if err != nil {
		return err
	}
	metrics.RecordEnd = time.Now()

	metrics.ClipStart = time.Now()
	err = s.stitchFrames(camera)
	if err != nil {
		return err
	}
	metrics.ClipEnd = time.Now()

	metrics.UploadStart = time.Now()
	presignUrl, err := s.uploadAndPreSign(camera)
	if err != nil {
		return err
	}
	metrics.UploadEnd = time.Now()

	frameHeight, frameWidth, err := s.getImgDimensions(camera)
	if err != nil {
		return err
	}

	err = s.publishResult(camera, metrics, &resultDetails{
		presignURL: presignUrl,
		height:     frameHeight,
		width:      frameWidth,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *PreOptim) captureFrames(camera *FeedWorker) error {

	cmd := exec.Command(
		"ffmpeg",
		"-rtsp_transport", "tcp",
		"-i", camera.info.rtspUrl,
		"-t", fmt.Sprintf("%d", camera.config.captureDuration),
		"-vf", "fps="+fmt.Sprintf("%f", camera.config.captureFrameRate),
		camera.config.captureFramesNaming,
	)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *PreOptim) stitchFrames(camera *FeedWorker) error {

	cmd := exec.Command(
		"ffmpeg",
		"-framerate", fmt.Sprintf("%f", camera.config.clipFrameRate),
		"-i", camera.config.captureFramesNaming,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "23",
		"-pix_fmt", "yuv420p",
		camera.config.clipPath,
	)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (s *PreOptim) uploadAndPreSign(camera *FeedWorker) (string, error) {

	clipFile, err := os.ReadFile(camera.config.clipPath)
	if err != nil {
		return "", err
	}

	fileKey := fmt.Sprintf("%d.%d.clip.mp4", camera.info.cameraID, time.Now().UTC().Unix())
	uploadRes, err := s.AWSUploader.Upload(camera.ctx, &s3.PutObjectInput{
		Bucket: aws.String(camera.config.uploadBucket),
		Key:    aws.String(fileKey),
		Body:   bytes.NewReader([]byte(clipFile)),
	})
	if err != nil {
		return "", err
	}

	req, err := s.AWSPreSigner.PresignGetObject(camera.ctx, &s3.GetObjectInput{
		Bucket: aws.String(camera.config.uploadBucket),
		Key:    uploadRes.Key,
	}, func(po *s3.PresignOptions) {
		po.Expires = time.Duration(camera.config.preSignValidFor) * time.Second
	})
	if err != nil {
		return "", err
	}

	return req.URL, nil
}

func (s *PreOptim) getImgDimensions(camera *FeedWorker) (height, width int64, err error) {

	files, err := os.ReadDir(camera.config.baseFolder)
	if err != nil {
		return 0, 0, err
	}

	path := ""

	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "out_") {
			path = camera.config.baseFolder + "/" + file.Name()
			break
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	details, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return int64(details.Height), int64(details.Width), nil
}

type resultDetails struct {
	presignURL string
	height     int64
	width      int64
}

func (s *PreOptim) publishResult(camera *FeedWorker, metrics *InternalMetrics, details *resultDetails) error {
	uuid, err := uuid.NewV7()
	if err != nil {
		return err
	}

	result := &ClipResult{
		Info: JobInfo{
			UUID:       uuid.String(),
			JobId:      "job-",
			JunctionId: camera.info.junctionID,
			CameraId:   camera.info.cameraID,
			TimeStamp:  time.Now(),
		},
		Meta: ClipMeta{
			Start:     metrics.RecordStart,
			End:       metrics.RecordEnd,
			Duration:  camera.config.captureDuration,
			FrameRate: camera.config.captureFrameRate,
			Height:    details.height,
			Width:     details.width,
		},
		Location: ClipLocation{
			URL:         details.presignURL,
			ValidFor:    camera.config.preSignValidFor,
			Format:      camera.config.captureFrameExt,
			Size:        -1,
			CheckSumUrl: "na",
		},
		SystemMeta: SystemMeta{
			Producer:      "preoptim",
			PipelineStage: "preprocessed",
			LastJobId:     "na",
		},
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return err
	}

	_, err = s.Redis.Publish(camera.ctx, camera.config.redisPubChan, resultBytes).Result()
	if err != nil {
		return err
	}

	return nil
}
