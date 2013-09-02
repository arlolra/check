#!/usr/bin/env bash

cat /srv/tordnsel.torproject.org/state/exit-addresses /srv/tordnsel.torproject.org/state/exit-addresses.new > /home/arlo/git/check/public/exit-addresses
cp /home/arlo/.tor/cached-consensus /home/arlo/git/check/data/consensus

cd /home/arlo/git/check
PYTHONPATH=/home/arlo/git/stem scripts/exitips.py
kill -s SIGUSR2 `cat check.pid`

