import json
import csv
from pathlib import Path
import matplotlib.pyplot as plt
import pandas as pd
import statsmodels.api as sm
from statsmodels.formula.api import ols
from statsmodels.stats.multicomp import pairwise_tukeyhsd
import seaborn as sns

trials_latency = []
trials_throughput = []
trials_dup = []

files = ["BASE.csv", "FAILOVER.csv", "ELECTION-REPLICATION.csv"]


def trialData(data):
    global trials_latency
    trials_latency += data


def dupData(data, crash):
    global trials_dup
    data_t = []
    for t in data:
        last_finish = t[-1]
        stamp = last_finish - crash[0]
        td = pd.to_timedelta(stamp)

        data_t.append(td.total_seconds())

    trials_dup.append(data_t)


def trialThoughput(data):
    global trials_throughput
    data_withT = []
    for i in range(len(data)):
        data_withT.append((i, data[i]))
    trials_throughput += data_withT


def writeDup(filepath, file_name):
    path = filepath / "duplication"
    path.mkdir(parents=True, exist_ok=True)
    filepath = path / file_name
    global trials_dup

    with open(filepath, "w") as f:
        csv_file = csv.writer(f, delimiter=",")
        for trial in range(len(trials_dup)):
            for t in trials_dup[trial]:
                data_row = []
                data_row.append(trial)
                data_row.append(t)
                csv_file.writerow(data_row)

    trials_dup = []


def writeThoughput(filepath, file_name):
    path = filepath / "throughput"
    path.mkdir(parents=True, exist_ok=True)
    filepath = path / file_name

    global trials_throughput

    tps = {}
    agre_data = []
    for thoughput in trials_throughput:
        vals = tps.get(thoughput[0], [])
        vals.append(thoughput[1])
        tps[thoughput[0]] = vals

    for key, val in tps.items():
        ave = float(sum(val)) / len(val)
        agre_data.append((key, ave))

    trials_throughput.sort(key=lambda x: x[0])

    with open(filepath, "w") as f:
        csv_file = csv.writer(f, delimiter=",")
        for data in agre_data:
            csv_file.writerow([data[0], data[1]])

    trials_throughput = []


def writeLatency(filepath, file_name):
    global trials_latency
    path = filepath / "latency"
    path.mkdir(parents=True, exist_ok=True)

    filepath = path / file_name

    trials_latency.sort(key=lambda x: x[0])

    with open(filepath, "w") as f:
        csv_file = csv.writer(f, delimiter=",")
        for lat in trials_latency:
            td = pd.to_timedelta(lat)
            data = td.total_seconds()
            csv_file.writerow([data[0], data[1]])

    trials_latency = []


def writeToCSV(filepath, file_name=""):
    writeLatency(filepath, file_name)
    writeThoughput(filepath, file_name)
    writeDup(filepath, file_name)


def statsMult(df):
    df["config"] = df["config"].astype("category")
    df["load"] = df["load"].astype("category")

    model = ols("latency ~ C(config) * C(load)", data=df).fit()

    anova_table = sm.stats.anova_lm(model, typ=2)
    print("\n=== TWO-WAY ANOVA ===")
    print(anova_table)

    anova_table["eta_sq"] = anova_table["sum_sq"] / anova_table["sum_sq"].sum()

    print("\n=== ANOVA WITH EFFECT SIZE ===")
    print(anova_table)

    print("\n=== TUKEY HSD PER LOAD ===")

    tukeys = {}

    for l in df["load"].cat.categories:
        subset = df[df["load"] == l]

        tukey = pairwise_tukeyhsd(
            endog=subset["latency"], groups=subset["config"], alpha=0.05
        )

        print(f"\n--- Load = {l} ---")
        print(tukey)

        tukey_df = pd.DataFrame(
            tukey.summary().data[1:], columns=tukey.summary().data[0]
        )

        tukeys[l] = tukey_df.to_dict()

    res = {"anova": anova_table, "tukey": tukeys}
    return res


def plotRow(axes, df, label, row_num, rows=2, typ="Latency", unit=""):
    sns.pointplot(
        data=df,
        x="load",
        y="latency",
        hue="config",
        errorbar="sd",
        order=[50, 100, 150],
        ax=axes[0],
    )
    if row_num == 0:
        axes[0].set_title(f"{typ} vs Load")
    axes[0].set_xlabel("Load (tasks/sec)")
    axes[0].set_ylabel(f"{typ} {unit}")
    axes[0].get_legend().remove()
    axes[0].grid(True)
    if rows > 1:
        axes[0].text(
            -0.1,
            1.1,
            label,
            transform=axes[0].transAxes,
            fontsize=14,
            fontweight="bold",
        )
    sns.stripplot(
        data=df,
        x="load",
        y="latency",
        hue="config",
        dodge=True,
        alpha=0.3,
        order=[50, 100, 150],
        ax=axes[1],
    )
    # if row_num == 0:
    axes[1].set_title(f"Raw {typ} vs Load")
    axes[1].set_xlabel("Load (tasks/sec)")
    axes[1].set_ylabel("")
    axes[1].get_legend().remove()

    axes[1].grid(True)

    if rows > 1:
        axes[1].text(
            -0.1,
            1.1,
            label,
            transform=axes[1].transAxes,
            fontsize=14,
            fontweight="bold",
        )


