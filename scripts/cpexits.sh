#!/usr/bin/env bash

CHECK=/opt/check
TORDATA=/srv/tor
DNSEL=/srv/tordnsel.torproject.org/state
NOW=$(date +"%Y-%m-%d-%H-%M-%S")

find $CHECK/data/exit-lists -type f -mtime +1 -delete
cat $DNSEL/exit-addresses $DNSEL/exit-addresses.new > $CHECK/data/exit-lists/$NOW

find $CHECK/data/consensuses -type f -mtime +1 -delete
cp $TORDATA/cached-consensus $CHECK/data/consensuses/$NOW-consensus

cat $TORDATA/cached-descriptors $TORDATA/cached-descriptors.new > $CHECK/data/cached-descriptors

cd $CHECK
scripts/exitips.py -n 1
kill -s SIGUSR2 `cat check.pid`
