package midoptim

import "time"

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Common DTO

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

type SystemMeta struct {
	Producer      string `json:"producer"`
	PipelineStage string `json:"pipelineStage"`
	LastJobId     string `json:"lastJobId"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Request DTO

type FrameSummary struct {
	FrameID     int64            `json:"frame_id"`
	ClassCounts map[string]int64 `json:"class_counts"`
	AvgSpeed    float32          `json:"avg_speed_px_per_sec"`
}

type VehicleStats struct {
	TrackID     int64   `json:"track_id"`
	ArrivalTime float32 `json:"arrival_time"`
	ExitTime    float32 `json:"exit_time"`
	AvgSpeed    float32 `json:"avg_speed_px_per_sec"`
}

type ProcessedSummary struct {
	FrameSummaries []FrameSummary `json:"frame_summaries"`
	VehicleStats   []VehicleStats `json:"vehicle_stats"`
	TotalVehicles  int64          `json:"total_vehicles"`
}

type YoloDSReq struct {
	Info             JobInfo          `json:"info"`
	Meta             ClipMeta         `json:"meta"`
	ProcessedSummary ProcessedSummary `json:"processed_summary"`
	SystemMeta       SystemMeta       `json:"systemMeta"`
}

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// Response DTO

type ClipMetrics struct {
	VehiclesCount map[string]int64 `json:"vehicleCounts"`
	Occupancy     float32          `json:"occupancy"`
	FlowRate      int64            `json:"flowRate"`
	AvgSpeed      float32          `json:"avgSpeed"`
	AvgWaitTime   int64            `json:"avgWaitTime"`
	AvgHeadway    int64            `json:"avgHeadway"`
	QueueLength   int64            `json:"queueLength"`
	Anomalies     map[string]int64 `json:"anomalies"`
	SpillOver     bool             `json:"spillOver"`
	RoadType      string           `json:"roadType"`
	Confidence    float32          `json:"confidence"`
}

type MetricsResult struct {
	Info       JobInfo     `json:"info"`
	Meta       ClipMeta    `json:"meta"`
	Metrics    ClipMetrics `json:"metrics"`
	SystemMeta SystemMeta  `json:"systemMeta"`
}
