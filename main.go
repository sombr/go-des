package main

import (
	"fmt"
	"math/rand"

	heap "github.com/sombr/go-container-heap"
)

type EventQueue[T any] interface {
	Push(element T) error
	Pop() (T, error)
	Peek() (T, error)
	Size() int
}

type Event struct {
	time int
	kind byte
}

type Simulation struct {
	// min-queue for tracking repair times
	eventQueue EventQueue[Event]
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
	var inflight int
	var broken int
	for done < s.passengerCount {
		// roll broken dice
		for idx := 0; idx < s.gateCount && broken < s.gateCount; idx++ {
			if rgen.Float32() < s.breakChance {
				broken++
				repairTime := s.repairTime + int(rgen.NormFloat64()*float64(s.repairTime)/4)
				if repairTime < 0 {
					repairTime = 0
				}

				repairedAt := time + repairTime
				s.eventQueue.Push(Event{
					kind: 'R', // Repair
					time: repairedAt,
				})
			}
		}

		// top up inflight processing
		for inflight < s.gateCount-broken {
			inflight++
			processingTime := s.processingTime + int(rgen.NormFloat64()*float64(s.processingTime)/4)
			if processingTime < 0 {
				processingTime = 0
			}

			doneAt := time + processingTime
			s.eventQueue.Push(Event{
				kind: 'D',
				time: doneAt,
			})
		}

		nextEvent, _ := s.eventQueue.Pop()
		if nextEvent.kind == 'R' { // gate got repaired
			broken--
		} else { // passenger got processed
			done++
			inflight--
		}
		time = nextEvent.time

		// PPF metrics (percentiles of done)
		ppfIndex := 100 * done / s.passengerCount
		if percentileTime[ppfIndex] == 0 {
			percentileTime[ppfIndex] = time
		}
	}

	return
}

func main() {
	// we need 20 capacity for max 10 in flight and 10 repair events
	h := heap.NewHeap[Event](20, func(a, b *Event) bool { return b == nil || a != nil && a.time < b.time })

	stepSim := Simulation{
		eventQueue:     h,
		passengerCount: 1_000_000,
		gateCount:      10,
		breakChance:    0.05,
		repairTime:     120,
		processingTime: 15,
	}

	res := stepSim.run(100)

	fmt.Println(res)
}
