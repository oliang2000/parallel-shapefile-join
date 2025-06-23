#%%
import pandas as pd
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

# read in results
df = pd.read_csv("benchmark_results.csv")

# calculate speedups 
sequential_ca_time = df.loc[(df["function"] == "sequential") & (df["file_size"] == "big"), "time"].values[0]
sequential_de_time = df.loc[(df["function"] == "sequential") & (df["file_size"] == "small"), "time"].values[0]

# Filter and calculate speedups for California (big file)
ps_ca_df = df[(df["function"] == "ps") & (df["file_size"] == "big")].copy()
ps_ca_df["speedup"] = sequential_ca_time / ps_ca_df["time"]
pb_ca_df = df[(df["function"] == "pb") & (df["file_size"] == "big")].copy()
pb_ca_df["speedup"] = sequential_ca_time / pb_ca_df["time"]
# Filter and calculate speedups for Delaware (small file)
ps_de_df = df[(df["function"] == "ps") & (df["file_size"] == "small")].copy()
ps_de_df["speedup"] = sequential_de_time / ps_de_df["time"]
pb_de_df = df[(df["function"] == "pb") & (df["file_size"] == "small")].copy()
pb_de_df["speedup"] = sequential_de_time / pb_de_df["time"]

# plot
plt.figure(figsize=(10, 6))
plt.plot(ps_ca_df["nthreads"].values, ps_ca_df["speedup"].values, marker='o', linestyle='-', label="Parallel Work Stealing (big file)")
plt.plot(pb_ca_df["nthreads"].values, pb_ca_df["speedup"].values, marker='s', linestyle='--', label="Paralle Basic (big file)")
plt.plot(ps_de_df["nthreads"].values, ps_de_df["speedup"].values, marker='x', linestyle='-', label="Parallel Work Stealing Small (small file)")
plt.plot(pb_de_df["nthreads"].values, pb_de_df["speedup"].values, marker='^', linestyle='--', label="Paralle Basic Small (small file)")
plt.title("Speedup vs. Number of Threads")
plt.xlabel("Number of Threads")
plt.ylabel("Speedup")
plt.grid(True)
plt.legend()
plt.tight_layout()
plt.savefig("speedup_plot.png")

# %%
