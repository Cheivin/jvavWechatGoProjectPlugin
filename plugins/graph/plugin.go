package graph

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"fmt"
	"github.com/vicanso/go-charts/v2"
	"log/slog"
	"strconv"
	"strings"
	"time"
	"wechat-hub-plugin/hub"
)

//go:embed NotoSansCJKsc-VF.ttf
var fontBytes []byte

func init() {
	err := charts.InstallFont("noto", fontBytes)
	if err != nil {
		panic(err)
	}
	font, _ := charts.GetFont("noto")
	charts.SetDefaultFont(font)
}

type Plugin struct {
}

func (p Plugin) match(rawContent string) (content string, matched bool) {
	keywords := []string{
		"活跃度",
	}
	for _, keyword := range keywords {
		if strings.HasPrefix(rawContent, "#"+keyword) {
			matched = true
			content = strings.TrimPrefix(rawContent, keyword)
			return
		}
	}
	return
}
func (p Plugin) Handle(ctx *hub.Context) error {
	_, matched := p.match(ctx.Content)
	if !matched {
		return nil
	}
	now := time.Now()
	nowDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowDay := nowDay.Add(24 * time.Hour)
	last30Day := nowDay.Add(-30 * 24 * time.Hour)

	today, err := p.Today(ctx.DB, ctx.GID, ctx.UID, nowDay.Unix(), tomorrowDay.Unix())
	if err != nil {
		slog.Error("[活跃度]获取今日数据失败", "error", err)
		_ = ctx.ReplayText("[活跃度]获取今日数据失败")
		ctx.Abort()
		return nil
	}
	avgDay, err := p.AvgDay(ctx.DB, ctx.GID, ctx.UID, last30Day.Unix(), nowDay.Unix())
	if err != nil {
		slog.Error("[活跃度]获取近30天数据失败", "error", err)
		_ = ctx.ReplayText("[活跃度]获取近30天数据失败")
		ctx.Abort()
		return nil
	}
	img, err := p.Draw(ctx.Username, today, avgDay)
	if err != nil {
		slog.Error("[活跃度]生成图片失败", "error", err)
		_ = ctx.ReplayText("[活跃度]生成图片失败")
		ctx.Abort()
		return nil
	}
	if err := ctx.ReplayImg(fmt.Sprintf("%x.png", md5.Sum([]byte(ctx.Content+ctx.UID))), bytes.NewReader(img)); err != nil {
		slog.Error("[活跃度]上传图片失败", "error", err)
		_ = ctx.ReplayText("[活跃度]上传图片失败")
		ctx.Abort()
		return nil
	}
	return nil
}

type Statistic struct {
	Hour  string
	Total float64
}

func mapToStatistic(result []map[string]any) []Statistic {
	if result == nil {
		result = []map[string]any{}
	}
	statistics := map[string]Statistic{}
	for _, v := range result {
		vStr := fmt.Sprint(v["total"]) // 暴力转换
		val, _ := strconv.ParseFloat(vStr, 64)
		statistics[v["h"].(string)] = Statistic{
			Hour:  v["h"].(string),
			Total: val,
		}
	}

	list := make([]Statistic, 24)
	for i := 0; i < 24; i++ {
		hour := fmt.Sprintf("%02d", i)
		if _, ok := statistics[hour]; !ok {
			list[i] = Statistic{
				Hour: hour,
			}
			continue
		}
		list[i] = statistics[hour]
	}
	return list
}
func (p Plugin) Today(db hub.DBInterface, gid, uid string, startTime int64, endTime int64) ([]Statistic, error) {
	result, err := db.QueryAll("SELECT DATE_FORMAT(FROM_UNIXTIME(`time`),'%H') AS h,count(*) total FROM message WHERE  gid =? and uid=? and `time` >=? and `time` <? and COALESCE(JSON_VALUE(content, '$.content'),'') not REGEXP '^#' GROUP BY h", gid, uid, startTime, endTime)
	if err != nil {
		return nil, err
	}
	return mapToStatistic(result), nil
}

func (p Plugin) AvgDay(db hub.DBInterface, gid, uid string, startTime int64, endTime int64) ([]Statistic, error) {
	result, err := db.QueryAll("select h ,AVG(total) as total from (SELECT DATE_FORMAT(FROM_UNIXTIME(`time`),'%m-%d') AS d,DATE_FORMAT(FROM_UNIXTIME(`time`),'%H') AS h,count(*) total FROM message WHERE  gid =? and uid=? and `time`>=? and `time`<? and COALESCE(JSON_VALUE(content, '$.content'),'') not REGEXP '^#' GROUP BY d,h) t GROUP BY h", gid, uid, startTime, endTime)
	if err != nil {
		return nil, err
	}
	list := mapToStatistic(result)
	return list, nil
}

func (p Plugin) Draw(user string, nowActivity []Statistic, avgActivity []Statistic) ([]byte, error) {
	var values [][]float64
	var maxY float64 = 0

	var today []float64
	for _, v := range nowActivity {
		today = append(today, v.Total)
		if v.Total > maxY {
			maxY = v.Total
		}
	}
	var avg []float64
	for _, v := range avgActivity {
		avg = append(avg, v.Total)
		if v.Total > maxY {
			maxY = v.Total
		}
	}
	values = append(values, today, avg)

	var xAxis []string
	for i := 0; i < 24; i++ {
		xAxis = append(xAxis, fmt.Sprintf("%02d:00", i))
	}

	pa, err := charts.LineRender(
		values,
		charts.TitleTextOptionFunc(fmt.Sprintf("@%s活跃度", user)),
		charts.XAxisDataOptionFunc(xAxis),
		charts.YAxisOptionFunc(charts.YAxisOption{Max: &maxY, Show: charts.TrueFlag()}),
		charts.LegendLabelsOptionFunc([]string{
			"今日",
			"近30D平均",
		}, charts.PositionRight),
	)
	if err != nil {
		return nil, err
	}
	return pa.Bytes()
}
