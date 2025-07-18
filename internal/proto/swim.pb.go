// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.29.1
// source: swim.proto

package fringe

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type State int32

const (
	State_ALIVE   State = 0
	State_SUSPECT State = 1
	State_DEAD    State = 2
	State_LEFT    State = 3
)

// Enum value maps for State.
var (
	State_name = map[int32]string{
		0: "ALIVE",
		1: "SUSPECT",
		2: "DEAD",
		3: "LEFT",
	}
	State_value = map[string]int32{
		"ALIVE":   0,
		"SUSPECT": 1,
		"DEAD":    2,
		"LEFT":    3,
	}
)

func (x State) Enum() *State {
	p := new(State)
	*p = x
	return p
}

func (x State) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (State) Descriptor() protoreflect.EnumDescriptor {
	return file_swim_proto_enumTypes[0].Descriptor()
}

func (State) Type() protoreflect.EnumType {
	return &file_swim_proto_enumTypes[0]
}

func (x State) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use State.Descriptor instead.
func (State) EnumDescriptor() ([]byte, []int) {
	return file_swim_proto_rawDescGZIP(), []int{0}
}

type MembershipUpdate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NodeId      string `protobuf:"bytes,1,opt,name=node_id,json=nodeId,proto3" json:"node_id,omitempty"`
	Address     string `protobuf:"bytes,2,opt,name=address,proto3" json:"address,omitempty"`
	Incarnation uint64 `protobuf:"varint,3,opt,name=incarnation,proto3" json:"incarnation,omitempty"`
	State       State  `protobuf:"varint,4,opt,name=state,proto3,enum=godis.State" json:"state,omitempty"`
}

func (x *MembershipUpdate) Reset() {
	*x = MembershipUpdate{}
	mi := &file_swim_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *MembershipUpdate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MembershipUpdate) ProtoMessage() {}

func (x *MembershipUpdate) ProtoReflect() protoreflect.Message {
	mi := &file_swim_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MembershipUpdate.ProtoReflect.Descriptor instead.
func (*MembershipUpdate) Descriptor() ([]byte, []int) {
	return file_swim_proto_rawDescGZIP(), []int{0}
}

func (x *MembershipUpdate) GetNodeId() string {
	if x != nil {
		return x.NodeId
	}
	return ""
}

func (x *MembershipUpdate) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *MembershipUpdate) GetIncarnation() uint64 {
	if x != nil {
		return x.Incarnation
	}
	return 0
}

func (x *MembershipUpdate) GetState() State {
	if x != nil {
		return x.State
	}
	return State_ALIVE
}

type Ping struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SenderId      string              `protobuf:"bytes,1,opt,name=sender_id,json=senderId,proto3" json:"sender_id,omitempty"`
	SenderAddress string              `protobuf:"bytes,2,opt,name=sender_address,json=senderAddress,proto3" json:"sender_address,omitempty"`
	TargetId      string              `protobuf:"bytes,3,opt,name=target_id,json=targetId,proto3" json:"target_id,omitempty"`
	Updates       []*MembershipUpdate `protobuf:"bytes,4,rep,name=updates,proto3" json:"updates,omitempty"`
}

func (x *Ping) Reset() {
	*x = Ping{}
	mi := &file_swim_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Ping) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Ping) ProtoMessage() {}

func (x *Ping) ProtoReflect() protoreflect.Message {
	mi := &file_swim_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Ping.ProtoReflect.Descriptor instead.
func (*Ping) Descriptor() ([]byte, []int) {
	return file_swim_proto_rawDescGZIP(), []int{1}
}

func (x *Ping) GetSenderId() string {
	if x != nil {
		return x.SenderId
	}
	return ""
}

func (x *Ping) GetSenderAddress() string {
	if x != nil {
		return x.SenderAddress
	}
	return ""
}

func (x *Ping) GetTargetId() string {
	if x != nil {
		return x.TargetId
	}
	return ""
}

func (x *Ping) GetUpdates() []*MembershipUpdate {
	if x != nil {
		return x.Updates
	}
	return nil
}

type PingReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SenderId       string              `protobuf:"bytes,1,opt,name=sender_id,json=senderId,proto3" json:"sender_id,omitempty"`
	SenderAddress  string              `protobuf:"bytes,2,opt,name=sender_address,json=senderAddress,proto3" json:"sender_address,omitempty"`
	TargetId       string              `protobuf:"bytes,3,opt,name=target_id,json=targetId,proto3" json:"target_id,omitempty"`
	TargetAddress  string              `protobuf:"bytes,4,opt,name=target_address,json=targetAddress,proto3" json:"target_address,omitempty"`
	RequestId      string              `protobuf:"bytes,5,opt,name=request_id,json=requestId,proto3" json:"request_id,omitempty"`
	RequestAddress string              `protobuf:"bytes,6,opt,name=request_address,json=requestAddress,proto3" json:"request_address,omitempty"`
	Updates        []*MembershipUpdate `protobuf:"bytes,7,rep,name=updates,proto3" json:"updates,omitempty"`
}

