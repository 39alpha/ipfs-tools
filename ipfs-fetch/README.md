# ipfs-fetch

Fetch assets on the IPFS network and store them on your local file system.

One of the great things about IPFS is content addressing. One of the painful things about IPFS is
content addressing. The issue is that there aren't a lot of great ways of keeping track of IPFS
hashes and what data they correspond to. This is particularly problematic if one of your objectives
is to host data for the world to access. It also makes getting the data for an analysis someone
tedious. Enter the `ipfs-put` and `ipfs-fetch` companion applications.

What `ipfs-fetch` does is take a JSON-formatted file mapping IPFS hashes to filesystem paths, gets
each hash from IPFS and then stores them at the desired path. This is ideal for reporducing a
particular file/directory structure for say, an analysis. A great way of generating such a JSON file
is with the companion tool `ipfs-put`.

## Getting Started

Provided you have [Go](https://golang.org) installed, you can install `ipfs-fetch` by running

```shell
$ go get -u github.com/39alpha/ipfs-tools/ipfs-fetch
```

and you are all set!

## Usage

Let's say you are working in a git repository with a file called `data.json` that looks something
like this:

```
{
  "QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc": "data/2021-03-05",
  "QmUgcKN8xiEC5ce8RMHo9SEMdnMJPhNV6FSFMXtmw53eZo": "data/mnist.gz"
}
```

This file specifies that there are two datasets `data/2021-03-05` and `data/mnist.gz` with the
respective hashes. To fetch these data, all you need to do is

```
$ ipfs-fetch data.json
```

and the data will be downloaded through your locally running IPFS node, pinned (by default), and
stored at the paths specified by the data file. This may, for example, result in a directory
structure that looks something like this:

```
data/
  2021-03-05/
    foo.csv
    bar.csv
  mnist.gz
```

That's pretty much it! If you subsequently pull from your git remote and find that `data.json` has
changed, you can update your dataset by simply removing the old dataset (e.g. `rm -r data`) and
fetching again. A subsequent version will handle this processes more gracefully, of course.

### Additional Flags

```shell
$ ipfs-fetch
Usage: ipfs-fetch [OPTION]... [FILE]...

Fetch assets specified by [FILE]...

Options:

  -ipfsurl string
    	URL to running IPFS node (default "127.0.0.1:5001")
  -nopin
    	Do not pin the fetched asset
```

At the moment, `ipfs-fetch` only has two flags: `-ipfsurl` and `-nopin`. 

If you have a non-standard IPFS configuration, then your node might not be listening for API
requests at the default `127.0.0.1:5001`. You can specify the correct address via `-ipfsurl`.

By default, when `ipfs-fetch` downloads a dataset, it automatically pins them in you IPFS repo. This
is great in that it helps ensure that the data is distributed over the network and not just stored
in a centralized manner (say on the [39 Alpha gateway](https://gateway.39alpharesearch.org).
However, pinning ensures that the data will not be garbaged collected by your IPFS node and could
eat into your disk space. If you **do not** want to pin the data, just provide the `-nopin` flag.

## Planned Features

* Gracefully handle overwriting old/outdated dataset.
* Allow selective fetches: `ipfs-fetch data.json -- data/mnist.gz` to only download a single
  dataset
