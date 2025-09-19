package postoptim

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

type MLNetworkSummary struct {
	TotalJunctions   int64    `json:"total_junctions"`
	UpdatedJunctions int64    `json:"updated_junctions"`
	AvgCongestionLvl float32  `json:"avg_congestion_level"`
	CriticalEvents   []string `json:"critical_events"`
}

type MLUpdatePlan struct {
	CycleLength int64 `json:"cycle_length"`
	GreenTime   int64 `json:"green_time"`
	YellowTime  int64 `json:"yellow_time"`
	RedTime     int64 `json:"red_time"`
}

type MLUpdateAdjustmentsDynamicRule struct {
	Rule   string `json:"rule"`
	Params any    `json:"params"`
}

type MLUpdateAdjustments struct {
	PriorityOverride   bool                             `json:"priority_override"`
	OffsetFromBaseline int64                            `json:"offset_from_baseline"`
	DynamicRules       []MLUpdateAdjustmentsDynamicRule `json:"dynamic_rules"`
}

type MLUpdate struct {
	CameraID    int64               `json:"cameraId"`
	Plan        MLUpdatePlan        `json:"plan"`
	Status      string              `json:"status"`
	Confidence  float32             `json:"confidence"`
	Reasoning   string              `json:"reasoning"`
	Adjustments MLUpdateAdjustments `json:"adjustments"`
	ValidUntil  time.Time           `json:"valid_until"`
}

type MLReq struct {
	Info           JobInfo          `json:"info"`
	Meta           ClipMeta         `json:"meta"`
	NetworkSummary MLNetworkSummary `json:"network_summary"`
	Updates        []MLUpdate       `json:"updates"`
	SystemMeta     SystemMeta       `json:"systemMeta"`
}
