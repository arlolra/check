#!/usr/bin/env bash

BASE=/home/arlo/git/check
$BASE/check -base="$BASE" -log="$BASE/check.log" -pid="$BASE/check.pid" &

