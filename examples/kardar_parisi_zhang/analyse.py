from matplotlib import rc
rc('font',**{'family':'sans-serif','sans-serif':['Helvetica'], 'size': 18})
rc('text', usetex=True)
from matplotlib import pyplot as plt
import numpy as np
from scipy.stats import linregress

def show(data):
    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)
    ax.imshow(data, cmap="nipy_spectral")
    return ax

def roughness(data):
    return np.std(data)

def main():
    rough = []
    for i in range(0, 200):
        data = np.fromfile("/work/sophus/kpz/run5/kpz_height_{}.bin".format(i), dtype=">f8")
        N = int(np.sqrt(len(data)))
        data = np.reshape(data, (N, N))
        rough.append(roughness(data))
    
    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)
    t = np.arange(1.0, len(rough)+1)
    start = 30
    end = 79
    slope, interscept, _, _, _ = linregress(np.log(t)[start:end], np.log(rough)[start:end])
    ax.plot(t, rough, 'o', mfc='none', markersize=3)
    ax.plot(t, np.exp(interscept)*t**slope)
    ax.set_xscale('log')
    ax.set_yscale('log')
    print("Slope: {}".format(slope))
    np.savetxt("roughness.txt", rough, header="roughness")
    plt.show()

def exponent_with_unc(data):
    num = 10000
    slopes = []
    t = np.arange(1.0, data.shape[0]+1)
    for i in range(num):
        idx = np.random.randint(0, high=data.shape[1], size=data.shape[1])
        logT = []
        logDat =[]
        for j in idx:
            logT += np.log(t).tolist()
            logDat += np.log(data[:, j]).tolist()
        
        slope, _, _, _, _ = linregress(logT, logDat)
        slopes.append(slope)
    return np.mean(slopes), np.std(slopes)


def average_roughness():
    fname = "/home/gudrun/davidkl/Nedlastinger/KPZRoughness.csv"
    data = np.loadtxt(fname, delimiter=',', skiprows=1)
    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)
    t = np.arange(1.0, data.shape[0]+1)
    logT = []
    logDat = []
    for i in range(data.shape[1]):
        ax.plot(t, data[:, i], 'o', mfc='none', color='#515156', alpha=0.05, markersize=2)
        logT += np.log(t).tolist()
        logDat += np.log(data[:, i]).tolist()
    ax.set_xscale('log', basex=2)
    ax.set_yscale('log', basey=2)
    slope, interscept, _, _, _ = linregress(logT, logDat)
    ax.plot(t, np.exp(interscept)*t**slope, color='#8d1e0b', lw=2)
    ax.set_xlabel("Time")
    ax.set_ylabel("Average width")
    ax.text(12.0, 0.010, '$\propto t^{0.235}$')
    ax.spines['right'].set_visible(False)
    ax.spines['top'].set_visible(False)
    print("Slope: {}".format(slope))

    slope_mean, slope_std = exponent_with_unc(data)
    print("Slope bootstrap: {} +- {}".format(slope_mean, slope_std))
    plt.show()

#main()
average_roughness()