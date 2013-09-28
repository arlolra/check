#!/usr/bin/env bash

HOME=/home/arlo
GIT=$HOME/git
TAR=$HOME/tar
CHECK=$GIT/check
TORDATA=/srv/tor
DNSEL=/srv/tordnsel.torproject.org/state
NOW=$(date +"%Y-%m-%d-%H-%M-%S")

cat $DNSEL/exit-addresses $DNSEL/exit-addresses.new > $CHECK/data/exit-lists/$NOW
cp $TORDATA/cached-consensus $CHECK/data/consensuses/$NOW-consensus
cat $TORDATA/cached-descriptors $TORDATA/cached-descriptors.new > $CHECK/data/cached-descriptors

cd $CHECK
PYTHONPATH=$GIT/stem:$TAR/six:$TAR/dateutil scripts/exitips.py
kill -s SIGUSR2 `cat check.pid`
