package plugins

import (
	"log/slog"
	"wechat-hub-plugin/hub"
)

type SamePlugin struct {
}

func (p *SamePlugin) Handle(ctx *hub.Context) error {
	slog.Info("SamePlugin receive message", "type", ctx.MsgType, "content", ctx.Content)
	defer ctx.Abort()
	if "#same" == ctx.Content {
		return ctx.ReplayText("hello same")
	}
	return nil
}
