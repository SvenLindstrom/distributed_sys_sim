import json
import pandas as pd

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
def get_scheduler_data():
    pass

## METRICS

# Task Latency
def get_latency():
    pass

# Task Throughput
def get_throughput():
    pass

# Task Duplication
def get_duplication():
    pass

## ANALYSIS LEVELS

# Per Trial
def run_trial():
    pass

# Per Configuration
def run_configuration():
    pass

# Everything
def run_all():
    pass

## MAIN

def main():
    gen_logs = get_logs('logs/generator.log')
    gen_data = get_generator_data(gen_logs)

    print(gen_data)

if __name__ == "__main__":
    main()