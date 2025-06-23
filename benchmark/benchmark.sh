#!/bin/bash
#
#SBATCH --mail-user=oliang@uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj3_benchmark
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/oliang/project-3-oliang2000/proj3/benchmark
#SBATCH --partition=debug 
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=4:00:00

module load golang/1.19

#!/bin/bash

### DATA DOWNLOAD ###
python ../data/download_data.py

### TEST ###
OUTPUT_FILE="benchmark_results.csv"
THREAD_COUNTS=(2 4 6 8 10 12)
AGGREGATE_PATH="../aggregate/aggregate.go"

echo "function,nthreads,time,file_size" > $OUTPUT_FILE

# Sequential benchmark for Delaware (small file)
total=0
for i in {1..5}; do
    output=$(go run "$AGGREGATE_PATH" "de")
    time_line=$(echo "$output" | grep -A1 "Running sequential..." | tail -n1)
    total=$(echo "$total + $time_line" | bc)
done
avg=$(echo "scale=2; $total / 5" | bc)
echo "sequential,1,$avg,small" >> $OUTPUT_FILE

# Parallel (basic) benchmarks for Delaware (small file)
for n in "${THREAD_COUNTS[@]}"; do
    total=0
    for i in {1..5}; do
        output=$(go run "$AGGREGATE_PATH" "de" pb $n)
        time_line=$(echo "$output" | grep -A1 "Running basic parallel..." | tail -n1)
        total=$(echo "$total + $time_line" | bc)
    done
    avg=$(echo "scale=2; $total / 5" | bc)
    echo "pb,$n,$avg,small" >> $OUTPUT_FILE
done

# Parallel (work stealing) benchmarks for Delaware (small file)
for n in "${THREAD_COUNTS[@]}"; do
    total=0
    for i in {1..5}; do
        output=$(go run "$AGGREGATE_PATH" "de" ps $n)
        time_line=$(echo "$output" | grep -A1 "Running parallel with work stealing ..." | tail -n1)
        total=$(echo "$total + $time_line" | bc)
    done
    avg=$(echo "scale=2; $total / 5" | bc)
    echo "ps,$n,$avg,small" >> $OUTPUT_FILE
done


# Sequential benchmark for California (big file)
total=0
for i in {1..5}; do
    output=$(go run "$AGGREGATE_PATH" "ca")
    time_line=$(echo "$output" | grep -A1 "Running sequential..." | tail -n1)
    total=$(echo "$total + $time_line" | bc)
done
avg=$(echo "scale=2; $total / 5" | bc)
echo "sequential,1,$avg,big" >> $OUTPUT_FILE

# Parallel (basic) benchmarks for California (big file)
for n in "${THREAD_COUNTS[@]}"; do
    total=0
    for i in {1..5}; do
        output=$(go run "$AGGREGATE_PATH" "ca" pb $n)
        time_line=$(echo "$output" | grep -A1 "Running basic parallel..." | tail -n1)
        total=$(echo "$total + $time_line" | bc)
    done
    avg=$(echo "scale=2; $total / 5" | bc)
    echo "pb,$n,$avg,big" >> $OUTPUT_FILE
done

# Parallel (work stealing) benchmarks for California (big file)
for n in "${THREAD_COUNTS[@]}"; do
    total=0
    for i in {1..5}; do
        output=$(go run "$AGGREGATE_PATH" "ca" ps $n)
        time_line=$(echo "$output" | grep -A1 "Running parallel with work stealing ..." | tail -n1)
        total=$(echo "$total + $time_line" | bc)
    done
    avg=$(echo "scale=2; $total / 5" | bc)
    echo "ps,$n,$avg,big" >> $OUTPUT_FILE
done

# Plot speedup
python plot.py