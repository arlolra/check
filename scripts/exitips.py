#!/usr/bin/env python

import sys
import json
import getopt
import operator

from os import listdir
from dateutil.parser import parse
from dateutil.tz import tzutc, tzlocal
from datetime import datetime
from math import floor

from stem.descriptor import parse_file
from stem.exit_policy import AddressType


class Router():
    def __init__(self, router, tminus):
        self.Fingerprint = router.fingerprint
        self.Address = router.address
        self.IsAllowedDefault = router.exit_policy._is_allowed_default
        self.Rules = []
        self.Tminus = tminus


def get_hours(td):
    try:
        s = td.total_seconds()
    except AttributeError:
        # workaround for py2.6
        s = (td.microseconds + (td.seconds + td.days * 24 * 3600) * 1e6) / 1e6
    return int(floor(s / 3600))


def main(consensuses, exit_lists):
    exits = {}
    now = datetime.now(tzlocal())

    for f in consensuses:

        # strip -consensus
        d = f[:-10]

        # consensus from t hours ago
        p = parse(d).replace(tzinfo=tzutc())
        t = get_hours(now - p)

        # read in consensus and store routes in exits
        for router in parse_file("data/consensuses/" + f,
                                 "network-status-consensus-3 1.0"):
            if router.fingerprint in exits:
                continue
            if router.exit_policy.is_exiting_allowed():
                r = Router(router, t)
                for x in router.exit_policy._get_rules():
                    r.Rules.append({
                        "IsAddressWildcard": True,
                        "IsAccept": x.is_accept,
                        "MinPort": x.min_port,
                        "MaxPort": x.max_port
                    })
                exits[router.fingerprint] = r

        # get a corresponding exit list
        m = [x for x in exit_lists if x.startswith(d[:-5])]
        if len(m) == 0:
            continue

        # update exit addresses with data from TorDNSEL
        for descriptor in parse_file("data/exit-lists/" + m[0],
                                     "tordnsel 1.0"):
            descriptor.exit_addresses.sort(key=operator.itemgetter(1),
                                           reverse=True)
            e = exits.get(descriptor.fingerprint, None)
            if e is not None and e.Tminus == d:
                e.Address = descriptor.exit_addresses[0][0]

    # update all with server descriptor info
    for descriptor in parse_file("data/cached-descriptors",
                                 "server-descriptor 1.0"):
        if descriptor.fingerprint in exits:
            rules = []
            for x in descriptor.exit_policy._get_rules():
                is_address_wildcard = x.is_address_wildcard()
                mask = None
                if not is_address_wildcard:
                    address_type = x.get_address_type()
                    if (address_type == AddressType.IPv4 and
                        x._masked_bits != 32) or \
                        (address_type == AddressType.IPv6 and
                            x._masked_bits != 128):
                        mask = x.get_mask()
                rules.append({
                    "IsAddressWildcard": is_address_wildcard,
                    "Address": x.address,
                    "Mask": mask,
                    "IsAccept": x.is_accept,
                    "MinPort": x.min_port,
                    "MaxPort": x.max_port
                })
            exits[descriptor.fingerprint].Rules = rules

    # output exits to file
    with open("data/exit-policies", "w") as exit_file:
        for e in exits:
            exit_file.write(json.dumps(exits[e].__dict__) + "\n")


def print_help(c):
    print "Usage:", sys.argv[0], "-n <int>"
    sys.exit(c)

if __name__ == "__main__":

    # data
    consensuses = listdir("data/consensuses")
    consensuses.sort(reverse=True)
    exit_lists = listdir("data/exit-lists")

    # args
    try:
        opts, args = getopt.getopt(sys.argv[1:], "hn:", [])
    except getopt.GetoptError:
        print_help(2)
    for opt, arg in opts:
        if opt == "-h":
            print_help(0)
        elif opt in ("-n"):
            consensuses = consensuses[:int(arg)]

    main(consensuses, exit_lists)
