package hub

import "io"

type SenderInterface interface {
	SendText(gid string, content string) error
	SendNetworkImg(gid string, src string) error
	SendImg(gid string, filename string, file io.Reader) error
}

// Plugin 插件接口
type Plugin interface {
	Handle(message *Message, sender SenderInterface) error
}
