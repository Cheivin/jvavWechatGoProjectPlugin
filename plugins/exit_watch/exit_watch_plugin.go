package exit_watch

import (
	"encoding/json"
	"log/slog"
	"strings"
	"wechat-hub-plugin/hub"
)

type Plugin struct {
}

func (p Plugin) Handle(ctx *hub.Context) error {
	message := ctx.Message
	if message.Event != "ExitGroup" {
		return nil
	}
	jsonData, err := json.Marshal(message.Data)
	if err != nil {
		slog.Error("退群消息 解析Data失败", "err", err)
		return nil
	}
	exitUsers := new([]hub.EventExitGroupUser)
	if err = json.Unmarshal(jsonData, exitUsers); err != nil {
		slog.Error("退群消息 转换数据失败", "err", err)
		return nil
	}
	usernames := make([]string, 0, len(*exitUsers))
	for _, user := range *exitUsers {
		usernames = append(usernames, user.Name)
	}
	_ = ctx.Sender.SendText(ctx.GID, "检测到退群:\n"+strings.Join(usernames, "\n"))
	return nil

}
