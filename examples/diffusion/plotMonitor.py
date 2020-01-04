import json
from matplotlib import pyplot as plt

def main():
    fname = "diffusionMonitor.json"
    with open(fname, 'r') as infile:
        data = json.load(infile)

    conc = [1.0] + data[0]["Data"]
    time = list(range(0, len(conc)))

    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)
    ax.plot(time, conc, marker="o")
    ax.set_xlabel("Concentration")
    ax.set_ylabel("Time")
    ax.spines["right"].set_visible(False)
    ax.spines["top"].set_visible(False)
    ax.set_yscale("log")
    plt.show()

main()
    