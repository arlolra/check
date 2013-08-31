#!/usr/bin/env bash

cp /srv/tordnsel.torproject.org/state/exit-addresses /home/arlo/git/check/public/exit-addresses
cp /home/arlo/.tor/cached-consensus /home/arlo/git/check/data/consensus

cd /home/arlo/git/check
PYTHONPATH=/home/arlo/git/stem scripts/exitips.py
kill -s SIGUSR2 `cat check.pid`

