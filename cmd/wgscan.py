#!/usr/bin/env python
from multiprocessing import Pool
import pyarrow.parquet as pq
import pandas as pd
import json
import os
import subprocess
import sys


def to_array(file):
    tmp_config = {}
    tmp_config["file"] = os.path.abspath(TEMP_INPUT_FOLDER + "/" + file)
    tmp_config["output"] = os.path.abspath(OUTPUT_FOLDER + "/" + file + "_output.txt")
    data = pq.read_pandas(os.path.abspath(INPUT_FOLDER + "/" + file)).to_pandas()
    pd.np.savetxt(tmp_config["file"], data.T.values, delimiter=",", fmt="%s")
    tmp_config["length"] = data.values.shape[0]
    return tmp_config


INPUT_FOLDER = sys.argv[1]
OUTPUT_FOLDER = sys.argv[2]

TEMP_INPUT_FOLDER = "input/"

filenames = []
for _, _, _filenames in os.walk(INPUT_FOLDER):
    filenames.extend(_filenames)

if not os.path.exists(TEMP_INPUT_FOLDER):
    os.mkdir(TEMP_INPUT_FOLDER)

config = []

with Pool(processes=2) as pool:
    results = pool.map(to_array, filenames)
    pool.close()
[config.append(result) for result in results]

with open("config.json", "w") as f:
    json.dump(config, f)

subprocess.run("./wgscan")

for tmp_config in config:
    with open(tmp_config["output"], "r") as f:
        calc_scores = f.read().split(",")

    tmp = [i for i in range(2000, tmp_config["length"], 100)]
    tmp = tmp[:len(calc_scores)]
    pd.DataFrame({"score": calc_scores, "index": tmp}).to_csv(tmp_config["output"].split(".")[0] + ".csv", index=False)
    os.remove(tmp_config["output"])

# cleanup
filenames = []
for _, _, _filenames in os.walk(TEMP_INPUT_FOLDER):
    filenames.extend(_filenames)
for file in filenames:
    os.remove(os.path.abspath(TEMP_INPUT_FOLDER + "/" + file))
os.removedirs(TEMP_INPUT_FOLDER)
