#!/usr/bin/env bash

nohup ./godo -shell ./hello.sh -span 30s -nums 1-5 2>&1 >> godo.out &
