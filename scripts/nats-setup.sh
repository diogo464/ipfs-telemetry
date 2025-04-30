#!/usr/bin/env -S bash -x
set -e

check_stream_exists() {
  local stream_name="$1"
  if nats stream ls --names | grep -Fxq "$stream_name"; then
    return 0
  else
    return 1
  fi
}

check_consumer_exists() {
  local stream_name="$1"
  local consumer_name="$2"
  if nats consumer ls --names $stream_name | grep -Fxq "$consumer_name"; then
    return 0
  else
    return 1
  fi
}

if ! check_stream_exists "monitor" ; then
    nats stream add "monitor" \
        --subjects "monitor.*" \
        --description "monitor capture stream" \
        --retention limits \
        --storage file \
        --max-bytes 200GB \
        --defaults
fi

if ! check_consumer_exists monitor monitor-vm-otlp-exporter ; then
    nats consumer create monitor monitor-vm-otlp-exporter \
        --ack explicit \
        --filter "monitor.export" \
        --replay instant \
        --deliver all \
        --pull \
        --defaults
fi
