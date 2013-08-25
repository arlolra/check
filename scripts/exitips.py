#!/usr/bin/env python

import operator

from stem.descriptor import parse_file

exits = {}

# waiting on trac #8255
for descriptor in parse_file("public/exit-addresses", "tordnsel 1.0"):
    descriptor.exit_addresses.sort(key=operator.itemgetter(1), reverse=True)
    exits[descriptor.fingerprint] = descriptor.exit_addresses[0][0]

with open("data/exit-policies", "w") as exit_file:
    for router in parse_file("data/consensus",
                             "network-status-consensus-3 1.0"):
        if router.fingerprint in exits and \
                router.exit_policy.is_exiting_allowed():
            exit_file.write("%s %s\n" %
                            (exits[router.fingerprint], router.exit_policy))
