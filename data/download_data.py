import geopandas as gpd
import pandas as pd
import requests
import tempfile
import zipfile
import os


# CREATE TEST FOLDERS
base_folder = "../data/tests"
os.makedirs(f"{base_folder}/ca", exist_ok=True)
os.makedirs(f"{base_folder}/de", exist_ok=True)
base_folder = "../data/expected_outputs"
os.makedirs(f"{base_folder}/ca", exist_ok=True)
os.makedirs(f"{base_folder}/de", exist_ok=True)


# GET ZCTA SHP FILES
# Download all ZCTAs (ZIP Code Tabulation Areas)
zcta_url = "https://www2.census.gov/geo/tiger/GENZ2020/shp/cb_2020_us_zcta520_500k.zip"
with tempfile.TemporaryDirectory() as tmpdir:
    zip_path = os.path.join(tmpdir, "zcta.zip")
    # get file with SSL verification disabled
    response = requests.get(zcta_url, verify=False)
    with open(zip_path, "wb") as f:
        f.write(response.content)
    with zipfile.ZipFile(zip_path, "r") as zip_ref:
        zip_ref.extractall(tmpdir)
    # find the .shp
    shp_files = [f for f in os.listdir(tmpdir) if f.endswith(".shp")]
    shapefile_path = os.path.join(tmpdir, shp_files[0])
    # load into gpd
    zcta_gdf = gpd.read_file(shapefile_path)
# Ref: https://apps.naaccr.org/vpr-cls/about-postal-codes
# Filter to California ZCTAs (ZIP codes 90001 to 96199, not including the stand alone ones): 1802 rows
gdf_zcta_ca = zcta_gdf[zcta_gdf["ZCTA5CE20"].str.startswith(tuple(str(i) for i in range(900, 962)))]
gdf_zcta_ca.to_file("../data/tests/ca/ca_zipcode.geojson", driver="GeoJSON") 
# Filter to Delaware ZCTAs (ZIP codes 19701 to 19999): 68 rows
gdf_zcta_de = zcta_gdf[zcta_gdf["ZCTA5CE20"].str.startswith(tuple(str(i) for i in range(197, 200)))]
gdf_zcta_de.to_file("../data/tests/de/de_zipcode.geojson", driver="GeoJSON") 


# GET CENSUS TRACT SHP FILES
## California: 9109 rows ##
tract_url = "https://www2.census.gov/geo/tiger/GENZ2020/shp/cb_2020_06_tract_500k.zip"
tract_temp = tempfile.NamedTemporaryFile(delete=False, suffix=".zip")
tract_temp.write(requests.get(tract_url, verify=False).content)
tract_temp.close()
with zipfile.ZipFile(tract_temp.name, 'r') as zip_ref:
    extract_dir = tempfile.mkdtemp()
    zip_ref.extractall(extract_dir)
gdf_tracts_ca = gpd.read_file(extract_dir)
gdf_tracts_ca["GEOID"] = gdf_tracts_ca["GEOID"].astype(str)
## Delaware : 259 rows ##
tract_url = "https://www2.census.gov/geo/tiger/GENZ2020/shp/cb_2020_10_tract_500k.zip"
tract_temp = tempfile.NamedTemporaryFile(delete=False, suffix=".zip")
tract_temp.write(requests.get(tract_url, verify=False).content)
tract_temp.close()
with zipfile.ZipFile(tract_temp.name, 'r') as zip_ref:
    extract_dir = tempfile.mkdtemp()
    zip_ref.extractall(extract_dir)
gdf_tracts_de = gpd.read_file(extract_dir)
gdf_tracts_de["GEOID"] = gdf_tracts_de["GEOID"].astype(str)


# GET POPULATION DATA
## California: 9129 rows ##
pop_api_url = "https://api.census.gov/data/2020/dec/pl?get=P1_001N,NAME&for=tract:*&in=state:06"
response = requests.get(pop_api_url, verify=False)
pop_data = response.json()
pop_df_ca = pd.DataFrame(pop_data[1:], columns=pop_data[0])
pop_df_ca["GEOID"] = pop_df_ca["state"] + pop_df_ca["county"] + pop_df_ca["tract"]
pop_df_ca["P1_001N"] = pop_df_ca["P1_001N"].astype(int)
#combine with tract shapefiles
gdf_tracts_ca = gdf_tracts_ca.merge(pop_df_ca[["GEOID", "P1_001N"]], on="GEOID", how="left")
gdf_tracts_ca.to_file("../data/tests/ca/ca_tracts.geojson", driver="GeoJSON") 
## Delaware: 262 rows ##
pop_api_url = "https://api.census.gov/data/2020/dec/pl?get=P1_001N,NAME&for=tract:*&in=state:10"
response = requests.get(pop_api_url, verify=False)
pop_data = response.json()
pop_df_de = pd.DataFrame(pop_data[1:], columns=pop_data[0])
pop_df_de["GEOID"] = pop_df_de["state"] + pop_df_de["county"] + pop_df_de["tract"]
pop_df_de["P1_001N"] = pop_df_de["P1_001N"].astype(int)
#combine with tract shapefiles
gdf_tracts_de = gdf_tracts_de.merge(pop_df_de[["GEOID", "P1_001N"]], on="GEOID", how="left")
gdf_tracts_de.to_file("../data/tests/de/de_tracts.geojson", driver="GeoJSON")  


# MAKE ANSWER KEY (csv)
## California##
gdf_tracts_ca = gdf_tracts_ca.to_crs(epsg=3857)
gdf_zcta_ca = gdf_zcta_ca.to_crs(epsg=3857)
gdf_tracts_ca['centroid'] = gdf_tracts_ca.geometry.centroid
centroid_tracts_ca = gdf_tracts_ca[['AFFGEOID', 'P1_001N', 'centroid']].copy()
centroid_tracts_ca = gpd.GeoDataFrame(centroid_tracts_ca, geometry='centroid')
# Perform spatial join using centroids
ca_joined = gpd.sjoin(centroid_tracts_ca, gdf_zcta_ca, how='right', predicate='within')
# Calculate the sum of the population for each ZIP code
sum_pop = ca_joined.groupby('ZCTA5CE20')['P1_001N'].sum().reset_index()
sum_pop.to_csv('../data/expected_outputs/ca/sum_pop_by_zip.csv', index=False)
## Delaware ##
gdf_tracts_de = gdf_tracts_de.to_crs(epsg=3857)
gdf_zcta_de = gdf_zcta_de.to_crs(epsg=3857)
gdf_tracts_de['centroid'] = gdf_tracts_de.geometry.centroid
centroid_tracts_de = gdf_tracts_de[['AFFGEOID', 'P1_001N', 'centroid']].copy()
centroid_tracts_de = gpd.GeoDataFrame(centroid_tracts_de, geometry='centroid')
# Perform spatial join using centroids
de_joined = gpd.sjoin(centroid_tracts_de, gdf_zcta_de, how='right', predicate='within')
# Calculate the sum of the population for each ZIP code
sum_pop = de_joined.groupby('ZCTA5CE20')['P1_001N'].sum().reset_index()
sum_pop.to_csv('../data/expected_outputs/de/sum_pop_by_zip.csv', index=False)
