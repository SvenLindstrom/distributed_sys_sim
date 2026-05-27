import json
from os import times
import pandas as pd
from collections import defaultdict
from pathlib import Path
import argparse
from pprint import pprint
import outputdata

## UTILITIES


# Get logs
def get_logs(filename):
    logs = []

    with open(filename, "r") as file_content:
        for log in file_content:
            log = log.rstrip()

            if log:
                json_log = json.loads(log)
                logs.append(json_log)

    return logs


# Convert date string to pandas.Timestamp
def to_timestamp(date_str):
    return pd.Timestamp(date_str)


## DATA EXTRACTION


# From Task Generator
def get_generator_data(logs):
    tasks = {}

    for log in logs:
        tris = log["retryCount"]
        task_id = log["taskID"]
        timestamp = to_timestamp(log["time"])
        message = log["msg"]

        if tris == -1:
            if task_id + "0" in tasks:
                task_id = log["taskID"] + "1"
            else:
                task_id = log["taskID"] + "0"
        else:
            task_id = log["taskID"] + str(tris)

        if message == "submitted" or message == "re-submitted":
            if task_id not in tasks:
                tasks[task_id] = {"submitted": timestamp, "done": None}
            else:
                tasks[task_id]["submitted"] = timestamp
        # elif message == "re-submitted":
        #     if task_id not in tasks:
        #         tasks[task_id] = {"submitted": timestamp, "done": None}
        #     # tasks[task_id]["submitted"] = timestamp
        elif message == "done":
            if task_id in tasks:
                tasks[task_id]["done"] = timestamp
            else:
                tasks[task_id] = {"submitted": None, "done": timestamp}

    for key, val in tasks.copy().items():
        if val["submitted"] is None:
            print(key)
            del tasks[key]

    return tasks


# From Scheduler
def get_scheduler_data(logs):
    completed = {}
    crash = []

    for log in logs:
        task_id = log.get("taskID")
        log_type = log.get("type")
        timestamp = to_timestamp(log["time"])

        if log_type == "fault injector":
            crash.append(timestamp)

        # Ignore non-task-related logs
        if not task_id:
            continue

        message = log["msg"]

        if message == "completed":
            completed[task_id] = timestamp

    return completed, tuple(crash)


# From Worker
def get_worker_data(logs):
    completed = defaultdict(list)

    for log in logs:
        task_id = log.get("task")
        # Ignore non-task-related logs
        if not task_id:
            continue

        timestamp = to_timestamp(log["time"])
        completed[task_id].append(timestamp)

    return completed


## METRICS


# Task Latency
def get_latency(data, crash):
    latencies = []
    incomplete = []
    start, end = crash
    lps = defaultdict(list)

    timeSires_latency = []
    # Subtract Timestamps to get Timedeltas
    for task_id, timestamps in data.items():
        if timestamps["done"] is not None:
            diff = timestamps["done"] - timestamps["submitted"]
            latencies.append(diff)

            # Buckets for completed tasks' latencies per second
            since_crash = timestamps["done"] - start
            bucket = int((since_crash) / pd.Timedelta(seconds=1.0))
            lps[bucket].append(diff)
            timeSires_latency.append((since_crash, diff))

        else:
            incomplete.append(task_id)

    latencies = pd.Series(latencies).sort_values()
    outputdata.trialData(timeSires_latency)
    # Get mean latency per second for graphing
    for second, td_list in lps.copy().items():
        mean_latency = sum([td.total_seconds() for td in td_list]) / len(td_list)
        lps[second] = mean_latency

    latency_result = {
        "completed_count": len(latencies),
        "incomplete_count": len(incomplete),
        "min": latencies.min().total_seconds(),
        "max": latencies.max().total_seconds(),
        "mean": latencies.mean().total_seconds(),
        "std": latencies.std().total_seconds(),
        "latency_per_second": lps,
    }

    return latency_result


# Task Throughput
def get_throughput(data, crash):
    timestamps = sorted(data.values())
    window = pd.Timedelta(seconds=1.0)
    start, end = crash

    # Create buckets to keep track of TPS
    buckets = defaultdict(int)

    for ts in timestamps:
        bucket = int((ts - start) / window)
        buckets[bucket] += 1

    duration = (timestamps[-1] - t0).total_seconds()
    tps = pd.Series(buckets.values())

    outputdata.trialThoughput(tps.tolist())

    if len(timestamps) == sum(tps):
        total = sum(tps)

    throughput_result = {
        "total_completed": total,
        "window": window.total_seconds(),
        "tasks_per_second": tps.tolist(),
        "mean_tps": tps.mean(),
        "total_duration": duration,
        "peak_tps": max(tps),
        "std": tps.std(),
    }

    return throughput_result