func (x *PingReq) Reset() {
	*x = PingReq{}
	mi := &file_swim_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *PingReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PingReq) ProtoMessage() {}

func (x *PingReq) ProtoReflect() protoreflect.Message {
	mi := &file_swim_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PingReq.ProtoReflect.Descriptor instead.
func (*PingReq) Descriptor() ([]byte, []int) {
	return file_swim_proto_rawDescGZIP(), []int{2}
}

func (x *PingReq) GetSenderId() string {
	if x != nil {
		return x.SenderId
	}
	return ""
}

func (x *PingReq) GetSenderAddress() string {
	if x != nil {
		return x.SenderAddress
	}
	return ""
}

func (x *PingReq) GetTargetId() string {
	if x != nil {
		return x.TargetId
	}
	return ""
}

func (x *PingReq) GetTargetAddress() string {
	if x != nil {
		return x.TargetAddress
	}
	return ""
}

func (x *PingReq) GetRequestId() string {
	if x != nil {
		return x.RequestId
	}
	return ""
}

func (x *PingReq) GetRequestAddress() string {
	if x != nil {
		return x.RequestAddress
	}
	return ""
}

func (x *PingReq) GetUpdates() []*MembershipUpdate {
	if x != nil {
		return x.Updates
	}
	return nil
}

type Ack struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Response      string              `protobuf:"bytes,1,opt,name=response,proto3" json:"response,omitempty"`
	SenderId      string              `protobuf:"bytes,2,opt,name=sender_id,json=senderId,proto3" json:"sender_id,omitempty"`
	SenderAddress string              `protobuf:"bytes,3,opt,name=sender_address,json=senderAddress,proto3" json:"sender_address,omitempty"`
	Incarnation   uint64              `protobuf:"varint,4,opt,name=incarnation,proto3" json:"incarnation,omitempty"`
	TargetId      string              `protobuf:"bytes,5,opt,name=target_id,json=targetId,proto3" json:"target_id,omitempty"`
	Updates       []*MembershipUpdate `protobuf:"bytes,6,rep,name=updates,proto3" json:"updates,omitempty"`
}

func (x *Ack) Reset() {
	*x = Ack{}
	mi := &file_swim_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Ack) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Ack) ProtoMessage() {}

func (x *Ack) ProtoReflect() protoreflect.Message {
	mi := &file_swim_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Ack.ProtoReflect.Descriptor instead.
func (*Ack) Descriptor() ([]byte, []int) {
	return file_swim_proto_rawDescGZIP(), []int{3}
}

func (x *Ack) GetResponse() string {
	if x != nil {
		return x.Response
	}
	return ""
}

func (x *Ack) GetSenderId() string {
	if x != nil {
		return x.SenderId
	}
	return ""
}

func (x *Ack) GetSenderAddress() string {
	if x != nil {
		return x.SenderAddress
	}
	return ""
}

func (x *Ack) GetIncarnation() uint64 {
	if x != nil {
		return x.Incarnation
	}
	return 0
}

func (x *Ack) GetTargetId() string {
	if x != nil {
		return x.TargetId
	}
	return ""
}

func (x *Ack) GetUpdates() []*MembershipUpdate {
	if x != nil {
		return x.Updates
	}
	return nil
}

type Envelope struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Msg:
	//
	//	*Envelope_Ping
	//	*Envelope_Ack
	//	*Envelope_PingReq
	Msg isEnvelope_Msg `protobuf_oneof:"msg"`
}

