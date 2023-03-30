// Copyright (c) HashiCorp, Inc.

package main

import (
	"fmt"

	"github.com/hashicorp/consul/agent/structs"
	"github.com/hashicorp/consul/proto/pbpeering"
	"github.com/hashicorp/raft"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func decodeConsulLog(log *raft.Log) (interface{}, error) {
	if len(log.Data) < 1 {
		return nil, fmt.Errorf("unknown log encoding")
	}

	// Decode based on they first byte as Consul (currently) does.
	typeByte := structs.MessageType(log.Data[0])

	// Strip off the "ignore unknown type" bit if set
	typeByte &= ^structs.IgnoreUnknownTypeFlag

	typ, err := structForType(typeByte)
	if err != nil {
		return nil, err
	}
	switch v := typ.(type) {
	case Operation:
		// Just output it as it is
		return v, nil

	case protoreflect.ProtoMessage:
		// If it's PB, decode it as that
		if err := structs.DecodeProto(log.Data[1:], v); err != nil {
			return nil, err
		}
		return Operation{typeByte.String(), v}, nil

	default:
		// For everything else, assume it's msgpack encoded
		if err := structs.Decode(log.Data[1:], &typ); err != nil {
			return nil, err
		}
		return Operation{typeByte.String(), v}, nil
	}
}

type Operation struct {
	Type  string
	Value interface{} `json:",omitempty"`
}

func structForType(t structs.MessageType) (interface{}, error) {
	switch t {
	case structs.RegisterRequestType:
		return &structs.RegisterRequest{}, nil

	case structs.DeregisterRequestType:
		return &structs.DeregisterRequest{}, nil

	case structs.KVSRequestType:
		return &structs.KVSRequest{}, nil

	case structs.SessionRequestType:
		return &structs.SessionRequest{}, nil

	case structs.DeprecatedACLRequestType:
		return nil, fmt.Errorf("Legacy ACL request type removed since Consul 1.15")

	case structs.TombstoneRequestType:
		return &structs.TombstoneRequest{}, nil

	case structs.CoordinateBatchUpdateType:
		return &structs.Coordinates{}, nil

	case structs.PreparedQueryRequestType:
		return &structs.PreparedQueryRequest{}, nil

	case structs.TxnRequestType:
		return &structs.TxnRequest{}, nil

	case structs.AutopilotRequestType:
		return &structs.AutopilotSetConfigRequest{}, nil

	case structs.AreaRequestType:
		// Enterprise type, don't error but we don't have the code to decode OSS
		return Operation{Type: t.String()}, nil

	case structs.ACLBootstrapRequestType:
		return &structs.RegisterRequest{}, nil

	case structs.IntentionRequestType:
		return &structs.IntentionRequest{}, nil

	case structs.ConnectCARequestType:
		return &structs.CARequest{}, nil

	case structs.ConnectCAProviderStateType:
		return nil, fmt.Errorf("%s should only exist in FSM snapshots", t)

	case structs.ConnectCAConfigType:
		return nil, fmt.Errorf("%s should only exist in FSM snapshots", t)

	case structs.IndexRequestType:
		return nil, fmt.Errorf("%s should only exist in FSM snapshots", t)

	case structs.ACLTokenSetRequestType:
		return &structs.ACLTokenBatchSetRequest{}, nil

	case structs.ACLTokenDeleteRequestType:
		return &structs.ACLTokenBatchDeleteRequest{}, nil

	case structs.ACLPolicySetRequestType:
		return &structs.ACLPolicyBatchSetRequest{}, nil

	case structs.ACLPolicyDeleteRequestType:
		return &structs.ACLPolicyBatchDeleteRequest{}, nil

	case structs.ConnectCALeafRequestType:
		return &structs.CALeafRequest{}, nil

	case structs.ConfigEntryRequestType:
		return &structs.ConfigEntryRequest{}, nil

	case structs.ACLRoleSetRequestType:
		return &structs.ACLRoleBatchSetRequest{}, nil

	case structs.ACLRoleDeleteRequestType:
		return &structs.ACLRoleBatchDeleteRequest{}, nil

	case structs.ACLBindingRuleSetRequestType:
		return &structs.ACLBindingRuleBatchSetRequest{}, nil

	case structs.ACLBindingRuleDeleteRequestType:
		return &structs.ACLBindingRuleBatchDeleteRequest{}, nil

	case structs.ACLAuthMethodSetRequestType:
		return &structs.ACLAuthMethodBatchSetRequest{}, nil

	case structs.ACLAuthMethodDeleteRequestType:
		return &structs.ACLAuthMethodBatchDeleteRequest{}, nil

	case structs.ChunkingStateType:
		return nil, fmt.Errorf("%s should only exist in FSM snapshots", t)

	case structs.FederationStateRequestType:
		return &structs.FederationStateRequest{}, nil

	case structs.SystemMetadataRequestType:
		return &structs.SystemMetadataRequest{}, nil

	case structs.ServiceVirtualIPRequestType:
		return nil, fmt.Errorf("%s should only exist in FSM snapshots", t)

	case structs.FreeVirtualIPRequestType:
		return nil, fmt.Errorf("%s should only exist in FSM snapshots", t)

	case structs.KindServiceNamesType:
		return nil, fmt.Errorf("%s should only exist in FSM snapshots", t)

	case structs.PeeringWriteType:
		return &pbpeering.PeeringWriteRequest{}, nil

	case structs.PeeringDeleteType:
		return &pbpeering.PeeringDeleteRequest{}, nil

	case structs.PeeringTrustBundleWriteType:
		return &pbpeering.PeeringTrustBundleWriteRequest{}, nil

	case structs.PeeringTrustBundleDeleteType:
		return &pbpeering.PeeringTrustBundleDeleteRequest{}, nil

	case structs.PeeringSecretsWriteType:
		return &pbpeering.SecretsWriteRequest{}, nil

	case structs.RaftLogVerifierCheckpoint:
		return Operation{Type: "log verifier checkpoint"}, nil

	default:
		// Unknown type (possibly Enterprise only or new since this tool was
		// updated). Don't error but we don't have the code to decode OSS.
		return Operation{Type: t.String()}, nil
	}
}
