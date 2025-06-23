# %%
import geopandas as gpd
import matplotlib.pyplot as plt

# get data
gdf_de_tracts = gpd.read_file("tests/de/de_tracts.geojson")
gdf_de_zipcodes = gpd.read_file("tests/de/de_zipcode.geojson")

# plot
fig, ax = plt.subplots(figsize=(10, 10))
gdf_de_zipcodes.plot(ax=ax, edgecolor="blue", facecolor="none", linewidth=1, label="Zipcodes")
gdf_de_tracts.plot(ax=ax, edgecolor="orange", facecolor="none", linewidth=1, label="Tracts")
legend_handles = [
    plt.Line2D([0], [0], color="blue", lw=2, label="Zipcodes"),
    plt.Line2D([0], [0], color="orange", lw=2, label="Tracts"),]
ax.legend(handles=legend_handles, loc="upper right")
ax.set_title("Overlayed Zipcodes and Tracts, Delaware")
plt.savefig("de_overlap.png", dpi=300)
plt.show()