# Task Duplication
def get_duplication(data, crash):
    window = pd.Timedelta(seconds=30.0)
    start = crash[0]
    lower = crash[1] - crash[0]
    # start = min([ts_list[0] for ts_list in data.values()])
    # start, end = crash

    # All tasks that got assigned in the first minute
    in_window = {}
    for task_id, ts_list in data.items():
        ts = ts_list[0] - start
        if ts <= window and ts >= lower:
            in_window[task_id] = ts_list

    total = len(in_window)

    duplicated_ids = [
        task_id for task_id, ts_list in in_window.items() if len(ts_list) > 1
    ]

    duplicated_tasks = [
        ts_list for task_id, ts_list in in_window.items() if len(ts_list) > 1
    ]

    outputdata.dupData(duplicated_tasks, crash)

    duplicated = len(duplicated_ids)
    extra_assignments = sum(len(in_window[task_id]) - 1 for task_id in duplicated_ids)
    rate = (duplicated / total * 100) if total > 0 else 0.0

    duplication_result = {
        "window": window.total_seconds(),
        "total_tasks_in_window": total,
        "duplicated_tasks": duplicated,
        "extra_assignments": extra_assignments,
        "dup_rate": round(rate, 3),
        "duplicated_task_ids": duplicated_ids,
    }

    return duplication_result


## ANALYSIS LEVELS


# Per Trial
def run_trial(trial_dir):
    trial_path = Path(trial_dir)

    # Get logs
    generator_logs = get_logs(trial_path / "generator.log")
    scheduler_logs = get_logs(trial_path / "scheduler.log")
    worker_logs = get_logs(trial_path / "worker.log")

    # Data Extraction
    generator_data = get_generator_data(generator_logs)
    scheduler_data, crash = get_scheduler_data(scheduler_logs)
    worker_data = get_worker_data(worker_logs)

    # Get Metrics
    latency = get_latency(generator_data, crash)
    throughput = get_throughput(scheduler_data, crash)
    duplication = get_duplication(worker_data, crash)

    trial_result = {
        "trial_num": trial_path.name.split("_")[1],
        "latency": latency,
        "throughput": throughput,
        "duplication": duplication,
    }

    return trial_result


# Per Configuration
def run_configuration(config_dir, output_path):
    config_path = Path(config_dir)

    # get trial dirs
    trial_dirs = sorted([dir for dir in config_path.iterdir()])
    trial_results = []

    for trial_dir in trial_dirs:
        print(f"Analysing Trial {trial_dir.name[-1]} out of {len(trial_dirs)}")
        result = run_trial(trial_dir)
        trial_results.append(result)

    (output_path / "raw").mkdir(parents=True, exist_ok=True)

    outputdata.writeToCSV(output_path / "raw", f"{config_path.name}.csv")

    summary = summarise_trials(trial_results)

    config_result = {
        "configuration": config_path.name,
        "summary": summary,
    }

    # Save Results per Configuration
    res_path = output_path / f"{config_path.name}_results.json"
    data = {"configuration": config_path.name, "trials": trial_results}
    save_to_file(data, res_path)

    return config_result


def summarise_trials(trial_results):
    latency_summary = summarise_metric(
        [trial["latency"]["mean"] for trial in trial_results]
    )
    throughput_summary = summarise_metric(
        [trial["throughput"]["mean_tps"] for trial in trial_results]
    )
    duplication_summary = summarise_metric(
        [trial["duplication"]["dup_rate"] for trial in trial_results]
    )

    full_summary = {
        "trial_count": len(trial_results),
        "latency": latency_summary,
        "throughput": throughput_summary,
        "duplication": duplication_summary,
    }

    return full_summary


def summarise_metric(metric_values):
    values = pd.Series(metric_values)

    trial_summary = {
        "mean": values.mean(),
        "std": values.std(),
        "min": values.min(),
        "max": values.max(),
    }

    return trial_summary


# Everything
def run_all(logs_dir, output_dir):
    logs_path = logs_dir
    output_path = output_dir
    output_path.mkdir(parents=True, exist_ok=True)

    # get config dirs
    config_dirs = sorted([dir for dir in logs_path.iterdir()])
    config_results = {}

    for config_dir in config_dirs:
        print(f"Starting {config_dir.name} Configuration Analysis...")
        result = run_configuration(config_dir, output_path)
        print(f"{config_dir.name} Configuration Analysis Completed!")
        config_results[config_dir.name] = result

    print("Full Analysis Completed")

    # save general results
    res_path = output_path / "final_results.json"
    save_to_file(config_results, res_path)
    print(f"Saved All Analysis Results to JSON files in {output_path}/")

    return config_results


def run_confs(logs_dir, output_dir):
    logs_path = Path(logs_dir)
    output_path = Path(output_dir)
    output_path.mkdir(parents=True, exist_ok=True)

    # get config dirs
    config_dirs = sorted([dir for dir in logs_path.iterdir()])

    for config_dir in config_dirs:
        conf_out = output_path / f"{config_dir.name}"
        run_all(config_dir, conf_out)


## SAVE RESULTS


def save_to_file(data, path):
    with open(path, "w") as file:
        json.dump(data, file, indent=4)


## MAIN


def main():
    ## CLA
    parser = argparse.ArgumentParser()

    # Input directory to analyse
    parser.add_argument(
        "--input-dir",
        help="Root log directory containing all configuration subdirectories",
    )

    # Output directory to save results
    parser.add_argument(
        "--output-dir", help="Destination directory to write JSON result files"
    )

    args = parser.parse_args()

    ## ANALYSIS
    # run_all(args.input_dir, args.output_dir)
    run_confs(args.input_dir, args.output_dir)


if __name__ == "__main__":
    main()
