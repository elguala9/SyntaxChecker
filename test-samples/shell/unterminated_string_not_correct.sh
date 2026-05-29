#!/usr/bin/env bash

# The double-quoted string is never closed.
message="hello world
echo "$message"
