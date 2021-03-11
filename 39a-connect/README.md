# 39a-connect

Painlessly connect to the 39 Alpha Research IPFS gateway <https://gateway.39alpharesearch.org/>.

When downloading content from IPFS, you generally benefit by being directly connected to an IPFS
peer which can provide the data you are looking for. If that's not the case, then your request will
have to crawl around a bit until such a node is found. Since all of 39 Alpha's data is pinned to our
IPFS gateway node, you can expedite downloads by first establishing a peer connection.

As part of our general API, we provide the `GET https://39alpharesearch.org/api/v0/ipfs/addr`
endpoint which responds with our IPFS multiaddrs. What `39a-connect` does is simply make a request
to our API and then establish a connection using your local node's `POST /api/v0/swarm/connect`.

## Getting Started

Provided you have [Go](https://golang.org) installed, you can install `39a-connect` by running

```shell
$ go get -u github.com/39alpha/ipfs-tools/39a-connect
```

##  Usage

For the most part, all you have to do is
```shell
$ 39a-connect
```
And you are done!

The command takes only a handful of command line flags.

```shell
$ 39a-connect -h
Usage of 39a-connect:
  -api string
    	39A API version (default "api/v0")
  -ipfsurl string
    	URL to running IPFS node (default "127.0.0.1:5001")
  -v	Print verbose status messages
```

The most useful flag is `-v` which simply prints the status of the connection process, in particular
the multiaddr you're connecting to:

```shell
$ 39a-connect -v
Connecting to /ip4/44.241.131.183/udp/4001/quic/p2p/12D3KooWADaUtfpf7hahJooJoxB9fNgMcb7kXRUWkvFmPT1kxb5Q ... done
$ ipfs swarm peers | grep 12D3KooWADaUtfpf7hahJooJoxB9fNgMcb7kXRUWkvFmPT1kxb5Q
/ip4/44.241.131.183/udp/4001/quic/p2p/12D3KooWADaUtfpf7hahJooJoxB9fNgMcb7kXRUWkvFmPT1kxb5Q
```

In the event that you have a non-standard IPFS configuration, you may need to specify the address
and port at which your local node is listening for API requests. That can be done with the
`-ipfsurl` flag:

```shell
$ 39a-connect -ipfsurl 127.0.0.1:8080
ERROR: cannot establish connection to IPFS shell — shell is not up
```

Of course, if your node is not actually listening there, you'll get an error message.
