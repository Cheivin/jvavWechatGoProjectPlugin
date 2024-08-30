package nga

import (
	"io/fs"
	"log/slog"
	"math/rand"
	"strings"
	"wechat-hub-plugin/hub"
)

type Plugin struct {
	f fs.FS
}

func (p Plugin) match(rawContent string) (content string, matched bool) {
	keywords := []string{
		"nga",
	}
	for _, keyword := range keywords {
		if strings.HasPrefix(rawContent, keyword) {
			matched = true
			content = strings.TrimPrefix(rawContent, keyword)
			return
		}
	}
	return
}
func New(f fs.FS) hub.Plugin {
	return &Plugin{f: f}
}
func (p Plugin) Handle(ctx *hub.Context) error {
	_, matched := p.match(ctx.Content)
	if !matched {
		return nil
	}
	img, err := p.getImage()
	if err != nil {
		slog.Error("[NGA]获取图片失败", "error", err)
		ctx.Abort()
		return nil
	}
	defer func() {
		_ = img.Close()
	}()
	info, err := img.Stat()
	if err != nil {
		slog.Error("[NGA]获取图片信息失败", "error", err)
		ctx.Abort()
		return nil
	}
	if err := ctx.ReplayImg(info.Name(), img); err != nil {
		slog.Error("[NGA]上传图片失败", "error", err)
		ctx.Abort()
		return nil
	}
	return nil
}

func (p Plugin) getImage() (fs.File, error) {
	var files []string
	if err := fs.WalkDir(p.f, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(d.Name()), ".jpg") || strings.HasSuffix(strings.ToLower(d.Name()), ".jpeg") || strings.HasSuffix(strings.ToLower(d.Name()), ".png") {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, nil
	}
	return p.f.Open(files[rand.Intn(len(files))])
}
