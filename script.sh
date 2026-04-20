#!/bin/bash

set -e

TRIALS=1
TRIAL_DURATION=10
LOGDIR="logs"
WORKERS=3
FOLLOWERS=2

verients=(
	"BASE"
	"FAIL_OVER"
	"REPLICATION"
	)

runTrial(){
	VERIANT=$1
	LOGPATH=$2

	SETUP=$VERIANT LOGPATH=$LOGPATH docker compose up -d --scale worker=${WORKERS} --scale follower=${FOLLOWERS}

	echo "trail will run for $TRIAL_DURATION second"
	sleep $TRIAL_DURATION
	echo "trial finished"
	docker compose down
}

runVerient(){
	VERIANT=$1
	for i in $(seq 1 $TRIALS); do
		echo "Running $VERIANT trail #$i"
		LOGPATH="$LOGDIR/$VERIANT/trial_$i"
		mkdir -p "$LOGPATH"
		runTrial "$VERIANT" "$LOGPATH/"
	done
}

for i in "${verients[@]}"; do
	runVerient "$i"
done
