package postoptim

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

func (s *PostOptim) InitPostOptim() error {

	s.Manager = new(Manager)

	s.Manager.ctx, s.Manager.ctxCnl = context.WithCancel(context.Background())

	s.Manager.pullPolicyChan = make(chan *MLReq, 100) // TODO: capacity of 100 ???
	s.Manager.wg = sync.WaitGroup{}
	s.Manager.done = make(chan struct{})

	s.Manager.redisChanGetInp = os.Getenv("RedisPostOptimInpChan")
	if s.Manager.redisChanGetInp == "" {
		return fmt.Errorf("post optim: redis inp chan not set in .env")
	}

	s.Manager.redisChanSetOut = os.Getenv("RedisPostOptimOutChan")
	if s.Manager.redisChanSetOut == "" {
		return fmt.Errorf("post optim: redis out chan not set in .env")
	}

	err := s.initRedisPull()
	if err != nil {
		return err
	}

	err = s.initPolicyApplier()
	if err != nil {
		return err
	}

	return nil
}

func (s *PostOptim) StopPostOptim() {
	close(s.Manager.done)

	s.Manager.wg.Wait()
}

func (s *PostOptim) initRedisPull() error {

	s.Manager.wg.Add(1)

	go func() {
		defer s.Manager.wg.Done()

		redisInp := s.Redis.Subscribe(s.Manager.ctx, s.Manager.redisChanGetInp)
		defer redisInp.Close()

		inpChan := redisInp.Channel()

		for {
			select {
			case <-s.Manager.done:
				fmt.Println("closing post optim redis pull")
				return
			case msg := <-inpChan:

				var data MLReq
				err := json.Unmarshal([]byte(msg.Payload), &data)
				if err != nil {
					fmt.Println(err)
					return
				}

				s.Manager.pullPolicyChan <- &data
			}
		}

	}()

	return nil
}

func (s *PostOptim) initPolicyApplier() error {

	s.Manager.wg.Add(1)

	go func() {
		defer s.Manager.wg.Done()

		for {
			select {
			case <-s.Manager.done:
				fmt.Println("closing policy applier")
				return
			case data := <-s.Manager.pullPolicyChan:

				err := s.applyPolicies(data)
				if err != nil {
					fmt.Println(err)
					return
				}

				dataBytes, err := json.Marshal(data)
				if err != nil {
					fmt.Println(err)
					return
				}

				err = s.Redis.Publish(s.Manager.ctx, s.Manager.redisChanSetOut, dataBytes).Err()
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}

	}()

	return nil
}

func (s *PostOptim) applyPolicies(data *MLReq) error {

	for _, update := range data.Updates {

		policies, err := s.Queries.GetPoliciesForCameraID(s.Manager.ctx, update.CameraID)
		if err != nil {
			return err
		}

		if update.Plan.GreenTime < int64(policies.GreenMin) {
			update.Plan.GreenTime = int64(policies.GreenMin)
		}
		if update.Plan.GreenTime > int64(policies.GreenMax) {
			update.Plan.GreenTime = int64(policies.GreenMax)
		}

		if update.Plan.RedTime < int64(policies.RedMin) {
			update.Plan.RedTime = int64(policies.RedMin)
		}
		if update.Plan.RedTime > int64(policies.RedMax) {
			update.Plan.RedTime = int64(policies.RedMax)
		}

	}

	return nil
}
