#!/usr/bin/env bash
set -euo pipefail

BIN="./overlay"
BASE_PORT=9000
N="${1:-20}"
CONNCAP="${2:-3}"
QUEUECAP=1000
DEPTH="${3:-3}"


LIFE_DEATH_MIN=3
LIFE_DEATH_MAX=15

mkdir -p bench_logs
rm -f bench_logs/*.log

go build -o "$BIN" .

# Start bootstrap node
BOOTSTRAP_LIFELINE=$((RANDOM % (LIFE_DEATH_MAX - LIFE_DEATH_MIN + 1) + LIFE_DEATH_MIN))
BOOTSTRAP_DEATH=$((RANDOM % (LIFE_DEATH_MAX - LIFE_DEATH_MIN + 1) + LIFE_DEATH_MIN))

echo "Starting bootstrap node on port $BASE_PORT with lifeline=$BOOTSTRAP_LIFELINE death=$BOOTSTRAP_DEATH"

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

  NODE_LIFELINE=$((RANDOM % (LIFE_DEATH_MAX - LIFE_DEATH_MIN + 1) + LIFE_DEATH_MIN))
  NODE_DEATH=$((RANDOM % (LIFE_DEATH_MAX - LIFE_DEATH_MIN + 1) + LIFE_DEATH_MIN))

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