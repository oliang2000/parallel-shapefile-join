# Parallel Shapefile Join

This project implements a parallelized aggregation pipeline to join geospatial data from areal units that are not areal units are not nested or hierarchically related: specifically, aggregating U.S. Census tract-level population data to ZIP code areas using centroid-based spatial joins. The program is written in Go and demonstrates the use of parallel programming to accelerate the join operation, particularly for large shapefiles.

## Introduction

In geographic information science (GISc), different data sources often comes in incompatible spatial units. For instance, population is typically measured at the census tract level, while other data (e.g., counts of stray animals) may be collected at the ZIP code level. Since ZIP codes and census tracts are not hierarchically aligned, aggregating across these units requires spatial processing techniques.

A common method is to associate smaller unit data with larger units based on centroid containment. However, this process is computationally expensive, especially for large datasets. This project parallelizes that operation using Go, a work-stealing scheduler, and a MapReduce architecture.

## Methodology

Given:
- Census tracts with population values (GeoJSON)
- ZIP code area boundaries (GeoJSON)

The program:
1. Computes centroids of census tracts
2. Checks which ZIP area each centroid falls into
3. Aggregates population values to ZIP areas

The core algorithm uses:
- A **MapReduce** design: the "Map" step assigns tracts to ZIP codes, the "Reduce" step aggregates populations.
- **Parallelization** through multithreaded workers
- A **work-stealing** strategy to balance load dynamically

## Repository Structure

```text
.
├── aggregate/           # Command-line interface & execution logic
│   └── aggregate.go
├── benchmark/           # Benchmarking and visualization
│   ├── benchmark.sh
│   └── plot.py
├── data/                # Data and helper scripts
│   ├── tests/
│   │   ├── ca/          # California test case (large)
│   │   └── de/          # Delaware test case (small)
│   ├── expected_outputs/
│   ├── output/
│   ├── download_data.py
│   └── generate_map.py
├── scheduler/           # Core logic
│   ├── scheduler.go
│   ├── sequential.go
│   ├── parbasic.go
│   └── parsteal.go
```

## Results


