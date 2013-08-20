#!/usr/bin/env python

from stem.descriptor import parse_file, DocumentHandler

exits = {}

# replace with stem tordnsel parser
with open('public/exit-addresses') as file:
    exit_node = ""
    for line in file:
        line = line.split()
        if line[0] == "ExitNode":
            exit_node = line[1]
        elif line[0] == "ExitAddress":
            exits[exit_node] = line[1]

with open('data/consensus', 'rb') as consensus_file, \
        open('data/exit-policies', 'w') as exit_file:
    for router in parse_file(
        consensus_file,
        'network-status-consensus-3 1.0',
        document_handler = DocumentHandler.ENTRIES
    ):
        if router.fingerprint in exits and \
                router.exit_policy.is_exiting_allowed():
            exit_file.write("%s %s\n" %
                (exits[router.fingerprint], router.exit_policy))
