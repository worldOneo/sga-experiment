import sys
import pandas as pd
import numpy as np
from matplotlib import pyplot as plt

ms_ns = 1_000_000


def pt(y, q):
    return f"{round(y.quantile(q)/ms_ns, 2)}ms"


for name in sys.argv[1:]:
    plt.rcParams["figure.figsize"] = [7.00, 3.50]
    plt.rcParams["figure.autolayout"] = True
    df = pd.read_csv(f"{name}", delimiter=",")
    fd = df[df.columns[2]]
    y = df[df.columns[1]]
    x = df[df.columns[0]].values

    plt.plot(x, y.values, lw=0.4)
    plt.plot(x, y.rolling(100, min_periods=1).mean(), lw=0.4)
    plt.plot(x, y.rolling(1_000, min_periods=1).mean(), lw=0.4)
    plt.plot(x, y.rolling(10_000, min_periods=1).mean(), lw=0.4)
    plt.figtext(
        0, 0, f"10%={pt(y,.1)} 50%={pt(y,.5)} 90%={pt(y,.9)} 99%={pt(y,.99)} 100%={pt(y,1)} F/S={round(fd[fd==True].count()/fd.count(), 2)}%")
    plt.ylabel("Time [ns]")
    plt.savefig(f"{name}.png", dpi=200, bbox_inches="tight")
    plt.clf()
