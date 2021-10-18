#!/usr/bin/env bash

# set -x

if [ "$GODO_NUM" == "exit" ]; then
  if ps -p ${PID} >/dev/null; then
    echo "CONTINUE"
  else
    echo "EXIT"
  fi
else
  top -b -n 1 -p ${PID}
fi
