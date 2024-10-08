package plugins

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"wechat-hub-plugin/hub"
)

type SamePlugin struct {
	image_arr []string
	Model     string // 模型名称
}

// 构造函数
func NewSamePlugin() *SamePlugin {
	return &SamePlugin{Model: "realisticVisionV13_v13"}
}

func (p *SamePlugin) Init() {
	p.Model = "realisticVisionV13_v13"
	slog.Info("SamePlugin init")
}

func (p *SamePlugin) refreshImages() string {

	// 获取当前images目录下的所有图片的路径并保存到image_arr中
	imageDir := "cache/images"
	if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
		slog.Error("Failed to create directory", "error", err)
		return ""
	}
	files, err := os.ReadDir(imageDir)
	if err != nil {
		slog.Error("Failed to read images directory", "error", err)
		return "Failed to refresh images"
	}

	var images []string
	for _, file := range files {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" {
				images = append(images, filepath.Join(imageDir, file.Name()))
			}
		}
	}
	p.image_arr = images
	slog.Info("Images refreshed successfully", "count", len(p.image_arr))
	return "Images refreshed successfully"
}

func (p *SamePlugin) randomImage() string {
	// 随机获取一个图片的路径
	if len(p.image_arr) == 0 {
		return "No image"
	}
	return p.image_arr[rand.Intn(len(p.image_arr))]
}

func (p *SamePlugin) checkoutModel(name string) error {

	models := model_list()
	if len(models) == 0 {
		return errors.New("No models found")
	}
	// 如果name不存在或为空，则随机选择一个模型
	found := false
	for _, model := range models {
		if model == name {
			found = true
			break
		}
	}
	if !found {
		slog.Warn("Model not found", "model", name)
		p.Model = models[rand.Intn(len(models))]
		return errors.New("Model not found")
	}
	if name == "" {
		slog.Info("Model not specified, choosing a random one")
		p.Model = models[rand.Intn(len(models))]
		return errors.New("Model not specified")
	}
	p.Model = name
	slog.Info("Model checked out", "model", p.Model)
	return nil
}

func (p *SamePlugin) textToImage(prompt string) string {
	url := "http://127.0.0.1:7860/sdapi/v1/txt2img"
	payload := map[string]interface{}{
		"prompt": prompt,
		"steps":  25,
		"override_settings": map[string]interface{}{
			"sd_model_checkpoint": p.Model,
			// "CLIP_stop_at_last_layers": 2,
		},
	}

	// 将负载编码为JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal payload", "error", err)
		return ""
	}

	// 发送POST请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error("Failed to send POST request", "error", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		return ""
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("Failed to unmarshal response", "error", err)
		return ""
	}

	// 解码并保存图像
	imageData, err := base64.StdEncoding.DecodeString(result["images"].([]interface{})[0].(string))
	if err != nil {
		slog.Error("Failed to decode image data", "error", err)
		return ""
	}
	// 通过prompt的hash值生成图片文件名
	imageDir := "cache/images"
	imagePath := filepath.Join(imageDir, fmt.Sprintf("%x.png", md5.Sum([]byte(prompt))))
	if err := ioutil.WriteFile(imagePath, imageData, 0644); err != nil {
		slog.Error("Failed to write image file", "error", err)
		return ""
	}

	slog.Info("Image saved successfully", "path", imagePath)
	return imagePath
}
func model_list() []string {
	models := []string{
		"chikmix_V1",
		"chilloutmix_NiPrunedFp32Fix",
		"deliberate_v2",
		"novalai_model",
		"perfectWorld_perfectWorldBakedVAE",
		"realcartoon3d_v8",
		"realdosmix_ (1)",
		"realdosmix_",
		"realisticVisionV13_v13",
		"samaritan3dCartoon_samaritan3dCartoonV3",
	}
	return models
}

type CommandHandler func(ctx *hub.Context, p *SamePlugin) error

var handlers = map[string]CommandHandler{
	"#same":        handleSame,
	"#same_setu":   handleSameSetu,
	"#txt2img":     handleTxt2Img,
	"#check_model": handleCheckModel,
	"#model_list":  handleModelList,
	"#model":       handleModel,
}

func (p *SamePlugin) Handle(ctx *hub.Context) error {
	if "same day" != ctx.Username || ctx.UID != "f1ed61fbef4e6a63" {
		slog.Error("Unauthorized user", "username", ctx.Username, "uid", ctx.UID)
		return nil
	}
	p.refreshImages()
	slog.Info("SamePlugin receive message", "type", ctx.MsgType, "content", ctx.Content)

	for cmd, handler := range handlers {
		if strings.HasPrefix(ctx.Content, cmd) {
			return handler(ctx, p)
		}
	}

	// currentEnv := os.Getenv("PLUGIN_ENV")
	// if "same" != currentEnv {
	// 	slog.Info("current env is not same, textToImage will not be called")
	// 	return nil
	// }

	return nil
}

func handleSame(ctx *hub.Context, p *SamePlugin) error {
	return ctx.Sender.SendText(ctx.GID, "hello same")
}

func handleSameSetu(ctx *hub.Context, p *SamePlugin) error {
	filePath := p.randomImage()
	slog.Info("handle same_setu", "file_path", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		slog.Error("Failed to open image", "error", err)
		return nil
	}
	return ctx.Sender.SendImg(ctx.GID, filePath, file)
}

func handleTxt2Img(ctx *hub.Context, p *SamePlugin) error {
	prompt := ctx.Content[8:]
	slog.Info("handle txt2img", "prompt", prompt)
	ctx.Sender.SendText(ctx.GID, "正在生成图片，请稍等")
	imagePath := p.textToImage(prompt)
	if imagePath == "" {
		slog.Error("Failed to generate image")
		return ctx.Sender.SendText(ctx.GID, "Failed to generate image")
	}
	slog.Info("handle txt2img", "imagePath", imagePath)
	file, err := os.Open(imagePath)
	if err != nil {
		slog.Error("Failed to open image", "error", err)
		return nil
	}
	return ctx.Sender.SendImg(ctx.GID, imagePath, file)
}

func handleCheckModel(ctx *hub.Context, p *SamePlugin) error {
	name := ctx.Content[13:]
	slog.Info("handle check_model", "name", name)
	if err := p.checkoutModel(name); err != nil {
		return ctx.Sender.SendText(ctx.GID, "Failed to check out model")
	}
	return ctx.Sender.SendText(ctx.GID, "Model checked out successfully")
}

func handleModel(ctx *hub.Context, p *SamePlugin) error {
	return ctx.Sender.SendText(ctx.GID, "当前模型："+p.Model)
}

func handleModelList(ctx *hub.Context, p *SamePlugin) error {
	models := model_list()
	modelsStr := "模型列表：\n"
	for _, model := range models {
		modelsStr += model + "\n"
	}
	return ctx.Sender.SendText(ctx.GID, modelsStr)
}
