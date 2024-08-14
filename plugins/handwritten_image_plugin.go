package plugins

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"wechat-hub-plugin/hub"
)

type HandwrittenImagePlugin struct {
}

func NewHandwrittenImagePlugin() hub.Plugin {
	return &HandwrittenImagePlugin{}
}

func (h HandwrittenImagePlugin) Handle(ctx *hub.Context) error {
	if !strings.HasPrefix(ctx.Content, "#手写 ") {
		return nil
	}
	content := strings.TrimSpace(strings.TrimPrefix(ctx.Content, "#手写 "))
	if content == "" {
		return nil
	}
	img, err := h.getImage(content)
	if err != nil {
		slog.Error("[手写]获取图片失败", "error", err)
		ctx.Abort()
		return nil
	}
	if err := ctx.ReplayImg(fmt.Sprintf("%x.png", md5.Sum([]byte(content))), bytes.NewReader(img)); err != nil {
		slog.Error("[手写]上传图片失败", "error", err)
		ctx.Abort()
		return nil
	}
	return nil
}

func (h HandwrittenImagePlugin) getImage(content string) ([]byte, error) {
	resp, err := http.Get("https://api.52vmy.cn/api/img/tw?msg=" + url.QueryEscape(content))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
