import matplotlib as mpl
mpl.rcParams.update({'font.family': 'serif', 'font.size': 11, 'svg.fonttype': 'none'})
from matplotlib import pyplot as plt
import numpy as np

N = 128
fname = "density.csv"

def main():
    data = np.loadtxt(fname, delimiter=",", skiprows=1)
    array = np.zeros((N, N))
    for i in range(data.shape[0]):
        array[int(data[i, 0]), int(data[i, 1])] = data[i, 3]

    fig = plt.figure(figsize=(4, 3))
    ax = fig.add_subplot(1, 1, 1)
    im = ax.imshow(array*100, origin="lower", cmap="bone")
    cbar = fig.colorbar(im)
    cbar.set_label("Charge density")
    fig.tight_layout()
    plot_current()

def plot_current():
    data = np.loadtxt("current.csv", delimiter=",", skiprows=1)
    currentX = data[:, 0].reshape((128, 128))
    currentY = data[:, 1].reshape((128, 128))

    fig = plt.figure(figsize=(4, 3))
    ax = fig.add_subplot(1, 1, 1)
    im = ax.imshow(np.sqrt(currentX**2 + currentY**2), origin="lower", cmap="bone")

    x = list(range(128))
    X, Y = np.meshgrid(x, x)
    step = 10
    ax.quiver(X[::step, ::step], Y[::step, ::step],
             currentX[::step, ::step], currentY[::step, ::step],
             scale_units='xy', scale=0.5, angles='xy')
    cbar = fig.colorbar(im)
    cbar.set_label("Current density")

    meanX = np.mean(currentX)
    meanY = np.mean(currentY)
    Ex = 1.0

    sigma_xx = meanX/Ex
    sigma_xy = meanY/Ex
    print(sigma_xx, sigma_xy)

    
main()
plt.show()