import json

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

## DATA EXTRACTION

# From Task Generator
def get_generator_data():
    pass

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
    print(gen_logs)

if __name__ == "__main__":
    main()