func (x *Envelope) Reset() {
	*x = Envelope{}
	mi := &file_swim_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Envelope) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Envelope) ProtoMessage() {}

func (x *Envelope) ProtoReflect() protoreflect.Message {
	mi := &file_swim_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Envelope.ProtoReflect.Descriptor instead.
func (*Envelope) Descriptor() ([]byte, []int) {
	return file_swim_proto_rawDescGZIP(), []int{4}
}

func (m *Envelope) GetMsg() isEnvelope_Msg {
	if m != nil {
		return m.Msg
	}
	return nil
}

func (x *Envelope) GetPing() *Ping {
	if x, ok := x.GetMsg().(*Envelope_Ping); ok {
		return x.Ping
	}
	return nil
}

func (x *Envelope) GetAck() *Ack {
	if x, ok := x.GetMsg().(*Envelope_Ack); ok {
		return x.Ack
	}
	return nil
}

func (x *Envelope) GetPingReq() *PingReq {
	if x, ok := x.GetMsg().(*Envelope_PingReq); ok {
		return x.PingReq
	}
	return nil
}

type isEnvelope_Msg interface {
	isEnvelope_Msg()
}

type Envelope_Ping struct {
	Ping *Ping `protobuf:"bytes,1,opt,name=ping,proto3,oneof"`
}

type Envelope_Ack struct {
	Ack *Ack `protobuf:"bytes,2,opt,name=ack,proto3,oneof"`
}

type Envelope_PingReq struct {
	PingReq *PingReq `protobuf:"bytes,3,opt,name=ping_req,json=pingReq,proto3,oneof"`
}

func (*Envelope_Ping) isEnvelope_Msg() {}

func (*Envelope_Ack) isEnvelope_Msg() {}

func (*Envelope_PingReq) isEnvelope_Msg() {}

var File_swim_proto protoreflect.FileDescriptor

var file_swim_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x73, 0x77, 0x69, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05, 0x67, 0x6f,
	0x64, 0x69, 0x73, 0x22, 0x8b, 0x01, 0x0a, 0x10, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68,
	0x69, 0x70, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x6e, 0x6f, 0x64, 0x65,
	0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6e, 0x6f, 0x64, 0x65, 0x49,
	0x64, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x69,
	0x6e, 0x63, 0x61, 0x72, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x0b, 0x69, 0x6e, 0x63, 0x61, 0x72, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x22, 0x0a,
	0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x0c, 0x2e, 0x67,
	0x6f, 0x64, 0x69, 0x73, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74,
	0x65, 0x22, 0x9a, 0x01, 0x0a, 0x04, 0x50, 0x69, 0x6e, 0x67, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65,
	0x6e, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73,
	0x65, 0x6e, 0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b,
	0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x64, 0x12, 0x31, 0x0a, 0x07, 0x75,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67,
	0x6f, 0x64, 0x69, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x22, 0x8c,
	0x02, 0x0a, 0x07, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65,
	0x6e, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73,
	0x65, 0x6e, 0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x73, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1b,
	0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x49, 0x64, 0x12, 0x25, 0x0a, 0x0e, 0x74,
	0x61, 0x72, 0x67, 0x65, 0x74, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x74, 0x61, 0x72, 0x67, 0x65, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65,
	0x73, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x69, 0x64,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x49,
	0x64, 0x12, 0x27, 0x0a, 0x0f, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x5f, 0x61, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x72, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x31, 0x0a, 0x07, 0x75, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x73, 0x18, 0x07, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f,
	0x64, 0x69, 0x73, 0x2e, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x52, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x22, 0xd7, 0x01,
	0x0a, 0x03, 0x41, 0x63, 0x6b, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x1b, 0x0a, 0x09, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x25,
	0x0a, 0x0e, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x69, 0x6e, 0x63, 0x61, 0x72, 0x6e, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x04, 0x52, 0x0b, 0x69, 0x6e, 0x63, 0x61,
	0x72, 0x6e, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x1b, 0x0a, 0x09, 0x74, 0x61, 0x72, 0x67, 0x65,
	0x74, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x61, 0x72, 0x67,
	0x65, 0x74, 0x49, 0x64, 0x12, 0x31, 0x0a, 0x07, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x18,
	0x06, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67, 0x6f, 0x64, 0x69, 0x73, 0x2e, 0x4d, 0x65,
	0x6d, 0x62, 0x65, 0x72, 0x73, 0x68, 0x69, 0x70, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x52, 0x07,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x73, 0x22, 0x81, 0x01, 0x0a, 0x08, 0x45, 0x6e, 0x76, 0x65,
	0x6c, 0x6f, 0x70, 0x65, 0x12, 0x21, 0x0a, 0x04, 0x70, 0x69, 0x6e, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0b, 0x2e, 0x67, 0x6f, 0x64, 0x69, 0x73, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x48,
	0x00, 0x52, 0x04, 0x70, 0x69, 0x6e, 0x67, 0x12, 0x1e, 0x0a, 0x03, 0x61, 0x63, 0x6b, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x67, 0x6f, 0x64, 0x69, 0x73, 0x2e, 0x41, 0x63, 0x6b,
	0x48, 0x00, 0x52, 0x03, 0x61, 0x63, 0x6b, 0x12, 0x2b, 0x0a, 0x08, 0x70, 0x69, 0x6e, 0x67, 0x5f,
	0x72, 0x65, 0x71, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x67, 0x6f, 0x64, 0x69,
	0x73, 0x2e, 0x50, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x71, 0x48, 0x00, 0x52, 0x07, 0x70, 0x69, 0x6e,
	0x67, 0x52, 0x65, 0x71, 0x42, 0x05, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x2a, 0x33, 0x0a, 0x05, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x12, 0x09, 0x0a, 0x05, 0x41, 0x4c, 0x49, 0x56, 0x45, 0x10, 0x00, 0x12,
	0x0b, 0x0a, 0x07, 0x53, 0x55, 0x53, 0x50, 0x45, 0x43, 0x54, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04,
	0x44, 0x45, 0x41, 0x44, 0x10, 0x02, 0x12, 0x08, 0x0a, 0x04, 0x4c, 0x45, 0x46, 0x54, 0x10, 0x03,
	0x42, 0x20, 0x5a, 0x1e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6a,
	0x73, 0x63, 0x6f, 0x74, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6f, 0x6d, 0x2f, 0x66, 0x72, 0x69, 0x6e,
	0x67, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_swim_proto_rawDescOnce sync.Once
	file_swim_proto_rawDescData = file_swim_proto_rawDesc
)

