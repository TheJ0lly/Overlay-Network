#!/usr/bin/env bash

function get_random_time() {
    echo $((RANDOM % ($1 - $2 + 1) + $2))
}

function get_time() {
    echo $1
}

set -euo pipefail

BIN="./overlay"
BASE_PORT=9000
RAND="${1:-0}"
N="${2:-20}"
CONNCAP="${3:-3}"
QUEUECAP=1000
DEPTH="${4:-3}"
LIFE=2
DEATH=4

LIFE_DEATH_MIN=3
LIFE_DEATH_MAX=15

mkdir -p bench_logs
rm -f bench_logs/*.log

mkdir -p stats
rm -f stats/*.json

go build -o "$BIN" .

if [[ $RAND -eq 0 ]]; then
    BOOTSTRAP_LIFELINE=$LIFE
    BOOTSTRAP_DEATH=$DEATH
else
    BOOTSTRAP_LIFELINE=$(get_random_time LIFE_DEATH_MAX LIFE_DEATH_MIN)
    BOOTSTRAP_DEATH=$(get_random_time LIFE_DEATH_MAX LIFE_DEATH_MIN)
fi

echo "Starting bootstrap node on port $BASE_PORT with lifeline=$BOOTSTRAP_LIFELINE death=$BOOTSTRAP_DEATH"

# Start bootstrap node
$BIN \
  -ip 127.0.0.1 \
  -port $BASE_PORT \
  -conncap $CONNCAP \
  -queuecap $QUEUECAP \
  -lifeline $BOOTSTRAP_LIFELINE \
  -death $BOOTSTRAP_DEATH \
  -depth $DEPTH \
  -debug > "bench_logs/node_$BASE_PORT.log" 2>&1 &

sleep 1

# Join remaining nodes
for i in $(seq 1 $((N - 1))); do
  PORT=$((BASE_PORT + i))

  if [[ $RAND -eq 0 ]]; then
    NODE_LIFELINE=$LIFE
    NODE_DEATH=$DEATH
  else
    NODE_LIFELINE=$(get_random_time LIFE_DEATH_MAX LIFE_DEATH_MIN)
    NODE_DEATH=$(get_random_time LIFE_DEATH_MAX LIFE_DEATH_MIN)
  fi


  echo "Starting node on port $PORT with lifeline=$NODE_LIFELINE death=$NODE_DEATH"

  $BIN \
    -ip 127.0.0.1 \
    -port $PORT \
    -conncap $CONNCAP \
    -queuecap $QUEUECAP \
    -lifeline $NODE_LIFELINE \
    -death $NODE_DEATH \
    -depth $DEPTH \
    -newnet \
    -connip 127.0.0.1 \
    -connport $BASE_PORT \
    -debug > "bench_logs/node_$PORT.log" 2>&1 &

  sleep 7
done

echo "Started $N nodes"

sleep 60

pkill -9 -f overlay

echo "Done testing"