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

func (s *Simulation) run(seed int64) (percentileTime [101]int) {
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
		nextTime := time + s.processingTime
		if s.eventQueue.Size() > 0 {
			nextTime, _ = s.eventQueue.Peek()
		}

		done += (s.gateCount - broken) * (nextTime - time) / s.processingTime
		time = nextTime

		// roll a chance to of breakage for all healthy gates
		for idx := 0; idx < s.gateCount-broken; idx++ {
			if rgen.Float32() < s.breakChance {
				broken++
				s.eventQueue.Push(time + s.repairTime)
			}
		}

		// PPF metrics (percentiles of done)
		ppfIndex := 100 * done / s.passengerCount
		if percentileTime[ppfIndex] == 0 {
			percentileTime[ppfIndex] = time
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

	fmt.Println(res)
}
