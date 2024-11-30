package generator

import (
	"bytes"
	"io"
	"text/template"
)

var clientStubTemplate = template.Must(template.New("client").Parse(`
// Code generated by stream-rpc. DO NOT EDIT.
package {{.PackageName}}

import (
	rpc "stream-rpc"
)

type {{.ServiceName}}Client struct {
	peer *rpc.RpcPeer
}

func New{{.ServiceName}}Client(peer *rpc.RpcPeer) *{{.ServiceName}}Client {
	return &{{.ServiceName}}Client{peer: peer}
}

{{range .Methods}}
func (c *{{$.ServiceName}}Client) {{.Name}}(req *{{.InputType}}) *{{.OutputType}} {
	resp := &{{.OutputType}}{}
	err := c.peer.Call("{{$.ServiceName}}.{{.Name}}", req, resp)
	if err != nil {
		return nil
	}
	return resp
}
{{end}}
`))

var serverStubTemplate = template.Must(template.New("server").Parse(`
// Code generated by stream-rpc. DO NOT EDIT.
package {{.PackageName}}

import (
	rpc "stream-rpc"
	"context"
)

// UnimplementedCalculatorServer can be embedded to have forward compatible implementations
type Unimplemented{{.ServiceName}}Server struct {}

type {{.ServiceName}}Server interface {
	{{range .Methods}}
	{{.Name}}(context.Context, *{{.InputType}}) (*{{.OutputType}})
	{{end}}
}

type {{.ServiceName}}ServerImpl struct {
	impl {{.ServiceName}}Server
}

func Register{{.ServiceName}}Server(peer *rpc.RpcPeer, impl {{.ServiceName}}Server) {
	server := &{{.ServiceName}}ServerImpl{impl: impl}
	peer.RegisterService("{{.ServiceName}}", server)
}

{{range .Methods}}
func (s *Unimplemented{{$.ServiceName}}Server) {{.Name}}(ctx context.Context, req *{{.InputType}}) (*{{.OutputType}}) {
	return nil
}
{{end}}

{{range .Methods}}
func (s *{{$.ServiceName}}ServerImpl) {{.Name}}(ctx context.Context, req *{{.InputType}}) (*{{.OutputType}}) {
	return s.impl.{{.Name}}(ctx, req)
}
{{end}}
`))

var skeletonTemplate = template.Must(template.New("skeleton").Parse(`
// Code generated by protoc-gen-stream-rpc. DO NOT EDIT.
package service

import (
	"context"
	proto "{{.ProtoPackage}}"
)

// {{.ServiceName}}Service implements the {{.ServiceName}} service
type {{.ServiceName}}Service struct {
	proto.Unimplemented{{.ServiceName}}Server
}

{{range .Methods}}
// {{.Name}} implements {{$.ServiceName}}Server
func (s *{{$.ServiceName}}Service) {{.Name}}(ctx context.Context, req *proto.{{.InputType}}) *proto.{{.OutputType}} {
	// TODO: Implement your logic here
	return &proto.{{.OutputType}}{}
}
{{end}}
`))

func GenerateSkeleton(w io.Writer, data TemplateData) error {
	var buf bytes.Buffer
	if err := skeletonTemplate.Execute(&buf, data); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}
