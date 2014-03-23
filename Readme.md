[![Build Status](https://travis-ci.org/arlolra/check.png?branch=master)](https://travis-ci.org/arlolra/check)

## A new check.torproject.org, hopefully

> Check could really use some love. Any volunteers please?  
>   --Roger
>
> https://lists.torproject.org/pipermail/tor-talk/2013-August/029306.html

## Documentation

See `/docs` for an idea of what's going on here.

## Development

Generating the exit list requires [stem](https://stem.torproject.org/), Tor's `python` controller library. Assuming you have a `virtualenv` ready, just:

    pip install -r requirements.txt

Alternatively,

    apt-get install python-dateutil python-stem  # stem 1.1 is in unstable

For the server itself, you'll need `go` and `gettext`. Installing that might look like:

    apt-get install git golang gettext
    go get github.com/samuel/go-gettext/gettext

Then you can run `make` and wait for `git` and `rsync` to fetch all the data and launch the server.

Please run the tests before sending a pull request:

    make test

## Production

The data that `make start` pulls in will quickly become stale. What you want to do is run a `tor` instance with the following configurations in your `torrc`:

    FetchDirInfoEarly 1
    FetchDirInfoExtraEarly 1
    FetchUselessDescriptors 1
    UseMicrodescriptors 0
    DownloadExtraInfo 1

Then setup a cron job to run a script like `scripts/cpexits.sh` every hour. Setting up TorDNSEL to get the exit addresses is beyond the scope of this readme.


## Setup

Assuming debian, install the dependencies,

    apt-get install git golang gettext python-dateutil python-stem
    go get github.com/samuel/go-gettext/gettext

The cron job and init script assume a base directory of `/opt/check`.

    git clone https://git.torproject.org/check.git /opt/check
    cd /opt/check
    make i18n
    make exits
    make install

Then you start it up,

    /etc/init.d/check start

## /exit-addresses

The production check.tpo symlinks TorDNSEL's state file, `exit-addresses`,
to its `DocumentRoot`. This is to replace what was formerly at exitlist.tpo

## Translations

Translations are maintained in [Transifex][1]. See Tor's
[translation overview][2] to get involved.

[1]: https://www.transifex.com/projects/p/torproject/resource/2-torcheck-torcheck-pot/
[2]: https://www.torproject.org/getinvolved/translation-overview.html.en

## Capacity planning

> 54.7 requests/sec - 47.0 kB/second - 879 B/request  
> 487 requests currently being processed, 63 idle workers  
> that's check right now (43 million hits over the past 9 days)

## License

MIT
