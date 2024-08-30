package write

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

type Plugin struct {
}

func New() hub.Plugin {
	return &Plugin{}
}
func (h Plugin) match(rawContent string) (content string, matched bool) {
	keywords := []string{
		"手写",
		"write",
	}
	for _, keyword := range keywords {
		if strings.HasPrefix(rawContent, "#"+keyword+" ") {
			matched = true
			content = strings.TrimPrefix(rawContent, keyword)
			return
		}
	}
	return
}
func (h Plugin) Handle(ctx *hub.Context) error {
	content, matched := h.match(ctx.Content)
	if !matched || content == "" {
		return nil
	}
	img, err := h.getImage(content, ctx.Username)
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

func (h Plugin) getImage(content string, author string) ([]byte, error) {
	resp, err := http.Get("https://api.52vmy.cn/api/img/tw?msg=" + url.QueryEscape(content+"\n\u202E——"+author))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}
