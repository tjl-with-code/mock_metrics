package metric_data

import (
	"bytes"
	"encoding/json"
	"fake-metrics/utils"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/prometheus/common/expfmt"
)

func TestParseText(t *testing.T) {
	tp := textParser{}
	buf := bytes.NewReader([]byte(`
node_mem{a="1"} 2
node_mem{a="2"} 3
node_mem{a="3"} 4
node_mem{a="4"} 5
node_mem{a="5"} 9
node_cpu{a="1"} 6
	`))
	s, _ := tp.Encode(buf)
	for _, item := range s {
		fmt.Println(item)
	}
	mf := s.ToMetricsFM()
	b, err := json.MarshalIndent(mf, "", "	")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(b))

	b2, err2 := json.MarshalIndent(s, "", "\t")
	if err2 != nil {
		log.Panic(err2)
	}
	fmt.Printf("string(b2): %v\n", utils.YoloString(b2))

	e := expfmt.NewEncoder(os.Stdout, expfmt.FmtText)
	for _, item := range mf {
		err3 := e.Encode(item)
		if err3 != nil {
			log.Panic(err3)
		}
	}
}
