import numpy as np
import matplotlib as mpl
mpl.rcParams.update({'font.family': 'serif', 'font.size': 11})
from matplotlib import pyplot as plt
import sys

def defect(x, y):
    alpha = 0.99
    return 1.0 - alpha + alpha*np.tanh((x + 0.2*np.cos(2.0*np.pi*y))/0.01)**4

def plot(fname):
    data = np.loadtxt(fname, delimiter=',', skiprows=1)
    N = int(np.max(data[:, 0]) + 1)
    array = np.zeros((N, N))
    for i in range(data.shape[0]):
        ix = int(data[i, 0])
        iy = int(data[i, 1])
        v = data[i, 3]
        array[ix, iy] = v

    fig = plt.figure(figsize=(4, 3))
    ax = fig.add_subplot(1, 1, 1)
    im = ax.imshow(array.T, origin='lower', cmap='magma')

    x = np.linspace(-0.5, 0.5, N)
    X, Y = np.meshgrid(x, x)
    Z = defect(X, Y)
    ax.imshow(Z, alpha=0.3, cmap='magma')
    cbar = fig.colorbar(im)
    cbar.set_label("Concentration")

    plt.show()

def plotForceTorque(fname):
    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)

    data = np.loadtxt(fname, delimiter=',', skiprows=1)
    force = data[:, 2]
    torque = data[:, 1]

    ax.plot(force)
    ax.set_yscale('log')
    ax2 = ax.twinx()
    ax2.plot(torque)
    ax2.set_yscale('log')
    plt.show()

if __name__ == '__main__':
    fname = sys.argv[1]
    #plot(fname)
    plotForceTorque(fname)