[![Build Status](https://travis-ci.org/arlolra/check.png?branch=master)](https://travis-ci.org/arlolra/check)

== A new check.torproject.org, hopefully

  > Check could really use some love. Any volunteers please?
  >   --Roger
  >
  > https://lists.torproject.org/pipermail/tor-talk/2013-August/029306.html

== Setup

  for check.go:

    apt-get install gettext
    go get github.com/samuel/go-gettext/gettext
    make i18n
    make 

  for scripts/exitips.py:

    pip install -r requirements.txt

== Capacity planning

  54.7 requests/sec - 47.0 kB/second - 879 B/request
  487 requests currently being processed, 63 idle workers
  that's check right now
  (43 million hits over the past 9 days)

== License

  MIT
