import matplotlib as mpl
import numpy as np
mpl.rcParams.update({'font.family': 'serif', 'font.size': 11, 'svg.fonttype': 'none'})
from matplotlib import pyplot as plt

FNAME = "timeseries.csv"
CONC_FNAME = "conc.csv"

def plot():
    fig = plt.figure(figsize=(4, 3))
    ax = fig.add_subplot(1, 1, 1)
    data = np.loadtxt(FNAME, delimiter=',', skiprows=1)
    ax.plot(data[:, 1], color='#1c1c14')
    ax.plot(data[:, 3], color='#742d18')
    ax.set_xlabel("Time")
    ax.set_ylabel("Mean concentration")
    ax2 = ax.twinx()
    ax2.plot(data[:, 0], ls='--', color='#1c1c14')
    ax2.plot(data[:, 2], ls='--', color='#742d18')
    ax2.set_ylabel("Std. concentration")
    fig.tight_layout()

def plot_conc():
    data = np.loadtxt(CONC_FNAME, delimiter=',', skiprows=1)
    nx = int(np.max(data[:, 0])) + 1
    ny = int(np.max(data[:, 1])) + 1
    array = np.zeros((nx ,ny))
    for i in range(data.shape[0]):
        array[int(data[i, 0]), int(data[i, 1])] = data[i, 3]

    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)
    ax.imshow(array, cmap="gray", origin="lower")
    circle1 = plt.Circle((54, 54), 5, fill=False, lw=2)
    circle2 = plt.Circle((72, 72), 5, fill=False, lw=2)
    ax.add_artist(circle1)
    ax.add_artist(circle2)
    ax.text(60, 48, "(54, 54)")
    ax.text(76, 76, "(72, 72)")

plot()
plot_conc()
plt.show()