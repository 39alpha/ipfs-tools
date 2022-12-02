# ipfs-put

Put assets on the IPFS network and record the resulting hashes.

One of the great things about IPFS is content addressing. One of the painful things about IPFS is
content addressing. The issue is that there aren't a lot of great ways of keeping track of IPFS
hashes and what data they correspond to. This is particularly problematic if one of your objectives
is to host data for the world to access. It also makes getting the data for an analysis someone
tedious. Enter the `ipfs-put` and `ipfs-fetch` companion applications.

What `ipfs-put` does is take files or directories on your local file system, adds them to IPFS, and
stores a file mapping the hashes to the path of the file. This is ideal for providing a recipe for
downloading data for say, an analysis, and putting the files in exactly the right places (which is
where `ipfs-fetch` comes into play). It also, of course, ensures you have some record of what the
hashes are!

## Getting Started

Provided you have [Go](https://golang.org) installed, you can install `ipfs-put` by running

```shell
$ go get -u github.com/39alpha/ipfs-tools/ipfs-put
```

and you are all set!

## Usage

Let's say you are working in a git repository with the following structure:
```
data/
  2021-03-05/
    foo.csv
    bar.csv
  mnist.gz
```
which consists of two distinct datasets `data/2021-03-05` and `data/mnist.gz`. Suppose these files
are far to large to be handled cleanly via git or binary rather than plain text. You can store these
data on IPFS pretty easily, but keeping track of the hashes and where each dataset should be located
in order for your project to find them isn't something IPFS will do for you out of the box. So...

```
$ ipfs-put -o data.json data/2021-03-05 mnist.gz
```

will handle the bookkeeping for you. The result is a `data.json` file which looks something like
```
{
  "QmQPeNsJPyVWPFDVHb77w8G42Fvo15z4bG2X8D2GhfbSXc": "data/2021-03-05",
  "QmUgcKN8xiEC5ce8RMHo9SEMdnMJPhNV6FSFMXtmw53eZo": "data/mnist.gz"
}
```

From this file, it's possible in principle (or practice if you have `ipfs-fetch` installed) to
download these data directly to the locations specified to reproduce exactly the arrangement of data
on disk that you need!

In the event that you add new file, say `data/iris.csv`, you can just run
```
$ ipfs-put -o data.json data/iris.csv
```
and the hash and path will be added to the `data.json` file (or the file will be created if it
doesn't already exist).

The same thing goes for a modified file. Say you change the `data/2021-03-05/foo.csv` file, then you
can run
```
$ ipfs-put -o data.json data/2021-03-05
```
and the hash will for the dataset will be updated.


### Additional Flags

```shell
$ ipfs-put [OPTIONS] [FILE|DIR...]

Add assets from one or more files or directories to an IPFS node and
write a JSON object mapping CIDs to the path of the asset.

Options:
  -ipfsurl string
    	URL to running IPFS node (default "127.0.0.1:5001")
  -o string
    	File to which to write the payload
```

Aside from `-o` which specifies where the hash-to-path map should be stored, the only other flag
currently supported is `-ipfsurl`. If you have a non-standard IPFS configuration, then your node
might not be listening for API requests at the default `127.0.0.1:5001`. You can specify the correct
address via `-ipfsurl`.

## Planned Features

* `ipfs-put -o data.json -update`: re-add all of the datasets stored in the data.json file
