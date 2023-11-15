package loader

import (
	"bytes"
	"encoding/json"
	"fake-metrics/utils"
	"fmt"
	"testing"
	"time"

	"github.com/patrickmn/go-cache"
)

func TestMetricsLoader_mergeMetrics(t *testing.T) {
	cache := cache.New(5*time.Second, 100*time.Second)
	ss := make(utils.Set)
	ml := NewMetricsLoader(cache, ss)
	input1 := `
node_mem{a="1"} 2
	`
	input2 := `
node_mem{a="4"} 3
	`
	irl, _ := ml.loadMetrics(bytes.NewBufferString(input1))
	irl2, _ := ml.loadMetrics(bytes.NewBufferString(input2))
	fmt.Println(irl.ToFilterMap())
	fmt.Println(irl2.ToFilterMap())

	irl3 := ml.mergeMetrics(irl, irl2)
	b, _ := json.MarshalIndent(irl3, "", " ")
	fmt.Printf("string(b): %v\n", string(b))
}
