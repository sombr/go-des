package main

import (
	"fmt"
	"math/rand"

	roundrobin "github.com/sombr/go-container-roundrobin"
)

type EventQueue[T any] interface {
	Push(element T) error
	Pop() (T, error)
	Peek() (T, error)
	Size() int
}

type Simulation struct {
	// min-queue for tracking repair times
	eventQueue EventQueue[int]
	// total volume we need to process
	passengerCount int
	// number of parallel processing gates
	gateCount int
	// probability of a gate breaking
	breakChance float32
	// units of time required for repairs
	repairTime int
	// processing time
	processingTime int
}

func (s *Simulation) run(seed int64) (percentileTime [3]int) {
	rgen := rand.New(rand.NewSource(seed))

	var time int
	var done int
	var broken int
	for done < s.passengerCount {
		// fix broken
		for s.eventQueue.Size() > 0 {
			ts, _ := s.eventQueue.Peek()
			if ts > time {
				break
			}
			s.eventQueue.Pop()
			broken--
		}

		// process through all healthy gates
		time += s.processingTime
		done += s.gateCount - broken

		// roll a chance to of breakage for all healthy gates
		for idx := 0; idx < s.gateCount-broken; idx++ {
			if rgen.Float32() < s.breakChance {
				broken++
				s.eventQueue.Push(time + s.repairTime)
			}
		}

		// PPF metrics (percentiles of done)
		percentDone := 100 * done / s.passengerCount
		percentiles := []int{50, 95, 99}
		for idx, p := range percentiles {
			if percentDone >= p && percentileTime[idx] == 0 {
				percentileTime[idx] = time
			}
		}
	}

	return
}

func main() {
	r := roundrobin.NewRingQueue[int](10)

	stepSim := Simulation{
		eventQueue:     r,
		passengerCount: 1_000_000,
		gateCount:      10,
		breakChance:    0.05,
		repairTime:     120,
		processingTime: 15,
	}

	res := stepSim.run(100)

	fmt.Printf("Time to X%%: 50%% = %d, 95%% = %d, 99%% = %d\n", res[0], res[1], res[2])
	fmt.Printf("Relative Time to X%%: 50%% = 1x, 95%% = %.2fx, 99%% = %.2fx\n", float32(res[1])/float32(res[0]), float32(res[2])/float32(res[0]))
}
