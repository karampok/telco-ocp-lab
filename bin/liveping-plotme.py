#!/usr/bin/env python3

import os
import pandas as pd
import matplotlib.pyplot as plt
from http.server import SimpleHTTPRequestHandler
from socketserver import TCPServer
import threading

def plot_csv(file_path, ax):
    df = pd.read_csv(file_path, comment='#')  # Ignore lines starting with '#'


    date_column = df.columns[0]
    value_column = df.columns[1]

    df[date_column] = pd.to_datetime(df[date_column])

    ax.plot(df[date_column], df[value_column], marker='o')
    ax.set_title(os.path.basename(file_path))
    ax.set_xlabel('time')
    ax.set_ylabel('on/off')
    ax.set_ylim(0, 2)
    ax.yaxis.set_major_locator(plt.MultipleLocator(1))
    ax.grid(True)

def process_csv_files(folder_path):
    csv_files = [f for f in os.listdir(folder_path) if f.endswith('.csv')]

    num_files = len(csv_files)
    num_rows = num_files // 2
    num_cols = 2

    fig, axes = plt.subplots(num_files, 1, figsize=(10, 5 * num_rows))
    axes = axes.flatten()

    for i, csv_file in enumerate(csv_files):
        file_path = os.path.join(folder_path, csv_file)
        plot_csv(file_path, axes[i])

    plt.tight_layout()
    image_path = os.path.join(folder_path, 'liveping.png')
    plt.savefig(image_path)
    #plt.show()
    import webbrowser
    webbrowser.open(image_path)

def get_folder_path():
    folder_path = os.environ.get('CSVPATH')
    if folder_path is None:
        folder_path = './data/'

    return folder_path


if __name__ == "__main__":
    csv_folder_path = get_folder_path()
    process_csv_files(csv_folder_path)
