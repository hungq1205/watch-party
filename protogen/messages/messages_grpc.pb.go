// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: messages.proto

package messages

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	MessageService_Sent_FullMethodName                = "/MessageService/Sent"
	MessageService_Delete_FullMethodName              = "/MessageService/Delete"
	MessageService_GetMessage_FullMethodName          = "/MessageService/GetMessage"
	MessageService_QueryMessages_FullMethodName       = "/MessageService/QueryMessages"
	MessageService_DeleteMessages_FullMethodName      = "/MessageService/DeleteMessages"
	MessageService_CreateMessageBox_FullMethodName    = "/MessageService/CreateMessageBox"
	MessageService_DeleteMessageBox_FullMethodName    = "/MessageService/DeleteMessageBox"
	MessageService_GetMessageBox_FullMethodName       = "/MessageService/GetMessageBox"
	MessageService_GetMessageBoxOfUser_FullMethodName = "/MessageService/GetMessageBoxOfUser"
	MessageService_RemoveUserFromBox_FullMethodName   = "/MessageService/RemoveUserFromBox"
	MessageService_AddUserToBox_FullMethodName        = "/MessageService/AddUserToBox"
)

// MessageServiceClient is the client API for MessageService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MessageServiceClient interface {
	Sent(ctx context.Context, in *SentRequest, opts ...grpc.CallOption) (*MessageIdentifier, error)
	Delete(ctx context.Context, in *MessageIdentifier, opts ...grpc.CallOption) (*ActionResponse, error)
	GetMessage(ctx context.Context, in *MessageIdentifier, opts ...grpc.CallOption) (*Message, error)
	QueryMessages(ctx context.Context, in *QueryMessageRequest, opts ...grpc.CallOption) (*QueryMessageResponse, error)
	DeleteMessages(ctx context.Context, in *QueryMessageRequest, opts ...grpc.CallOption) (*ActionResponse, error)
	CreateMessageBox(ctx context.Context, in *UserGroup, opts ...grpc.CallOption) (*MessageBoxIdentifier, error)
	DeleteMessageBox(ctx context.Context, in *MessageBoxIdentifier, opts ...grpc.CallOption) (*ActionResponse, error)
	GetMessageBox(ctx context.Context, in *MessageBoxIdentifier, opts ...grpc.CallOption) (*UserGroup, error)
	GetMessageBoxOfUser(ctx context.Context, in *UserIdentifier, opts ...grpc.CallOption) (*BoxGroup, error)
	RemoveUserFromBox(ctx context.Context, in *UserBox, opts ...grpc.CallOption) (*ActionResponse, error)
	AddUserToBox(ctx context.Context, in *UserBox, opts ...grpc.CallOption) (*ActionResponse, error)
}

type messageServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewMessageServiceClient(cc grpc.ClientConnInterface) MessageServiceClient {
	return &messageServiceClient{cc}
}

