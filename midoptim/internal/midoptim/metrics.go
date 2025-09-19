package midoptim

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

func (s *MidOptim) InitMidOptim() error {

	s.Manager = new(Manager)

	s.Manager.ctx, s.Manager.ctxCnl = context.WithCancel(context.Background())

	s.Manager.pullCalcChan = make(chan *YoloDSReq, 100) // TODO: capacity of 100 ??
	s.Manager.wg = sync.WaitGroup{}
	s.Manager.done = make(chan struct{})

	s.Manager.occupancyMap = make(map[int64]int64)

	s.Manager.redisChanGetInp = os.Getenv("RedisMidOptimInpChan")
	if s.Manager.redisChanGetInp == "" {
		return fmt.Errorf("mid optim: redis inp chan not set in .env")
	}

	s.Manager.redisChanSetOut = os.Getenv("RedisMidOptimOutChan")
	if s.Manager.redisChanSetOut == "" {
		return fmt.Errorf("mid optim: redis out chan not set in .env")
	}

	err := s.initRedisPull()
	if err != nil {
		return err
	}

	err = s.initCalc()
	if err != nil {
		return err
	}

	return nil
}

func (s *MidOptim) StopMidOptim() {

	close(s.Manager.done)

	s.Manager.wg.Wait()
}

func (s *MidOptim) initRedisPull() error {

	s.Manager.wg.Add(1)

	go func() {
		defer s.Manager.wg.Done()

		redisInp := s.Redis.Subscribe(s.Manager.ctx, s.Manager.redisChanGetInp)
		defer redisInp.Close()

		inpChan := redisInp.Channel()

		for {
			select {
			case <-s.Manager.done:
				fmt.Println("closing mid optim redis pull")
				return
			case msg := <-inpChan:

				var data YoloDSReq
				err := json.Unmarshal([]byte(msg.Payload), &data)
				if err != nil {
					fmt.Println(err)
					return
				}

				s.Manager.pullCalcChan <- &data
			}
		}
	}()

	return nil
}

func (s *MidOptim) initCalc() error {

	s.Manager.wg.Add(1)

	go func() {
		defer s.Manager.wg.Done()

		for {
			select {
			case <-s.Manager.done:
				fmt.Println("closing mid optim metrics calculator")
				return
			case data := <-s.Manager.pullCalcChan:

				err := s.processMetrics(data)
				if err != nil {
					newErr(err)
					continue
				}

			}
		}
	}()

	return nil
}

func (s *MidOptim) processMetrics(data *YoloDSReq) error {

	metrics, err := s.calcMetrics(data)
	if err != nil {
		return err
	}

	err = s.publishMetrics(metrics)
	if err != nil {
		return err
	}

	return nil
}

func (s *MidOptim) calcMetrics(data *YoloDSReq) (*MetricsResult, error) {

	// classMap := make(map[string]int64)
	// for _, summary := range data.ProcessedSummary.FrameSummaries {
	// 	for class, count := range summary.ClassCounts {
	// 		classMap[class] = max(classMap[class], count)
	// 	}
	// }

	// s.Manager.occupancyMap[data.CameraId] = max(data.ProcessedSummary.TotalVehicles, s.Manager.occupancyMap[data.CameraId])
	// var occupancy float32
	// if s.Manager.occupancyMap[data.CameraId] == 0 {
	// 	occupancy = float32(data.ProcessedSummary.TotalVehicles)
	// } else {
	// 	occupancy = float32(data.ProcessedSummary.TotalVehicles) / float32(s.Manager.occupancyMap[data.CameraId])
	// }

	// flowRate := data.ProcessedSummary.TotalVehicles / data.Frame.Duration

	// speedSum := float32(0)
	// for _, summary := range data.ProcessedSummary.FrameSummaries {
	// 	speedSum += summary.AvgSpeed
	// }
	// avgSpeed := speedSum / float32(data.ProcessedSummary.TotalVehicles)

	// totalWait := float32(0)
	// for _, stat := range data.ProcessedSummary.VehicleStats {
	// 	totalWait += (stat.ExitTime - stat.ArrivalTime)
	// }
	// avgWaitTime := totalWait / float32(data.ProcessedSummary.TotalVehicles)

	// return &MidOptimOutput{
	// 	UUID:       data.UUID,
	// 	JobId:      data.JobId,
	// 	JunctionId: data.JunctionId,
	// 	CameraId:   data.CameraId,
	// 	TimeStamp:  data.TimeStamp,

	// 	Frame: data.Frame,

	// 	Metrics: Metrics{
	// 		VehiclesCount: classMap,
	// 		Occupancy:     occupancy,
	// 		FlowRate:      flowRate,
	// 		AvgSpeed:      avgSpeed,
	// 		AvgWaitTime:   int64(avgWaitTime),
	// 	},
	// 	RoadType:   "",
	// 	Confidence: 0.0,

	// 	SystemMeta: data.SystemMeta,
	// }, nil

	return &MetricsResult{}, nil
}

func (s *MidOptim) publishMetrics(metrics *MetricsResult) error {

	dataBytes, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	err = s.Redis.Publish(s.Manager.ctx, s.Manager.redisChanSetOut, dataBytes).Err()
	if err != nil {
		return err
	}

	return nil
}
