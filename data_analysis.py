import json

# read log file and return a list of logs
def get_logs(filename):
    logs = []

    with open(filename, "r") as file_content:
        for log in file_content:
            log = log.rstrip()

            if log:
                json_log = json.loads(log)
                logs.append(json_log)
    
    return logs

def main():
    gen_logs = get_logs('logs/generator.log')
    print(gen_logs)

if __name__ == "__main__":
    main()