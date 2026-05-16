#!/bin/bash

set -e

TRIALS=1
TRIAL_DURATION=50
LOGDIR="logs"
RESULTS_DIR="results"
WORKERS=10
FOLLOWERS=2

variants=(
	"BASE"
	"FAILOVER"
	"ELECTION-REPLICATION"
	)

runTrial(){
	VARIANT=$1
	LOGPATH=$2

	SETUP=$VARIANT LOGPATH=$LOGPATH docker compose up -d --scale worker=${WORKERS} --scale follower=${FOLLOWERS}

	echo "Trial will run for $TRIAL_DURATION seconds"
	sleep $TRIAL_DURATION
	echo "Trial completed"
	docker compose down
}

runVariant(){
	VARIANT=$1
	for i in $(seq 1 $TRIALS); do
		echo "Running $VARIANT Trial #$i"
		LOGPATH="$LOGDIR/$VARIANT/trial_$i"
		mkdir -p "$LOGPATH"
		runTrial "$VARIANT" "$LOGPATH/"
	done
}

for i in "${variants[@]}"; do
	runVariant "$i"
done

## Data Analysis

# create venv
if [ ! -d "venv" ]; then
	python -m venv venv
fi

# install modules
venv/bin/pip install -r requirements.txt

# run script
echo "Initialising Data Analysis of generated logs..."
venv/bin/python data_analysis.py --input-dir $LOGDIR/ --output-dir $RESULTS_DIR/
