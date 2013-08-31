#!/usr/bin/env python

import json
import operator

from stem.descriptor import parse_file


class Router():
    def __init__(self, address):
        self.Address = address
        self.Rules = []

exits = {}
for descriptor in parse_file("public/exit-addresses", "tordnsel 1.0"):
    descriptor.exit_addresses.sort(key=operator.itemgetter(1), reverse=True)
    exits[descriptor.fingerprint] = descriptor.exit_addresses[0][0]

with open("data/exit-policies", "w") as exit_file:
    for router in parse_file("data/consensus",
                             "network-status-consensus-3 1.0"):
        if router.fingerprint in exits and \
                router.exit_policy.is_exiting_allowed():
            r = Router(exits[router.fingerprint])
            for x in router.exit_policy._get_rules():
                r.Rules.append({
                    "Address": x.address,
                    "IsAccept": x.is_accept,
                    "MinPort": x.min_port,
                    "MaxPort": x.max_port
                })
            exit_file.write(json.dumps(r.__dict__) + "\n")
