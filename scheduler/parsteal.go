package scheduler

import (
	"fmt"
	"log"
	"sync"
	"sync/atomic"

	"github.com/paulmach/orb/planar"
)

// queue and worker implementations

const dequeSize = 8192

type Work struct {
	Tract CensusTract
}

type Deque struct {
	buffer [dequeSize]*Work
	head   int64
	tail   int64
}

func NewDeque() *Deque {
	return &Deque{}
}

func (d *Deque) Push(work *Work) {
	t := atomic.LoadInt64(&d.tail)
	h := atomic.LoadInt64(&d.head)
	if t-h >= dequeSize {
		panic("deque overflow")
	}
	d.buffer[t%dequeSize] = work
	atomic.StoreInt64(&d.tail, t+1)
}

func (d *Deque) Pop() *Work {
	t := atomic.AddInt64(&d.tail, -1) - 1
	h := atomic.LoadInt64(&d.head)
	if t < h {
		atomic.StoreInt64(&d.tail, h)
		return nil
	}
	return d.buffer[t%dequeSize]
}

func (d *Deque) Steal() *Work {
	for {
		h := atomic.LoadInt64(&d.head)
		t := atomic.LoadInt64(&d.tail)
		if h >= t {
			return nil
		}
		work := d.buffer[h%dequeSize]
		if atomic.CompareAndSwapInt64(&d.head, h, h+1) {
			return work
		}
	}
}

type ThiefWorker struct {
	ID           int
	Deque        *Deque
	Others       []*Deque
	ZipcodeAreas *[]ZipcodeArea
	Mu           *sync.Mutex
	Wg           *sync.WaitGroup
}

func (w *ThiefWorker) Start() {
	defer w.Wg.Done()
	for {
		work := w.Deque.Pop()
		if work == nil {
			for _, dq := range w.Others {
				if dq == w.Deque {
					continue
				}
				work = dq.Steal()
				if work != nil {
					break
				}
			}
			if work == nil {
				return
			}
		}
		w.process(work)
	}
}

func (w *ThiefWorker) process(work *Work) {
	for i := range *w.ZipcodeAreas {
		if planar.PolygonContains((*w.ZipcodeAreas)[i].Polygon, work.Tract.Centroid) {
			w.Mu.Lock()
			(*w.ZipcodeAreas)[i].PopSum += work.Tract.Population
			w.Mu.Unlock()
			break
		}
	}
}

// run!
func RunParallelSteal(config Config) {
	fmt.Println("Running parallel with work stealing ...")
	censusTracts, zipcodeAreas, outputPath := loadData("../data", config.DataDirs)

	//spawn workers
	nWorkers := config.ThreadCount
	deques := make([]*Deque, nWorkers)
	for i := range deques {
		deques[i] = NewDeque()
	}

	var wg sync.WaitGroup
	mu := &sync.Mutex{}

	workers := make([]*ThiefWorker, nWorkers)
	for i := 0; i < nWorkers; i++ {
		workers[i] = &ThiefWorker{
			ID:           i,
			Deque:        deques[i],
			Others:       deques,
			ZipcodeAreas: &zipcodeAreas,
			Mu:           mu,
			Wg:           &wg,
		}
		wg.Add(1)
		go workers[i].Start()
	}
	// producer
	for i, tract := range censusTracts {
		deques[i%nWorkers].Push(&Work{Tract: tract})
	}
	// for all work to be completed
	wg.Wait()
	//output
	if err := writeOutput(zipcodeAreas, outputPath); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
}
