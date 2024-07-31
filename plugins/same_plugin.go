package plugins

import (
	"log/slog"
	"wechat-hub-plugin/hub"
)

type SamePlugin struct {
}

func (p *SamePlugin) Handle(message *hub.Message, sender hub.SenderInterface) error {
	slog.Info("SamePlugin receive message", "type", message.MsgType, "content", message.Content)
	if "#same" == message.Content {
		return sender.SendText(message.GID, "hello same")
	}
	return nil
}
