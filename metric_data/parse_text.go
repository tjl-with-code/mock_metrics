package metric_data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sort"

	"github.com/grafana/regexp"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/prometheus/prometheus/model/textparse"
	"github.com/prometheus/prometheus/promql"
)

type InstantRes struct {
	Id     int           `json:"id,omitempty"`
	Value  float64       `json:"value,omitempty"`
	Labels labels.Labels `json:"labels,omitempty"`
	Name   string        `json:"name,omitempty"`
}

func (i *InstantRes) Hash() uint64 {
	return i.Labels.Hash()
}

func (i *InstantRes) ToMetric() *io_prometheus_client.Metric {
	ls := make([]*io_prometheus_client.LabelPair, len(i.Labels.WithoutLabels(labels.MetricName)))
	for index, l := range i.Labels.WithoutLabels(labels.MetricName) {
		temp := l
		ls[index] = &io_prometheus_client.LabelPair{
			Name:  &temp.Name,
			Value: &temp.Value,
		}
	}
	return &io_prometheus_client.Metric{
		Label: ls,
		Counter: &io_prometheus_client.Counter{
			Value: &i.Value,
		},
	}
}

func (i *InstantRes) MarshalJSON() ([]byte, error) {
	type tempLabel struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	}
	tt := i.Labels.WithoutLabels(labels.MetricName)
	tempLabels := make([]tempLabel, len(tt))
	for index, item := range tt {
		tempLabels[index] = tempLabel{
			Name:  item.Name,
			Value: item.Value,
		}
	}

	temp := struct {
		Id     int         `json:"id,omitempty"`
		Value  float64     `json:"value,omitempty,string"`
		Labels []tempLabel `json:"labels,omitempty"`
		Name   string      `json:"name,omitempty"`
	}{
		Id:     i.Id,
		Value:  i.Value,
		Labels: tempLabels,
		Name:   i.Name,
	}
	return json.Marshal(temp)
}

func (i *InstantRes) UnmarshalJSON(buf []byte) error {
	type tempLabel struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	}
	temp := struct {
		Id     int         `json:"id,omitempty"`
		Value  float64     `json:"value,omitempty,string"`
		Labels []tempLabel `json:"labels,omitempty"`
		Name   string      `json:"name,omitempty"`
	}{}
	err := json.Unmarshal(buf, &temp)
	if err != nil {
		return err
	}

	i.Id = temp.Id
	i.Name = temp.Name
	i.Value = temp.Value
	ls := make(labels.Labels, len(temp.Labels))

	for index, item := range temp.Labels {
		ls[index] = labels.Label{
			Name:  item.Name,
			Value: item.Value,
		}
	}
	i.Labels = ls

	return err
}

type InstantResList []*InstantRes

func (rl InstantResList) ToMetricsFM() []*io_prometheus_client.MetricFamily {
	sort.Slice(rl, func(i, j int) bool {
		return rl[i].Name < rl[j].Name
	})

	res := []*io_prometheus_client.MetricFamily{}

	thisName := ""
	var tempFm *io_prometheus_client.MetricFamily
	for _, item := range rl {
		temp := item
		if item.Name != thisName {
			if tempFm != nil {
				res = append(res, tempFm)
			}
			tempFm = &io_prometheus_client.MetricFamily{
				Name: &temp.Name,
			}
			thisName = item.Name
		}
		tempFm.Metric = append(tempFm.Metric, item.ToMetric())
	}
	if tempFm != nil {
		res = append(res, tempFm)
	}
	return res
}
func (rl InstantResList) ToFilterMap() map[uint64]*InstantRes {
	res := make(map[uint64]*InstantRes)

	for _, item := range rl {
		res[item.Hash()] = item
	}

	return res
}

type MetricsEncode interface {
	// 从reader读取指标将其填充到Sample中
	Encode(reader io.Reader) (InstantResList, error)
}

type textParser struct {
}

func NewTextParser() MetricsEncode {
	return &textParser{}
}

func (t *textParser) Encode(reader io.Reader) (InstantResList, error) {
	var err error
	res := make(InstantResList, 0)
	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, reader)
	if err != nil {
		log.Printf("copy error %s", err)
		return nil, err
	}
	p := textparse.NewPromParser(buf.Bytes())

	for {
		e, err := p.Next()
		if err != nil {
			if err == io.EOF {
				err = nil
				log.Println("end parse")
				break
			}
			log.Printf("parse metrics error %s", err)
			return nil, err
		}
		switch e {
		case textparse.EntrySeries:
			_, _, f := p.Series()
			var l labels.Labels
			p.Metric(&l)
			res = append(res, &InstantRes{
				Value:  f,
				Labels: l,
				Name:   l.Get(labels.MetricName),
			})
		}
	}

	return res, err
}

func ParseText(reader io.Reader) string {
	rConfig := relabel.Config{
		SourceLabels: []model.LabelName{
			"a",
		},
		Regex: relabel.Regexp{
			Regexp: regexp.MustCompile(`(\d+)`),
		},
		TargetLabel: "b",
		Replacement: "bb$1",
		Action:      relabel.Replace,
	}
	input := `
node_mem{a="aaaa1"} 12321 12312312
	`
	p := textparse.NewPromParser([]byte(input))
	e, err := p.Next()
	if err != nil {
		log.Panic(err)
	}
	switch e {
	case textparse.EntrySeries:
		m, ts, v := p.Series()
		fmt.Println(m, ts, v)
		var l labels.Labels
		s := p.Metric(&l)
		fmt.Println(s)
		fmt.Println(l)
		ss := promql.Sample{
			Point: promql.Point{
				T: *ts,
				V: v,
			},
			Metric: l,
		}
		fmt.Printf("ss: %v\n", ss)
		l = relabel.Process(l, &rConfig)
		fmt.Printf("l: %v\n", l)
	}
	return ""
}
