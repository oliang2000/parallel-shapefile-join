package scheduler

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

// struct for both input types: census tracts and zipcode areas
type CensusTract struct {
	Centroid   orb.Point
	Population int
}

type ZipcodeArea struct {
	Polygon orb.Polygon
	ZIP     string
	PopSum  int
}

// Functions to load geometry with UnmarshalFeature collection
func loadCensusTracts(path string) []CensusTract {
	//if file doesnt exist
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read census geojson: %v", err)
	}
	//if cant read geojson file
	fc, err := geojson.UnmarshalFeatureCollection(file)
	if err != nil {
		log.Fatalf("Failed to parse census geojson: %v", err)
	}
	//parse!
	var tracts []CensusTract
	for _, f := range fc.Features {
		geom, _ := f.Geometry.(orb.Polygon)
		popAny := f.Properties["P1_001N"]
		popFloat, _ := popAny.(float64)
		centroid, _ := planar.CentroidArea(geom)
		tracts = append(tracts, CensusTract{
			Centroid:   centroid,
			Population: int(popFloat),
		})
	}
	return tracts
}

func loadZipcodeAreas(path string) []ZipcodeArea {
	//open file
	file, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Failed to read zipcode geojson: %v", err)
	}
	fc, err := geojson.UnmarshalFeatureCollection(file)
	if err != nil {
		log.Fatalf("Failed to parse zipcode geojson: %v", err)
	}
	//read
	var areas []ZipcodeArea
	for _, f := range fc.Features {
		switch geom := f.Geometry.(type) {
		case orb.Polygon:
			// Single Polygon: add directly
			areas = append(areas, ZipcodeArea{
				Polygon: geom,
				ZIP:     f.Properties["ZCTA5CE20"].(string),
				PopSum:  0,
			})
		case orb.MultiPolygon:
			// Merge MultiPolygon into a single unified Polygon (this happens when a zip code area have disconnected parts)
			mergedPolygon := orb.Polygon{}
			for _, poly := range geom {
				mergedPolygon = append(mergedPolygon, poly...)
			}
			areas = append(areas, ZipcodeArea{
				Polygon: mergedPolygon,
				ZIP:     f.Properties["ZCTA5CE20"].(string),
				PopSum:  0,
			})
		default:
			log.Fatalf("Invalid geometry at Zipcode %v", f.Properties["ZCTA5CE20"].(string))
		}
	}
	return areas
}

func writeOutput(zips []ZipcodeArea, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	defer writer.Flush()

	writer.Write([]string{"ZIPCode", "P1_001N"})
	for _, zip := range zips {
		writer.Write([]string{zip.ZIP, strconv.Itoa(zip.PopSum)})
	}
	return nil
}

func loadData(basePath string, dir string) ([]CensusTract, []ZipcodeArea, string) {
	censusPath := filepath.Join(basePath, "tests", dir, dir+"_tracts.geojson")
	zipcodePath := filepath.Join(basePath, "tests", dir, dir+"_zipcode.geojson")
	outputPath := filepath.Join(basePath, "output", dir, dir+".csv")
	censusTracts := loadCensusTracts(censusPath)
	zipcodeAreas := loadZipcodeAreas(zipcodePath)
	return censusTracts, zipcodeAreas, outputPath
}

// run
func RunSequential(config Config) {
	fmt.Println("Running sequential...")
	censusTracts, zipcodeAreas, outputPath := loadData("../data", config.DataDirs)
	// Aggregate census to ZIP polygons by centroids
	for _, tract := range censusTracts {
		for i := range zipcodeAreas {
			//go through all zipcode areas
			if planar.PolygonContains(zipcodeAreas[i].Polygon, tract.Centroid) {
				zipcodeAreas[i].PopSum += tract.Population
				break
			}
		}
	}
	//write results to csv
	if err := writeOutput(zipcodeAreas, outputPath); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
}

// References
// Geojson: https://pkg.go.dev/github.com/paulmach/go.geojson#section-readme
// orb: https://pkg.go.dev/github.com/paulmach/orb/geojson#section-readme
// planar: https://pkg.go.dev/github.com/paulmach/orb/planar#CentroidArea
