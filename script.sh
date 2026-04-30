#!/bin/bash

set -e

TRIALS=1
TRIAL_DURATION=10
LOGDIR="logs"
WORKERS=3
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
