// Copyright 2020-2026 ONDEWO GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Everything in this file names the ONDEWO SIP API specifically: its service, its messages, its
// enum. It is the ONLY file that differs from the sibling go clients - generated_code_test.go and
// auth_test.go are product agnostic and are copied over unchanged.
//
// Two tests the sibling clients have are deliberately ABSENT here rather than faked:
//
//   - a proto3 explicit-presence test. ondewo/sip/sip.proto declares no `optional` scalar at all
//     (grep the stubs for `proto3,oneof`: zero hits), so there is no field whose presence could be
//     asserted. The sibling clients that do have one keep the test.
//   - a streaming test. ondewo.sip.Sip has no streaming RPC - all 11 of its methods are unary -
//     so ServiceDesc.Streams is empty and the unary sweep in generated_code_test.go covers the
//     whole service on its own.
package tests

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	sip "github.com/ondewo/ondewo-sip-client-go/v5/api/ondewo/sip"
)

// protoFileCount is the number of .proto files below ondewo-sip-api/ondewo that the compiler
// consumed - SIP ships a single service proto, ondewo/sip/sip.proto. Every one of them has to end
// up in the global descriptor registry when this package is linked; a proto that silently stopped
// being compiled is otherwise invisible until a consumer misses a type.
const protoFileCount = 1

// services is every gRPC service this product exposes, keyed by the fully qualified proto name
// the ServiceDesc must declare.
var services = map[string]*grpc.ServiceDesc{
	"ondewo.sip.Sip": &sip.Sip_ServiceDesc,
}

// clientConstructors is the generated New<Service>Client of every service above. A client SDK
// that compiles but whose constructors are missing is useless, and the two generators that
// produce them (protoc-gen-go, protoc-gen-go-grpc) can disagree - so both halves are listed.
var clientConstructors = map[string]func(grpc.ClientConnInterface) any{
	"ondewo.sip.Sip": func(cc grpc.ClientConnInterface) any { return sip.NewSipClient(cc) },
}

// expectedMethods pins RPCs by name. The descriptor cross-check in generated_code_test.go proves
// the two generators agree with each other; it cannot notice an RPC that was renamed upstream,
// because both halves would be renamed together. These are spelled out so that a rename is a
// failing test rather than a silently broken consumer. All 11 RPCs of the service are listed -
// SIP has no streaming RPC, so ServiceDesc.Methods is the whole surface.
var expectedMethods = map[string][]string{
	"ondewo.sip.Sip": {
		"SipStartSession",
		"SipEndSession",
		"SipStartCall",
		"SipEndCall",
		"SipTransferCall",
		"SipRegisterAccount",
		"SipGetSipStatus",
		"SipGetSipStatusHistory",
		"SipPlayWavFiles",
		"SipMute",
		"SipUnMute",
	},
}

// TestMessageRoundTripsThroughTheWire is the core assertion about generated message code: a value
// built in go, serialized and parsed back is the same value. SipStatus is the message every RPC of
// this service answers with, and it covers a scalar, an enum, a `map<string, string>` and a
// well-known Timestamp at once, so a generator that mis-numbers a field or loses a nested type
// fails here.
func TestMessageRoundTripsThroughTheWire(t *testing.T) {
	t.Parallel()

	original := &sip.SipStatus{
		AccountName:    "sip-user-1@mydomain.com:5099",
		Timestamp:      timestamppb.New(referenceTime),
		StatusType:     sip.SipStatus_OUTGOING_CALL_CONNECTED,
		CalleeId:       "sip-user-2@mydomain.com",
		TransferCallId: "sip-user-3@mydomain.com",
		Headers: map[string]string{
			"X-Ondewo-Session": "6b1d0c3e-0f3a-4a53-9a4c-1b0a5a6f7c8d",
			"X-Ondewo-Caller":  "+4312345678",
		},
		Description:    "the outbound call is connected",
		NluSessionName: "projects/4c9a/agent/sessions/1f2e",
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal(%T) failed: %v", original, err)
	}
	if len(wire) == 0 {
		t.Fatal("proto.Marshal produced 0 bytes for a fully populated message")
	}

	parsed := &sip.SipStatus{}
	if err := proto.Unmarshal(wire, parsed); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if !proto.Equal(original, parsed) {
		t.Fatalf("round trip changed the message:\n original = %v\n  parsed = %v", original, parsed)
	}
	if got, want := parsed.GetHeaders()["X-Ondewo-Caller"], "+4312345678"; got != want {
		t.Errorf("map value after round trip = %q, want %q", got, want)
	}
	if got, want := parsed.GetStatusType(), sip.SipStatus_OUTGOING_CALL_CONNECTED; got != want {
		t.Errorf("enum after round trip = %v, want %v", got, want)
	}
	if got, want := parsed.GetTimestamp().AsTime().UTC(), referenceTime.UTC(); !got.Equal(want) {
		t.Errorf("timestamp after round trip = %v, want %v", got, want)
	}
}