func (c *messageServiceClient) Sent(ctx context.Context, in *SentRequest, opts ...grpc.CallOption) (*MessageIdentifier, error) {
	out := new(MessageIdentifier)
	err := c.cc.Invoke(ctx, MessageService_Sent_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) Delete(ctx context.Context, in *MessageIdentifier, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, MessageService_Delete_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) GetMessage(ctx context.Context, in *MessageIdentifier, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := c.cc.Invoke(ctx, MessageService_GetMessage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) QueryMessages(ctx context.Context, in *QueryMessageRequest, opts ...grpc.CallOption) (*QueryMessageResponse, error) {
	out := new(QueryMessageResponse)
	err := c.cc.Invoke(ctx, MessageService_QueryMessages_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) DeleteMessages(ctx context.Context, in *QueryMessageRequest, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, MessageService_DeleteMessages_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) CreateMessageBox(ctx context.Context, in *UserGroup, opts ...grpc.CallOption) (*MessageBoxIdentifier, error) {
	out := new(MessageBoxIdentifier)
	err := c.cc.Invoke(ctx, MessageService_CreateMessageBox_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) DeleteMessageBox(ctx context.Context, in *MessageBoxIdentifier, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, MessageService_DeleteMessageBox_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) GetMessageBox(ctx context.Context, in *MessageBoxIdentifier, opts ...grpc.CallOption) (*UserGroup, error) {
	out := new(UserGroup)
	err := c.cc.Invoke(ctx, MessageService_GetMessageBox_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) GetMessageBoxOfUser(ctx context.Context, in *UserIdentifier, opts ...grpc.CallOption) (*BoxGroup, error) {
	out := new(BoxGroup)
	err := c.cc.Invoke(ctx, MessageService_GetMessageBoxOfUser_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) RemoveUserFromBox(ctx context.Context, in *UserBox, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, MessageService_RemoveUserFromBox_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *messageServiceClient) AddUserToBox(ctx context.Context, in *UserBox, opts ...grpc.CallOption) (*ActionResponse, error) {
	out := new(ActionResponse)
	err := c.cc.Invoke(ctx, MessageService_AddUserToBox_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MessageServiceServer is the server API for MessageService service.
// All implementations must embed UnimplementedMessageServiceServer
// for forward compatibility
type MessageServiceServer interface {
	Sent(context.Context, *SentRequest) (*MessageIdentifier, error)
	Delete(context.Context, *MessageIdentifier) (*ActionResponse, error)
	GetMessage(context.Context, *MessageIdentifier) (*Message, error)
	QueryMessages(context.Context, *QueryMessageRequest) (*QueryMessageResponse, error)
	DeleteMessages(context.Context, *QueryMessageRequest) (*ActionResponse, error)
	CreateMessageBox(context.Context, *UserGroup) (*MessageBoxIdentifier, error)
	DeleteMessageBox(context.Context, *MessageBoxIdentifier) (*ActionResponse, error)
	GetMessageBox(context.Context, *MessageBoxIdentifier) (*UserGroup, error)
	GetMessageBoxOfUser(context.Context, *UserIdentifier) (*BoxGroup, error)
	RemoveUserFromBox(context.Context, *UserBox) (*ActionResponse, error)
	AddUserToBox(context.Context, *UserBox) (*ActionResponse, error)
	mustEmbedUnimplementedMessageServiceServer()
}

// UnimplementedMessageServiceServer must be embedded to have forward compatible implementations.
type UnimplementedMessageServiceServer struct {
}

func (UnimplementedMessageServiceServer) Sent(context.Context, *SentRequest) (*MessageIdentifier, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Sent not implemented")
}
func (UnimplementedMessageServiceServer) Delete(context.Context, *MessageIdentifier) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}
func (UnimplementedMessageServiceServer) GetMessage(context.Context, *MessageIdentifier) (*Message, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessage not implemented")
}
func (UnimplementedMessageServiceServer) QueryMessages(context.Context, *QueryMessageRequest) (*QueryMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryMessages not implemented")
}
func (UnimplementedMessageServiceServer) DeleteMessages(context.Context, *QueryMessageRequest) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteMessages not implemented")
}
func (UnimplementedMessageServiceServer) CreateMessageBox(context.Context, *UserGroup) (*MessageBoxIdentifier, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateMessageBox not implemented")
}
func (UnimplementedMessageServiceServer) DeleteMessageBox(context.Context, *MessageBoxIdentifier) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteMessageBox not implemented")
}
func (UnimplementedMessageServiceServer) GetMessageBox(context.Context, *MessageBoxIdentifier) (*UserGroup, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessageBox not implemented")
}
func (UnimplementedMessageServiceServer) GetMessageBoxOfUser(context.Context, *UserIdentifier) (*BoxGroup, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMessageBoxOfUser not implemented")
}
func (UnimplementedMessageServiceServer) RemoveUserFromBox(context.Context, *UserBox) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveUserFromBox not implemented")
}
func (UnimplementedMessageServiceServer) AddUserToBox(context.Context, *UserBox) (*ActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddUserToBox not implemented")
}
func (UnimplementedMessageServiceServer) mustEmbedUnimplementedMessageServiceServer() {}

// UnsafeMessageServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MessageServiceServer will
// result in compilation errors.
type UnsafeMessageServiceServer interface {
	mustEmbedUnimplementedMessageServiceServer()
}

func RegisterMessageServiceServer(s grpc.ServiceRegistrar, srv MessageServiceServer) {
	s.RegisterService(&MessageService_ServiceDesc, srv)
}

func _MessageService_Sent_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SentRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).Sent(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_Sent_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).Sent(ctx, req.(*SentRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageIdentifier)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_Delete_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).Delete(ctx, req.(*MessageIdentifier))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_GetMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageIdentifier)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).GetMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_GetMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).GetMessage(ctx, req.(*MessageIdentifier))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_QueryMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).QueryMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_QueryMessages_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).QueryMessages(ctx, req.(*QueryMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_DeleteMessages_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).DeleteMessages(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_DeleteMessages_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).DeleteMessages(ctx, req.(*QueryMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_CreateMessageBox_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserGroup)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).CreateMessageBox(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_CreateMessageBox_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).CreateMessageBox(ctx, req.(*UserGroup))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_DeleteMessageBox_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageBoxIdentifier)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).DeleteMessageBox(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_DeleteMessageBox_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).DeleteMessageBox(ctx, req.(*MessageBoxIdentifier))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_GetMessageBox_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageBoxIdentifier)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).GetMessageBox(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_GetMessageBox_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).GetMessageBox(ctx, req.(*MessageBoxIdentifier))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_GetMessageBoxOfUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserIdentifier)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).GetMessageBoxOfUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_GetMessageBoxOfUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).GetMessageBoxOfUser(ctx, req.(*UserIdentifier))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_RemoveUserFromBox_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserBox)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).RemoveUserFromBox(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_RemoveUserFromBox_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).RemoveUserFromBox(ctx, req.(*UserBox))
	}
	return interceptor(ctx, in, info, handler)
}

