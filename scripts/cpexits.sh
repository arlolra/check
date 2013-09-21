#!/usr/bin/env bash

HOME=/home/arlo
GIT=$HOME/git
CHECK=$GIT/check
TORDATA=/srv/tor
DNSEL=/srv/tordnsel.torproject.org/state

cat $DNSEL/exit-addresses $DNSEL/exit-addresses.new > $CHECK/data/exit-addresses
cp $TORDATA/cached-consensus $CHECK/data/consensus
cat $TORDATA/cached-descriptors $TORDATA/cached-descriptors.new > $CHECK/data/all_descriptors

cd $CHECK
PYTHONPATH=$GIT/stem scripts/exitips.py
kill -s SIGUSR2 `cat check.pid`