// TestRepeatedNestedMessagesRoundTrip covers the shape SipGetSipStatusHistory answers with: a
// message holding a repeated field of another generated message type. A generator that loses a
// nested type or mis-numbers a repeated field fails here rather than in a consumer reading an
// empty list.
func TestRepeatedNestedMessagesRoundTrip(t *testing.T) {
	t.Parallel()

	original := &sip.SipStatusHistoryResponse{
		StatusHistory: []*sip.SipStatus{
			{
				AccountName: "sip-user-1@mydomain.com",
				StatusType:  sip.SipStatus_SESSION_STARTED,
				Timestamp:   timestamppb.New(referenceTime),
			},
			{
				AccountName: "sip-user-1@mydomain.com",
				StatusType:  sip.SipStatus_OUTGOING_CALL_INITIATED,
				Timestamp:   timestamppb.New(referenceTime.Add(90 * time.Second)),
				CalleeId:    "sip-user-2@mydomain.com",
			},
		},
	}

	wire, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("proto.Marshal(%T) failed: %v", original, err)
	}

	parsed := &sip.SipStatusHistoryResponse{}
	if err := proto.Unmarshal(wire, parsed); err != nil {
		t.Fatalf("proto.Unmarshal failed: %v", err)
	}

	if !proto.Equal(original, parsed) {
		t.Fatalf("round trip changed the message:\n original = %v\n  parsed = %v", original, parsed)
	}
	if got, want := len(parsed.GetStatusHistory()), 2; got != want {
		t.Fatalf("repeated message has %d entries after the round trip, want %d", got, want)
	}
	if got, want := parsed.GetStatusHistory()[1].GetStatusType(), sip.SipStatus_OUTGOING_CALL_INITIATED; got != want {
		t.Errorf("nested status type after round trip = %v, want %v", got, want)
	}
}

// TestEnumZeroValueIsPinned pins the member at 0 and the name maps generated beside it. SIP has no
// *_UNSPECIFIED enum at all: the zero member of SipStatus.StatusType is NO_SESSION, a real state.
// So the assertion is that the zero value is the one the API documents, not that it carries a
// particular name - renaming or reordering the members silently changes what an unset status field
// means, which is what this catches.
func TestEnumZeroValueIsPinned(t *testing.T) {
	t.Parallel()

	var zero sip.SipStatus_StatusType

	if zero != sip.SipStatus_NO_SESSION {
		t.Errorf("zero value of SipStatus_StatusType = %v, want NO_SESSION", zero)
	}
	if got, want := zero.String(), "NO_SESSION"; got != want {
		t.Errorf("SipStatus_StatusType(0).String() = %q, want %q", got, want)
	}
	if got, want := sip.SipStatus_StatusType_name[0], "NO_SESSION"; got != want {
		t.Errorf("SipStatus_StatusType_name[0] = %q, want %q", got, want)
	}
	if got, want := sip.SipStatus_StatusType_value["REGISTERED"], int32(sip.SipStatus_REGISTERED); got != want {
		t.Errorf("SipStatus_StatusType_value[REGISTERED] = %d, want %d", got, want)
	}
	if got, want := int32(sip.SipStatus_OUTGOING_CALL_CONNECTED), int32(5); got != want {
		t.Errorf("OUTGOING_CALL_CONNECTED = %d, want %d", got, want)
	}
}

