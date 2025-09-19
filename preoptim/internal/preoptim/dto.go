package preoptim

import "time"

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Response DTO

type JobInfo struct {
	UUID       string    `json:"uuid"`
	JobId      string    `json:"jobId"`
	JunctionId int64     `json:"junctionId"`
	CameraId   int64     `json:"cameraId"`
	TimeStamp  time.Time `json:"timestamp"`
}

type ClipMeta struct {
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Duration  int64     `json:"duration"`
	FrameRate float32   `json:"frameRate"`
	Height    int64     `json:"height"`
	Width     int64     `json:"width"`
}

type ClipLocation struct {
	URL         string `json:"url"`
	ValidFor    int64  `json:"validFor"`
	Format      string `json:"format"`
	Size        int64  `json:"size"`
	CheckSumUrl string `json:"checksumUrl"`
}

type SystemMeta struct {
	Producer      string `json:"producer"`
	PipelineStage string `json:"pipelineStage"`
	LastJobId     string `json:"lastJobId"`
}

type ClipResult struct {
	Info       JobInfo      `json:"info"`
	Meta       ClipMeta     `json:"meta"`
	Location   ClipLocation `json:"location"`
	SystemMeta SystemMeta   `json:"systemMeta"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>

type InternalMetrics struct {
	StartTime time.Time // entire process start time

	RecordStart time.Time // frame recording start time
	RecordEnd   time.Time // frame recording end time

	ClipStart time.Time // frames stitching start time
	ClipEnd   time.Time // frames stitching end time

	UploadStart time.Time // clip upload and presign start time
	UploadEnd   time.Time // clip upload and presign end time

	EndTime time.Time // entire process end time
}
