import numpy as np
import matplotlib as mpl
mpl.rcParams.update({'font.size': 14, 'svg.fonttype': 'none'})
from matplotlib import pyplot as plt
import json

DATA_FILE = "phaseData.csv"

def get_phase_data(phase):
    result = {
        'energy': [],
        'density': [],
        'temperature': [],
    }
    with open(DATA_FILE, 'r') as f:
        for line in f.readlines()[1:]:
            splitted = line.split(',')
            phase_on_line = splitted[-1].strip()
            if phase_on_line == phase:
                result['temperature'].append(float(splitted[0]))
                result['density'].append(float(splitted[1]))
                result['energy'].append(float(splitted[2]))
    return result

def plot_rough_phase_diagram():
    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)

    markers = {'liquid': 'o', 'square': 's', 'triangle': 'v'}
    for phase in ['liquid', 'square', 'triangle']:
        data = get_phase_data(phase)
        ax.plot(data['density'], data['temperature'], markers[phase], color="#AAAAAA")
    ax.set_xlabel("Density")
    ax.set_ylabel("Effective temperature")
    ax.spines['right'].set_visible(False)
    ax.spines['top'].set_visible(False)
    return fig

def plot_pair_correlation():
    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)
    f = np.linspace(0.0, 2.0, 500)
    colors = ['#1c1c14', '#742d18', '#816545']
    for i, t in enumerate([0.0, 0.5, 1.0]):
        p1 = np.exp(-t**2/4.0)
        f1 = np.exp(-(f-1)**2/(2.0*0.004))

        p2 = np.exp(-np.sqrt(2.0)*t**2/4.0)
        f2 = np.exp(-(f-np.sqrt(2.0))**2/(2.0*0.004))
        ax.plot(f, f1*p1, color=colors[i], label=f"\u03c3 = {t}")
        ax.plot(f, f2*p2, color=colors[i])
    ax.set_xlabel("Spatial frequency")
    ax.set_ylabel("Interaction strength")
    ax.spines['right'].set_visible(False)
    ax.spines['top'].set_visible(False)
    ax.legend(loc='best', frameon=False)

def main():
    #plot_rough_phase_diagram()
    plot_pair_correlation()
    plt.show()

if __name__ == '__main__':
    main()