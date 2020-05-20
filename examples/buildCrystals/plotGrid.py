from matplotlib import pyplot as plt
import numpy as np
import sys

def main(fname):
    with open(fname, 'r') as infile:
        data = []
        for line in infile:
            if 'x' in line:
                continue
            splitted = line.split(',')
            row = [int(x) for x in splitted[:-1]]
            row.append(float(splitted[-1]))
            data.append(row)

    # Convert to a numpy array
    nx = max(row[0] for row in data) + 1
    ny = max(row[1] for row in data) + 1
    array = np.zeros((nx, ny))
    for row in data:
        array[row[0], row[1]] = row[2]

    fig = plt.figure()
    ax = fig.add_subplot(1, 1, 1)
    ax.imshow(array, cmap='bone')
    plt.show()

if __name__ == '__main__':
    main(sys.argv[1])