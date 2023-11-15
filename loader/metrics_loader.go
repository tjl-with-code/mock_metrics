package loader

import (
	"bufio"
	"bytes"
	"fake-metrics/metric_data"
	"fake-metrics/utils"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type ParseReq struct {
	IsReset bool   `json:"is_reset,omitempty"`
	Text    string `json:"text,omitempty"`
}

type MetricsLoader struct {
	cache       *cache.Cache
	contextList utils.Set
}

func NewMetricsLoader(cache *cache.Cache, contextList utils.Set) MetricsLoader {
	return MetricsLoader{
		cache:       cache,
		contextList: contextList,
	}
}

// 从文本解析指标
func (m *MetricsLoader) ParseText(c *gin.Context) {
	var jsonReq ParseReq
	err := c.BindJSON(&jsonReq)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.String(http.StatusInternalServerError, "%s", err)
		return
	}
	reader := bytes.NewBufferString(jsonReq.Text)
	m.loadAndMerge(c, reader, jsonReq.IsReset)
}

// 从特定的url加载指标
func (m *MetricsLoader) ParseFromReq(c *gin.Context) {
	var jsonReq ParseReq
	err := c.BindJSON(&jsonReq)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		c.String(http.StatusInternalServerError, "%s", err)
		return
	}

	reader := m.reqUrl(c, jsonReq.Text)
	defer reader.Close()

	m.loadAndMerge(c, reader, jsonReq.IsReset)
}

func (m *MetricsLoader) reqUrl(c *gin.Context, url string) io.ReadCloser {
	r, err := http.Get(url)
	if err != nil {
		c.String(http.StatusOK, "请求发生错误:%s", err)
		return nil
	}
	return r.Body
}

// 解析指标并且合并
func (m *MetricsLoader) loadAndMerge(c *gin.Context, reader io.Reader, isReset bool) metric_data.InstantResList {
	metricsKey := c.GetString("context")

	newMetricsMap, err := m.loadMetrics(reader)
	if err != nil {
		c.String(http.StatusOK, "解析指标错误,%s", err)
		return nil
	}

	if !isReset {
		oldMetricsMap, ok := m.cache.Get(metricsKey)
		if ok {
			newMetricsMap = m.mergeMetrics(newMetricsMap, oldMetricsMap.(metric_data.InstantResList))
		}
	}
	m.cache.Set(metricsKey, newMetricsMap, -1)
	if metricsKey != "metrics" {
		m.contextList.Add(metricsKey)
	}
	c.String(http.StatusOK, "解析成功，共得到%d个指标", len(newMetricsMap))
	return newMetricsMap
}

func (m *MetricsLoader) loadMetrics(reader io.Reader) (metric_data.InstantResList, error) {
	me := metric_data.NewTextParser()
	return me.Encode(reader)
}

func (m *MetricsLoader) TestFetchUrl(c *gin.Context) {
	url, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusOK, "错误:%s", err)
		return
	}

	reader := m.reqUrl(c, utils.YoloString(url))
	defer reader.Close()
	br := bufio.NewScanner(reader)
	bf := bytes.NewBuffer([]byte{})

	for i := 5; i > 0; i-- {
		if !br.Scan() {
			break
		}
		bf.WriteString(fmt.Sprintf("%s<br> ", br.Text()))
	}

	c.String(http.StatusOK, "%s", bf.String())
}

// 将新旧数据合并，如果新旧数据中存在label相同的metrics，则用新数据的value更新旧数据的value
func (m *MetricsLoader) mergeMetrics(newMetrics, oldMetrics metric_data.InstantResList) metric_data.InstantResList {
	newMap := newMetrics.ToFilterMap()
	oldMap := oldMetrics.ToFilterMap()

	var res metric_data.InstantResList

	for key, value := range newMap {
		// 如果旧数据中存在相同label,用新的value更新旧的value
		if old, ok := oldMap[key]; ok {
			old.Value = value.Value
		} else {
			// 不存在则直接追加
			res = append(res, value)
		}
	}

	for _, value := range oldMap {
		res = append(res, value)
	}
	return res
}