func file_swim_proto_rawDescGZIP() []byte {
	file_swim_proto_rawDescOnce.Do(func() {
		file_swim_proto_rawDescData = protoimpl.X.CompressGZIP(file_swim_proto_rawDescData)
	})
	return file_swim_proto_rawDescData
}

var file_swim_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_swim_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_swim_proto_goTypes = []any{
	(State)(0),               // 0: godis.State
	(*MembershipUpdate)(nil), // 1: godis.MembershipUpdate
	(*Ping)(nil),             // 2: godis.Ping
	(*PingReq)(nil),          // 3: godis.PingReq
	(*Ack)(nil),              // 4: godis.Ack
	(*Envelope)(nil),         // 5: godis.Envelope
}
var file_swim_proto_depIdxs = []int32{
	0, // 0: godis.MembershipUpdate.state:type_name -> godis.State
	1, // 1: godis.Ping.updates:type_name -> godis.MembershipUpdate
	1, // 2: godis.PingReq.updates:type_name -> godis.MembershipUpdate
	1, // 3: godis.Ack.updates:type_name -> godis.MembershipUpdate
	2, // 4: godis.Envelope.ping:type_name -> godis.Ping
	4, // 5: godis.Envelope.ack:type_name -> godis.Ack
	3, // 6: godis.Envelope.ping_req:type_name -> godis.PingReq
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_swim_proto_init() }
func file_swim_proto_init() {
	if File_swim_proto != nil {
		return
	}
	file_swim_proto_msgTypes[4].OneofWrappers = []any{
		(*Envelope_Ping)(nil),
		(*Envelope_Ack)(nil),
		(*Envelope_PingReq)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_swim_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_swim_proto_goTypes,
		DependencyIndexes: file_swim_proto_depIdxs,
		EnumInfos:         file_swim_proto_enumTypes,
		MessageInfos:      file_swim_proto_msgTypes,
	}.Build()
	File_swim_proto = out.File
	file_swim_proto_rawDesc = nil
	file_swim_proto_goTypes = nil
	file_swim_proto_depIdxs = nil
}