// TestUnmarshalRejectsTruncatedInput asserts the generated message reports a parse error instead
// of accepting a malformed payload: field 1 (`account_name`) is announced as 5 bytes long but only
// 1 follows.
func TestUnmarshalRejectsTruncatedInput(t *testing.T) {
	t.Parallel()

	if err := proto.Unmarshal([]byte{0x0a, 0x05, 'a'}, &sip.SipStatus{}); err == nil {
		t.Fatal("proto.Unmarshal accepted a truncated payload, want an error")
	}
}

// sipServer is a fake ONDEWO server: it answers SipStartCall and inherits the "unimplemented"
// behaviour of the generated base type for every other RPC of the service.
type sipServer struct {
	sip.UnimplementedSipServer
}

func (sipServer) SipStartCall(_ context.Context, req *sip.SipStartCallRequest) (*sip.SipStatus, error) {
	return &sip.SipStatus{
		AccountName: "sip-user-1@mydomain.com",
		CalleeId:    req.GetCalleeId(),
		StatusType:  sip.SipStatus_OUTGOING_CALL_INITIATED,
		Headers:     req.GetHeaders(),
		Timestamp:   timestamppb.New(referenceTime),
	}, nil
}

// TestUnaryRPCRoundTripsOverAnInProcessServer drives the generated client stub, the generated
// server stub and the generated ServiceDesc against each other over a real gRPC connection - the
// request is marshalled, routed by the method name baked into the stub, and the response is
// parsed back. Nothing here is mocked except the transport, which is in memory.
func TestUnaryRPCRoundTripsOverAnInProcessServer(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		sip.RegisterSipServer(srv, sipServer{})
	})
	client := sip.NewSipClient(conn)

	const callee = "sip-user-2@mydomain.com"
	response, err := client.SipStartCall(t.Context(), &sip.SipStartCallRequest{
		CalleeId: callee,
		Headers:  map[string]string{"X-Ondewo-Caller": "+4312345678"},
	})
	if err != nil {
		t.Fatalf("SipStartCall failed: %v", err)
	}

	if got := response.GetCalleeId(); got != callee {
		t.Errorf("response callee id = %q, want %q", got, callee)
	}
	if got, want := response.GetStatusType(), sip.SipStatus_OUTGOING_CALL_INITIATED; got != want {
		t.Errorf("response status type = %v, want %v", got, want)
	}
	if got, want := response.GetHeaders()["X-Ondewo-Caller"], "+4312345678"; got != want {
		t.Errorf("the map sent in the request came back as %q, want %q", got, want)
	}
}

// TestUnimplementedMethodIsReportedAsUnimplemented pins the other half of the generated server
// contract: an RPC the server does not implement must come back as codes.Unimplemented, not as a
// routing failure or a panic. It also proves the method is routed at all - a method missing from
// the ServiceDesc would surface as a different code.
func TestUnimplementedMethodIsReportedAsUnimplemented(t *testing.T) {
	t.Parallel()

	conn := dialInProcess(t, nil, func(srv *grpc.Server) {
		sip.RegisterSipServer(srv, sipServer{})
	})
	client := sip.NewSipClient(conn)

	_, err := client.SipGetSipStatus(t.Context(), &emptypb.Empty{})
	if got := status.Code(err); got != codes.Unimplemented {
		t.Fatalf("SipGetSipStatus returned code %v (err = %v), want %v", got, err, codes.Unimplemented)
	}
}
