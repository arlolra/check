#!/usr/bin/env python

import json
import operator

from stem.descriptor import parse_file
from stem.exit_policy import AddressType


class Router():
    def __init__(self, router):
        self.Address = router.address
        self.IsAllowedDefault = router.exit_policy._is_allowed_default
        self.Rules = []

exit_addresses = {}
for descriptor in parse_file("public/exit-addresses", "tordnsel 1.0"):
    descriptor.exit_addresses.sort(key=operator.itemgetter(1), reverse=True)
    exit_addresses[descriptor.fingerprint] = descriptor.exit_addresses[0][0]

server_descriptors = {}
for descriptor in parse_file("data/all_descriptors", "server-descriptor 1.0"):
    server_descriptors[descriptor.fingerprint] = descriptor.exit_policy

with open("data/exit-policies", "w") as exit_file:
    for router in parse_file("data/consensus",
                             "network-status-consensus-3 1.0"):
        if router.exit_policy.is_exiting_allowed():
            r = Router(router)
            if router.fingerprint in exit_addresses:
                r.Address = exit_addresses[router.fingerprint]
            if router.fingerprint in server_descriptors:
                for x in server_descriptors[router.fingerprint]._get_rules():
                    is_address_wildcard = x.is_address_wildcard()
                    mask = None
                    if not is_address_wildcard:
                        address_type = x.get_address_type()
                        if (address_type == AddressType.IPv4 and
                            x._masked_bits != 32) or \
                            (address_type == AddressType.IPv6 and
                                x._masked_bits != 128):
                            mask = x.get_mask()
                    r.Rules.append({
                        "IsAddressWildcard": is_address_wildcard,
                        "Address": x.address,
                        "Mask": mask,
                        "IsAccept": x.is_accept,
                        "MinPort": x.min_port,
                        "MaxPort": x.max_port
                    })
            else:
                for x in router.exit_policy._get_rules():
                    r.Rules.append({
                        "IsAddressWildcard": True,
                        "IsAccept": x.is_accept,
                        "MinPort": x.min_port,
                        "MaxPort": x.max_port
                    })
            exit_file.write(json.dumps(r.__dict__) + "\n")
