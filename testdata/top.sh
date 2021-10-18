#!/usr/bin/env bash

# set -x

# usage example
# # busy &
# [1] 7448
# # 2021/10/18 12:49:05 busy starting, pid 7448
# 2021/10/18 12:49:05  run 100% of 2/2 CPU cores forever.
# PID=7448 godo -shell ./top.sh -span 5s

if [ "$GODO_NUM" == "exitCheck" ]; then
  if ps -p "${PID}" >/dev/null; then
    echo "CONTINUE"
  else
    echo "EXIT"
  fi
else
  top -b -n 1 -p "${PID}"
fi