func _MessageService_AddUserToBox_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UserBox)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MessageServiceServer).AddUserToBox(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: MessageService_AddUserToBox_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MessageServiceServer).AddUserToBox(ctx, req.(*UserBox))
	}
	return interceptor(ctx, in, info, handler)
}

// MessageService_ServiceDesc is the grpc.ServiceDesc for MessageService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MessageService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "MessageService",
	HandlerType: (*MessageServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Sent",
			Handler:    _MessageService_Sent_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _MessageService_Delete_Handler,
		},
		{
			MethodName: "GetMessage",
			Handler:    _MessageService_GetMessage_Handler,
		},
		{
			MethodName: "QueryMessages",
			Handler:    _MessageService_QueryMessages_Handler,
		},
		{
			MethodName: "DeleteMessages",
			Handler:    _MessageService_DeleteMessages_Handler,
		},
		{
			MethodName: "CreateMessageBox",
			Handler:    _MessageService_CreateMessageBox_Handler,
		},
		{
			MethodName: "DeleteMessageBox",
			Handler:    _MessageService_DeleteMessageBox_Handler,
		},
		{
			MethodName: "GetMessageBox",
			Handler:    _MessageService_GetMessageBox_Handler,
		},
		{
			MethodName: "GetMessageBoxOfUser",
			Handler:    _MessageService_GetMessageBoxOfUser_Handler,
		},
		{
			MethodName: "RemoveUserFromBox",
			Handler:    _MessageService_RemoveUserFromBox_Handler,
		},
		{
			MethodName: "AddUserToBox",
			Handler:    _MessageService_AddUserToBox_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "messages.proto",
}
