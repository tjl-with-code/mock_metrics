package main

import (
	"fake-metrics/loader"
	"fake-metrics/metric_data"
	"fake-metrics/utils"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/prometheus/common/expfmt"
)

type ParseReq struct {
	IsReset bool   `json:"is_reset,omitempty"`
	Text    string `json:"text,omitempty"`
}

func ContextFromPath() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		context := ctx.Param("context")
		if context == "default" || context == "/" {
			context = "metrics"
		}
		context = strings.Trim(context, "/")
		log.Printf("current ctx is %s.", context)
		ctx.Set("context", context)
		ctx.Next()
	}
}

func main() {
	tempCache := cache.New(5*time.Minute, 10*time.Minute)
	contextList := make(utils.Set)
	router := gin.Default()
	router.Static("/static/", "./static/js/")
	router.StaticFile("/favicon.ico", "./static/favicon.png")
	router.Delims("{[{", "}]}")
	router.LoadHTMLGlob("./static/template/**/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	metricsLoader := loader.NewMetricsLoader(tempCache, contextList)

	metricGroup := router.Group("/metric/:context", ContextFromPath())

	{
		metricGroup.POST("/parse", metricsLoader.ParseText)
		// 尝试从url读取metrics
		metricGroup.POST("/testQuery", metricsLoader.TestFetchUrl)
		// 从url加载metrics
		metricGroup.POST("/Query", metricsLoader.ParseFromReq)

		// 页面的可视化修改指标时需要的数据
		metricGroup.GET("/get", func(c *gin.Context) {
			context := c.GetString("context")

			metricsItem, ok := tempCache.Get(context)
			if !ok {
				c.String(http.StatusOK, "%s", "")
				return
			}
			if item, ok := metricsItem.(metric_data.InstantResList); ok && len(item) != 0 {
				c.JSON(http.StatusOK, item)
			}
		})
		metricGroup.POST("/put", func(c *gin.Context) {
			contextKey := c.GetString("context")
			var metricItems metric_data.InstantResList
			err := c.BindJSON(&metricItems)
			if err != nil {
				log.Panic(err)
				c.String(http.StatusOK, "%s", err)
				return
			}
			tempCache.Set(contextKey, metricItems, -1)
			c.JSON(http.StatusOK, metricItems)
		})
	}
	router.GET("/context/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, contextList.List())
	})
	// 解析指标
	router.GET("/metrics/*context", ContextFromPath(), func(c *gin.Context) {
		contextKey := c.GetString("context")
		metricsItem, ok := tempCache.Get(contextKey)
		if !ok {
			c.String(http.StatusOK, "%s", "")
			return
		}

		c.Stream(func(w io.Writer) bool {
			enc := expfmt.NewEncoder(w, expfmt.FmtText)
			if item, ok := metricsItem.(metric_data.InstantResList); ok && len(item) != 0 {
				mf := item.ToMetricsFM()
				for _, vv := range mf {
					enc.Encode(vv)
				}
			}
			return false
		})
	})
	router.Run(":8080")
}
