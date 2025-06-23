package scheduler

import (
	"fmt"
	"log"
	"sync"

	"github.com/paulmach/orb/planar"
)

type Worker struct {
	ID           int
	CensusTracts []CensusTract
	Populations  map[string][]int
	Wg           *sync.WaitGroup
}

func (w *Worker) Start(zipcodeareas *[]ZipcodeArea) {
	defer w.Wg.Done()
	for _, tract := range w.CensusTracts {
		for _, zipcodearea := range *zipcodeareas {
			if planar.PolygonContains(zipcodearea.Polygon, tract.Centroid) {
				zipCode := zipcodearea.ZIP
				w.Populations[zipCode] = append(w.Populations[zipCode], tract.Population)
				break
			}
		}
	}
}

func RunParallelBasic(config Config) {
	fmt.Println("Running basic parallel...")
	//load data
	censusTracts, zipcodeAreas, outputPath := loadData("../data", config.DataDirs)
	//spawn workers
	nWorkers := config.ThreadCount
	workers := make([]*Worker, nWorkers)
	var wg sync.WaitGroup
	// map
	for i := 0; i < nWorkers; i++ {
		start := i * len(censusTracts) / nWorkers
		end := (i + 1) * len(censusTracts) / nWorkers
		if i == nWorkers-1 {
			end = len(censusTracts)
		}
		workers[i] = &Worker{
			ID:           i,
			CensusTracts: censusTracts[start:end],
			Populations:  make(map[string][]int),
			Wg:           &wg,
		}
		wg.Add(1)
		go workers[i].Start(&zipcodeAreas)
	}
	// barrier!!!
	wg.Wait()
	// reduce!!!
	finalPopulation := make(map[string]int)
	for _, worker := range workers {
		for zip, populations := range worker.Populations {
			for _, population := range populations {
				finalPopulation[zip] += population
			}
		}
	}
	//output
	for i := range zipcodeAreas {
		zip := zipcodeAreas[i].ZIP
		if totalPop, exists := finalPopulation[zip]; exists {
			zipcodeAreas[i].PopSum = totalPop
		}
	}
	if err := writeOutput(zipcodeAreas, outputPath); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
}
