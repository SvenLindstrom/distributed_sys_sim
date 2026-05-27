#!/bin/bash

set -e
source .env

LOGDIR=$LOGPATH

variants=(
	"BASE"
	"FAILOVER"
	"ELECTION-REPLICATION"
	)
tps=(
	"50"
	"100"
	"150"
)

runTrial(){
	VARIANT=$1
	LOGPATH=$2
	CONF=$3

	SETUP=$VARIANT LOGPATH=$LOGPATH  TASK_RATE=$CONF docker compose up -d --scale worker=${WORKERS} --scale follower=${FOLLOWERS}

	echo "Trial will run for $TRIAL_DURATION seconds"
	sleep $TRIAL_DURATION
	echo "Trial completed"
	docker compose down
}

runVariant(){
	CONF=$1
	VARIANT=$2
	for i in $(seq 1 $TRIALS); do
		echo "Running $VARIANT Trial #$i"
		LOGPATH="$LOGDIR/tps_$CONF/$VARIANT/trial_$i"
		mkdir -p "$LOGPATH"
		runTrial "$VARIANT" "$LOGPATH" "$CONF"
	done
}

runConf(){
	RATE=$1
	for i in "${variants[@]}"; do
		echo "Running config $RATE tps"
		runVariant "$RATE" "$i"
	done
}

for i in "${tps[@]}"; do
	runConf "$i"
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
