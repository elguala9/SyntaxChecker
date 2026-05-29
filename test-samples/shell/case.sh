#!/usr/bin/env bash
set -euo pipefail

handle() {
  local cmd="$1"
  case "$cmd" in
    start)
      echo "starting"
      ;;
    stop|halt)
      echo "stopping"
      ;;
    *)
      echo "unknown: ${cmd}" >&2
      return 1
      ;;
  esac
}

declare -a actions=("start" "stop")
for action in "${actions[@]}"; do
  handle "$action"
done
