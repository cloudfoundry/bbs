package main

var grpcImports = `
import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9
`

var clientMethodName = `
{{.Service.Name}}_{{.MethodName}}_FullMethodName = "/models.{{.Service.Name}}/{{.MethodName}}"`

var grpcClient = `
const (
	{{.ClientMethodNames}}
)
// {{.Service.Name}}Client is the client API for {{.Service.Name}} service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type {{.Service.Name}}Client interface {
	{{.ClientInterfaceMethods}}
}

type {{.Service.Name | LowerFirst}}Client struct { 
	cc grpc.ClientConnInterface
}

func New{{.Service.Name}}Client(cc grpc.ClientConnInterface) {{.Service.Name}}Client {
	return &{{.Service.Name | LowerFirst}}Client{cc}
}

{{.ClientMethods}}
`

var clientInterfaceMethod = `
{{.MethodName}}(ctx context.Context, in *{{.MethodName}}Request, opts ...grpc.CallOption) (*{{.MethodName}}Response, error)`

var clientMethod = `
func (c *{{.Service.Name | LowerFirst}}Client) {{.MethodName}}(ctx context.Context, in *{{.MethodName}}Request, opts ...grpc.CallOption) (*{{.MethodName}}Response, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(Proto{{.MethodName}}Response)
	err := c.cc.Invoke(ctx, {{.Service.Name}}_{{.MethodName}}_FullMethodName, in.ToProto(), out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out.FromProto(), nil
}
`

var grpcServer = `
type {{.Service.Name}}Server interface {
	{{.ServerInterfaceMethods}}
	mustEmbedUnimplemented{{.Service.Name}}Server()
}

// Unimplemented{{.Service.Name}}Server must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type Unimplemented{{.Service.Name}}Server struct {}

{{.UnimplementedServerMethods}}
func (Unimplemented{{.Service.Name}}Server) mustEmbedUnimplemented{{.Service.Name}}Server() {}
func (Unimplemented{{.Service.Name}}Server) testEmbeddedByValue() {}

// Unsafe{{.Service.Name}}Server may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to {{.Service.Name}}Server will
// result in compilation errors.
type Unsafe{{.Service.Name}}Server interface {
	mustEmbedUnimplemented{{.Service.Name}}Server()
}

func Register{{.Service.Name}}Server(s grpc.ServiceRegistrar, srv {{.Service.Name}}Server) {
	// If the following call panics, it indicates Unimplemented{{.Service.Name}}Server was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&{{.Service.Name}}_ServiceDesc, srv)
}

{{.ServerHandlers}}
`

var unimplementedServerMethod = `
func (Unimplemented{{.Service.Name}}Server) {{.MethodName}}(context.Context, *{{.MethodName}}Request) (*{{.MethodName}}Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method {{.MethodName}} not implemented")
}`

var serverInterfaceMethod = `
{{.MethodName}}(context.Context, *{{.MethodName}}Request) (*{{.MethodName}}Response, error)`

var serverHandler = `
func _{{.Service.Name}}_{{.MethodName}}_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new({{.MethodName}}Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.({{.Service.Name}}Server).{{.MethodName}}(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: {{.Service.Name}}_{{.MethodName}}_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.({{.Service.Name}}Server).{{.MethodName}}(ctx, req.(*{{.MethodName}}Request))
	}
	return interceptor(ctx, in, info, handler)
}
`

var grpcServiceDescription = `
// {{.Service.Name}}_ServiceDesc is the grpc.ServiceDesc for {{.Service.Name}} service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var {{.Service.Name}}_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "models.{{.Service.Name}}",
	HandlerType: (*{{.Service.Name}}Server)(nil),
	{{.Methods}},
	Streams:  []grpc.StreamDesc{},
	Metadata: "{{.Service.Source}}",
}
`

var grpcMethodDesc = `
{
	MethodName: "{{.MethodName}}",
	Handler:    _{{.Service.Name}}_{{.MethodName}}_Handler,
}`
