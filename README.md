# consul-waldump

A quick, unofficial tool for debugging Consul's raft logs when using the [WAL
storage
backend](https://developer.hashicorp.com/consul/docs/agent/wal-logstore).

## Usage

```
$ consul-waldump [-after INDEX] [-before INDEX] [-t] [-short] /path/to/wal/dir
...
{"Index":54870,"Op":{"Type":"KVS","Value":{"Datacenter":"dc1","Op":"set","DirEnt":{"LockIndex":0,"Key":"quickly/eagerly/evolving/falcon","Flags":0,"Value":"UmdEUEE3VWowbXVkY0krZ3d0RmZzYlFTenRVWEJOVFltOEllV2ZnRjFwMnF0emZDMjF0Yk1jUXdqaGpnWFJwSm9jdWRUN0NRc0UycTlnR1hNbmpxS01RU213MjVLOC95NVZ4SnpDK014TGRxaURxSmJXd09hRGZxRjFNU1o0d0QvWTFhc0dWRW83a3ZMVEVJUGs2WXdZcGZvbTJyUjdEUVJTanhJQkZSYkZPMk1TM1JXREdxN2xFdUwwV0lRUll6RjdhOXR2bVJJbDFJcDczNktYQm1BeTl4QkxmZ3VPZGlZS0J1b3JPTk84QW54UDJjVlRCUWFBOUFEQnNiNm9QSTVBPT0=","CreateIndex":0,"ModifyIndex":0},"Token":""}},"AppendedAt":"2023-03-28T13:28:48.05928+01:00"}
...
```

Each log entry is written out as JSON followed by a newline. Top level fields
are:

```jsonc
{
  "Index": 12345, // The Raft index of the log entry
  "Op": {}, // The raw consul operation (or a description for internal raft types)
  "AppendedAt": "2023-03-28T13:28:48.05928+01:00" // The time the log was written on the leader
}
```

By default all Consul OSS operation types are dumped verbatim as JSON objects.
Consul Enterprise types are not available to OSS code like this so will show up
as "Unknown Type".

The `-short` flag will summarize the most common operation types:
 * **KV writes:** will output `Key` and `ValueSize` - the number of bytes in the value.
 * **Registrations:** will output `Node` name and, if present, `Service` and `Address`
 * All other types will just output the type description

`-after` and `-before` can be used to limit the range output. The tool scans all
segment files since it doesn't have metadata to know if they are sealed or where
the index is stored if they are. When `-after` or `-before` are non-zero, the
tool will skip segment files that appear to be entirely outside of the range.

`-t` or "tail" specifies that the tool when it reaches the end of the logs
stored should wait for 1 second and then try to read more. It will do this
repeatedly until terminated. Internally this re-scans the entire tail segment
every time so is not especially efficient, but on modern hardware scanning a
64MiB file shouldn't take that long. If you have very constrained disk IO or
RAM, this might impact performance of the running Consul server, but for typical
production-grade hardware this is very unlikely to have a measurable impact on
the running server since the most recent file data is likely in the OS page
cache anyway.

## Limitations

### No Consistency Guarantees

This tool is designed for debugging only. It does _not_ inspect the wal-meta
database. This has the nice property that you can safely dump the contexts of
WAL files even while the application is still writing to the WAL since we don't
have to take a lock on the meta database.

The downside is that this *not perfectly consistent* with the Consul process
writing the logs. **It must not be used to build change data capture or
replication solutions external to Consul.**

Since it's not consistent with the Consul process, it may miss logs, output
already truncated logs, output logs both before and after a truncation in
arbitrary order. Most importantly, logs being present on the disk of one server
doesn't mean they are committed by the cluster yet so even if it appears
perfect, it would be inconsistent to treat the logs output from one server's log
as a source of truth.

For debugging cases, these are very unlikely to be observed though.

### Consul Version support

This tool imports code that is normally internal to Consul and may change in
future releases. It was originally built against Consul 1.15.1. The current
version built against can be seen in `go.mod`.

No guarantees are made that it will be updated to work with any future version!
PRs are welcome though.

## Feedback

If you use this tool and find it useful or fork it to build something similar,
please consider opening an issue to let us know your use-case. If we see enough
interest it may become a more official tool as part of the Consul binary itself.
