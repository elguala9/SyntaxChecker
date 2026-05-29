#!/usr/bin/env bash
set -euo pipefail

greet() {
  local name="$1"
  echo "hello, ${name}"
}

for n in alice bob carol; do
  greet "$n"
done
