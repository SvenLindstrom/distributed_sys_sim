import json
import pandas as pd
from collections import defaultdict
from pathlib import Path
import argparse

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
        task_id = log["taskID"]
        timestamp = to_timestamp(log["time"])
        message = log["msg"]

        if message == "submitted":
            if task_id not in tasks:
                tasks[task_id] = {"submitted": timestamp, "done": None}
        elif message == "done":
            if task_id in tasks:
                tasks[task_id]["done"] = timestamp

    return tasks

# From Scheduler
def get_scheduler_data(logs):
    assigned = defaultdict(list)
    completed = {}

    for log in logs:
        task_id = log.get("taskID")
        # Ignore non-task-related logs
        if not task_id:
            continue

        timestamp = to_timestamp(log["time"])
        message = log["msg"]

        if message == "assigned":
            assigned[task_id].append(timestamp)
        elif message == "completed":
            completed[task_id] = timestamp

    return assigned, completed

## METRICS

# Task Latency
def get_latency(data):
    latencies = []
    incomplete = []

    # Subtract Timestamps to get Timedeltas
    for task_id, timestamps in data.items():
        if timestamps["done"] is not None:
            diff = timestamps["done"] - timestamps["submitted"]
            latencies.append(diff)
        else:
            incomplete.append(task_id)

    latencies = pd.Series(latencies).sort_values()
    
    latency_result = {
        "completed_count": len(latencies),
        "incomplete_count": len(incomplete),
        "min": latencies.min().total_seconds(),
        "max": latencies.max().total_seconds(),
        "mean": latencies.mean().total_seconds(),
        "std": latencies.std().total_seconds(),
    }

    return latency_result

# Task Throughput
def get_throughput(completed):
    timestamps = sorted(completed.values())
    window = pd.Timedelta(seconds=1.0)

    # Create buckets to keep track of TPS
    t0 = timestamps[0]
    buckets = defaultdict(int)

    for ts in timestamps:
        bucket = int((ts - t0) / window)
        buckets[bucket] += 1
    
    duration = (timestamps[-1] - t0).total_seconds()
    tps = pd.Series(buckets.values())
    
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
def get_duplication(assigned):
    window = pd.Timedelta(seconds=60.0)
    start = min([ts_list[0] for ts_list in assigned.values()])

    # All tasks that got assigned in the first minute
    in_window = {}
    for task_id, ts_list in assigned.items():
        if ts_list[0] - start <= window:
            in_window[task_id] = ts_list
    
    total = len(in_window)
    duplicated_ids = [task_id for task_id, ts_list in in_window.items() if len(ts_list) > 1]
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
    gen_logs = get_logs(trial_path / "generator.log")
    sch_logs = get_logs(trial_path / "scheduler.log")

    # Data Extraction
    gen_data = get_generator_data(gen_logs)
    assigned, completed = get_scheduler_data(sch_logs)

    # Get Metrics
    latency = get_latency(gen_data)
    throughput = get_throughput(completed)
    duplication = get_duplication(assigned)

    trial_result = {
        "trial_num": trial_path.name.split('_')[1],
        "latency": latency,
        "throughput": throughput,
        "duplication": duplication
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
    
    summary = summarise_trials(trial_results)

    config_result = {
        "configuration": config_path.name,
        "summary": summary,
    }

    # Save Results per Configuration
    res_path = output_path / f"{config_path.name}_results.json"
    data = {
        "configuration": config_path.name,
        "trials": trial_results
    }
    save_to_file(data, res_path)

    return config_result

def summarise_trials(trial_results):
    latency_summary = summarise_metric([trial["latency"]["mean"] for trial in trial_results])
    throughput_summary = summarise_metric([trial["throughput"]["mean_tps"] for trial in trial_results])
    duplication_summary = summarise_metric([trial["duplication"]["dup_rate"] for trial in trial_results])

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
        "std":  values.std(),
        "min":  values.min(),
        "max":  values.max(),
    }

    return trial_summary

# Everything
def run_all(logs_dir, output_dir):
    logs_path = Path(logs_dir)
    output_path = Path(output_dir)
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
        help="Root log directory containing all configuration subdirectories"
    )

    # Output directory to save results
    parser.add_argument(
        "--output-dir",
        help="Destination directory to write JSON result files"
    )

    args = parser.parse_args()

    ## ANALYSIS
    run_all(args.input_dir, args.output_dir)

if __name__ == "__main__":
    main()