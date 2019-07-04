#!/usr/bin/env python

import pandas as pd
import json
import os
import subprocess

INPUT_FOLDER = "input/"
OUTPUT_FOLDER = "output/"

CMD_FOLDER = "../cmd/"
TEMP_INPUT_FOLDER = CMD_FOLDER + "input/"

filenames = []
for _, _, _filenames in os.walk(INPUT_FOLDER):
    filenames.extend(_filenames)

if not os.path.exists(TEMP_INPUT_FOLDER):
    os.mkdir(TEMP_INPUT_FOLDER)

config = []
for file in filenames:
    tmp_config = {}
    tmp_config["file"] = os.path.abspath(TEMP_INPUT_FOLDER + "/" + file)
    tmp_config["output"] = os.path.abspath(OUTPUT_FOLDER + "/" + file.split(".")[0] + "_output.txt")
    data = pd.read_csv(os.path.abspath(INPUT_FOLDER + "/" + file), sep="\t")
    data.columns = ["chr_no", "index", "coverage"]
    data.drop(["chr_no", "index"], inplace=True, axis=1)
    pd.np.savetxt(tmp_config["file"], data.T.values, delimiter=",", fmt="%s")
    tmp_config["length"] = data.values.shape[0]
    config.append(tmp_config)

with open(os.path.abspath(CMD_FOLDER+"config.json"), "w") as f:
    json.dump(config, f)

subprocess.run(os.path.abspath(CMD_FOLDER+"./wgscan"), cwd=CMD_FOLDER)

with open("actual_output.txt", "r") as f:
    actual_scores = f.read().split(",")

for tmp_config in config:
    with open(tmp_config["output"], "r") as f:
        calc_scores = f.read().split(",")

    for i, j in zip(actual_scores, calc_scores):
        if (float(i)-float(j))/float(i) > 0.1:
            assert(False)
