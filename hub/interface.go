package hub

import (
	"io"
)

type SenderInterface interface {
	SendText(gid string, content string) error
	SendNetworkImg(gid string, src string) error
	SendImg(gid string, filename string, file io.Reader) error
}

// Plugin 插件接口
type Plugin interface {
	Handle(ctx *Context) error
}

type Context struct {
	*Message
	Sender SenderInterface
	abort  bool
}

func (ctx *Context) IsAbort() bool {
	return ctx.abort
}
func (ctx *Context) Abort() {
	ctx.abort = true
}

func (ctx *Context) ReplayText(content string) error {
	return ctx.Sender.SendText(ctx.GID, content)
}

func (ctx *Context) ReplayImg(filename string, file io.Reader) error {
	return ctx.Sender.SendImg(ctx.GID, filename, file)
}

func (ctx *Context) ReplayNetworkImg(src string) error {
	return ctx.Sender.SendNetworkImg(ctx.GID, src)
}
