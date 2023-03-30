// Copyright (c) HashiCorp, Inc.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/raft"
	wal "github.com/hashicorp/raft-wal"
	"github.com/hashicorp/raft-wal/fs"
	"github.com/hashicorp/raft-wal/segment"
	"github.com/hashicorp/raft-wal/types"
)

type opts struct {
	Dir    string
	After  uint64
	Before uint64
	Tail   bool
	Short  bool
}

func main() {
	var o opts
	flag.Uint64Var(&o.After, "after", 0, "specified a raft index to use as an exclusive lower bound when dumping log entries.")
	flag.Uint64Var(&o.Before, "before", 0, "specified a raft index to use as an exclusive upper bound when dumping log entries.")
	flag.BoolVar(&o.Tail, "t", false, "if specified the command will keep running and retry to print more logs every second when it gets to the end.")
	flag.BoolVar(&o.Short, "short", false, "if specified common commands will be displayed in a short form e.g. KV writes will just be key and payload size.")

	flag.Parse()

	// Accept dir as positional arg
	o.Dir = flag.Arg(0)
	if o.Dir == "" {
		fmt.Println("Usage: consul-waldump [-after INDEX] [-before INDEX] [-t] [-short] <path to WAL dir>")
		os.Exit(1)
	}
	if o.Tail && o.Before > 0 {
		fmt.Println("ERROR: can't use -before and -t (tail) at the same time")
		os.Exit(1)
	}

	vfs := fs.New()
	f := segment.NewFiler(o.Dir, vfs)

	codec := &wal.BinaryCodec{}
	var log raft.Log
	enc := json.NewEncoder(os.Stdout)

	type outputRecord struct {
		Index      uint64
		Op         interface{}
		AppendedAt time.Time
	}
	var record outputRecord

	dumpLog := func(info types.SegmentInfo, e types.LogEntry) (bool, error) {
		if info.Codec != wal.CodecBinaryV1 {
			return false, fmt.Errorf("unsupported codec %d in file %s", info.Codec, segment.FileName(info))
		}
		if err := codec.Decode(e.Data, &log); err != nil {
			return false, fmt.Errorf("error decoding wal entry index=%d: %w", e.Index, err)
		}

		record.Index = e.Index
		record.Op = fmt.Sprintf("internal raft log type %d", log.Type)
		record.AppendedAt = log.AppendedAt
		if log.Type == raft.LogCommand {
			// Decode consul operation
			out, err := decodeConsulLog(&log)
			if err != nil {
				return false, fmt.Errorf("error decoding consul operation index=%d: %w", e.Index, err)
			}
			if o.Short {
				if op, ok := out.(Operation); ok {
					out, err = summarize(op)
					if err != nil {
						return false, fmt.Errorf("error summarizing consul operation index=%d: %w", e.Index, err)
					}
				}
			}
			record.Op = out
		}

		// Output the record as JSON
		if err := enc.Encode(record); err != nil {
			return false, err
		}

		return true, nil
	}

	// Do initial dump
	err := f.DumpLogs(o.After, o.Before, dumpLog)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}

	if o.Tail {
		for {
			time.Sleep(1 * time.Second)
			// Dump anything new since the last log we found before
			err := f.DumpLogs(log.Index, 0, dumpLog)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
		}
	}
}