def makePlt(df, plot_name="interaction_plot.png", rows=2, typ="Latency", unit=""):
    plt.figure(figsize=(8, 5))
    fig, axes = plt.subplots(rows, 2, figsize=(15, 5), sharex=True)

    if rows == 1:
        axes = [axes]

    plotRow(axes[0], df, "A", 0, rows=rows, typ=typ, unit=unit)

    if rows > 1:

        df = df[(df["T"] < 20)]

        plotRow(axes[1], df, "B", 1, rows=rows, typ=typ, unit=unit)

    handles, labels = axes[0][0].get_legend_handles_labels()

    fig.legend(
        handles,
        labels,
        title="Configuration",
        loc="upper center",
        ncol=2,
        bbox_to_anchor=(0.5, 1.1),
    )
    plt.savefig(plot_name, dpi=300, bbox_inches="tight")


def pathFormat(file):
    return f"results/raw/{file}"


def read_data(typ="latency"):
    logs_path = Path("results")
    config_dirs = sorted([dir for dir in logs_path.iterdir()])
    config_dirs.remove(Path("results/stats"))
    data = []
    for config_dir in config_dirs:
        var_dir = Path(config_dir) / "raw" / typ
        for file in files:
            file_path = var_dir / file
            var_name = file.split(".")[0]
            with open(file_path, "r") as f:
                csv_file = csv.reader(f, delimiter=",")
                for row in csv_file:
                    tps = config_dir.name.split("_")[1]
                    data_row = []
                    data_row.append(row[0])
                    data_row.append(row[1])
                    data_row.append(var_name)
                    data_row.append(tps)
                    data.append(data_row)
    return data


def read_data_json():
    logs_path = Path("results")
    config_dirs = sorted([dir for dir in logs_path.iterdir()])
    config_dirs.remove(Path("results/stats"))
    data = []
    for config_dir in config_dirs:
        tps = config_dir.name.split("_")[1]
        var_dir = Path(config_dir)
        for file in files:
            var_name = file.split(".")[0]
            file_path = var_dir / f"{var_name}_results.json"
            with open(file_path, "r") as f:
                conf = json.load(f)
                for trial in conf["trials"]:
                    dup = trial["duplication"]
                    data_row = []
                    data_row.append(dup["dup_rate"])
                    data_row.append(var_name)
                    data_row.append(tps)
                    data.append(data_row)
    return data


def format_data(output_path, headers=["T", "latency", "config", "load"], typ="latency"):
    if typ == "duplication":
        data = read_data_json()
        headers = headers[1:]
    else:
        data = read_data(typ)

    data.sort(key=lambda x: x[0])
    data.insert(0, headers)
    output = output_path / f"comp_{typ}.csv"
    with open(output, "w") as f:
        csv_file = csv.writer(f, delimiter=",")
        csv_file.writerows(data)


def data_filter(df, typ):
    if typ == "duplication":
        df = df[df["config"] != "BASE"]
    else:
        df = df[(df["T"] > 2)]

    return df


def makeGrath(outpath, typ):
    grath_conf = {
        "latency": {"unit": " ms", "rows": 2},
        "throughput": {"unit": " task/sec", "rows": 1},
        "duplication": {"unit": " rate", "rows": 1},
    }

    input_path = outpath / f"comp_{typ}.csv"
    output_path = outpath / f"interaction_plot_{typ}"
    title = typ.capitalize()
    rows = grath_conf[typ]["rows"]
    unit = grath_conf[typ]["unit"]

    df = pd.read_csv(input_path)

    df = data_filter(df, typ)

    makePlt(
        df,
        plot_name=output_path,
        rows=rows,
        typ=title,
        unit=unit,
    )


def makeStats(outpath, typ):
    input_path = outpath / f"comp_{typ}.csv"
    df = pd.read_csv(input_path)
    df = data_filter(df, typ)
    stats = {}
    stats["full"] = statsMult(df)

    if typ == "latency":
        df = df[df["T"] < 20]
        stats["pre"] = statsMult(df)

    output_path = outpath / f"stats_{typ}.json"
    with open(output_path, "w") as f:
        stats_df = pd.DataFrame(data=stats)
        stats_df.to_json(f, orient="records", indent=4)


def main():

    path = Path("results/stats")

    (path).mkdir(parents=True, exist_ok=True)

    types = ["latency", "throughput", "duplication"]
    for t in types:
        print(f"doing: {t}")
        outpath = path / t
        (outpath).mkdir(parents=True, exist_ok=True)
        print("formating data")
        format_data(outpath, typ=t)
        print("making grath")
        makeGrath(outpath, t)
        print("calculating stats")
        makeStats(outpath, t)


if __name__ == "__main__":
    main()
