#!/bin/sh

if [ "$NET_DELAY" = "true" ]; then
	echo "seting delay"
	tc qdisc add dev eth0 root netem delay ${DELAY_MEAN:-30ms} ${DELAY_JITTER:-10ms} distribution normal
fi

exec ./app
