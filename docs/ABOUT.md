## What is check.torproject.org?

The core service is a website that allows a user to check if their connection seems to be a [Tor](https://www.torproject.org/) connection or not. This allows the user to know if they have successfully setup their client and are ready to browse the internet privately. It is the default homepage for users of the [Tor Browser Bundle (TBB)](https://www.torproject.org/projects/torbrowser.html.en) for versions below 3.X.

Users of the TBB will also recieve a notification if there are updates available for their software.

Check also provides a [bulk exit list tool](https://check.torproject.org/cgi-bin/TorBulkExitList.py) which website owners can enter their server IP and recieve a list of IPs that could potentially make connections to their server on a specific port from the Tor network.

## How does it work?

The Tor network has a public list of exit nodes (see below for more information). Each exit node has an exit policy which is a list of ip and port combinations that it accepts or rejects. For every user that hits the website, we check if their IP matches a known exit node's IP, and that its exit policy would allow accessing this website on port 443(HTTPS). If it is, then we can make a good assumption that the user has successfully connected via the Tor network.

### Determining an accurate, up-to-date list of exit nodes

The Tor network publishes an hourly [consensus](https://metrics.torproject.org/data.html#relaydesc) document which holds information about every known relay and exit node in the Tor network. We use this document to know which exit nodes are 'online' at any given moment.

Using the list of relays known to be online, we parse the full exit policy in each server's [server descriptor](https://metrics.torproject.org/data.html#relaydesc) file to determine what's allowed and what isn't.

The IPs from each exit node in the consensus are self published. Instead of trusting that as being correct, we use the [exit lists](https://metrics.torproject.org/data.html#exitlist) published by the [TorDNSel](https://www.torproject.org/projects/tordnsel.html.en) service to determine the public IPs per exit node if they're different from the self published ones. 

The formats for these files can be found [here](https://metrics.torproject.org/formats.html#serverdesc) and [here](https://metrics.torproject.org/formats.html#exitlist).

### Want to help?

Feel free to send a PR here if you spot anything you can help with. If you want to contribute in other ways, please consider:
 - [Volunteering](https://www.torproject.org/getinvolved/volunteer.html.en)
 - [Translating TorCheck](https://www.transifex.com/projects/p/torproject/resource/2-torcheck-torcheck-pot/)
 - [Translating other Tor projects](https://www.transifex.com/projects/p/torproject/resources/)
 - [Donating to The Tor Project](https://www.torproject.org/donate/donate.html.en)