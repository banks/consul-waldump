// Copyright (c) HashiCorp, Inc.

package main

import (
	"fmt"

	"github.com/hashicorp/consul/agent/structs"
)

func summarize(op Operation) (Operation, error) {
	switch v := op.Value.(type) {
	case *structs.KVSRequest:
		return Operation{
			Type:  op.Type,
			Value: KVSummary{v.DirEnt.Key, len(v.DirEnt.Value)},
		}, nil

	case *structs.RegisterRequest:
		addr := v.Address
		svc := ""
		if v.Service != nil {
			svc = v.Service.Service
			if v.Service.Address != "" {
				addr = v.Service.Address
			}
			if v.Service.Port > 0 {
				addr = fmt.Sprintf("%s:%d", addr, v.Service.Port)
			}
		}
		return Operation{
			Type: op.Type,
			Value: RegisterSummary{
				Node:    v.Node,
				Service: svc,
				Addr:    addr,
			},
		}, nil

	default:
		// For other types just return the type name
		return Operation{
			Type: op.Type,
		}, nil
	}
}

type KVSummary struct {
	Key       string
	ValueSize int
}

type RegisterSummary struct {
	Node    string
	Service string `json:",omitempty"`
	Addr    string `json:",omitempty"`
}
