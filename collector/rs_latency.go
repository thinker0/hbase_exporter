package collector

import (
	"fmt"
	"io/ioutil"
	"sync"

	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
	"hbase_exporter/utils"
)

var (
	defaultHBaseRsLatencyLabels = []string{"host", "role", "namespace", "htable"}
)

type hbaseLatencyJmxMetric struct {
	Namespace string
	Table     string
	Metric    string
	Value     float64
}

type RsLatency struct {
	logger log.Logger
	url    *url.URL

	metrics map[string]*prometheus.Desc
	mutex   sync.Mutex

	jmxs []*hbaseLatencyJmxMetric
}

func newMetric(metric string, doc string) *prometheus.Desc {
	subsystem := "latency"

	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, metric),
		doc,
		defaultHBaseRsLatencyLabels,
		nil,
	)
}

func NewRsLatency(logger log.Logger, url *url.URL) *RsLatency {
	return &RsLatency{
		logger: logger,
		url:    url,

		metrics: map[string]*prometheus.Desc{
			"deleteBatchTime_num_ops":              newMetric("deleteBatchTime_num_ops", "The description of deleteBatchTime_num_ops"),
			"deleteBatchTime_min":                  newMetric("deleteBatchTime_min", "The description of deleteBatchTime_min"),
			"deleteBatchTime_max":                  newMetric("deleteBatchTime_max", "The description of deleteBatchTime_max"),
			"deleteBatchTime_mean":                 newMetric("deleteBatchTime_mean", "The description of deleteBatchTime_mean"),
			"deleteBatchTime_25th_percentile":      newMetric("deleteBatchTime_25th_percentile", "The description of deleteBatchTime_25th_percentile"),
			"deleteBatchTime_median":               newMetric("deleteBatchTime_median", "The description of deleteBatchTime_median"),
			"deleteBatchTime_75th_percentile":      newMetric("deleteBatchTime_75th_percentile", "The description of deleteBatchTime_75th_percentile"),
			"deleteBatchTime_90th_percentile":      newMetric("deleteBatchTime_90th_percentile", "The description of deleteBatchTime_90th_percentile"),
			"deleteBatchTime_95th_percentile":      newMetric("deleteBatchTime_95th_percentile", "The description of deleteBatchTime_95th_percentile"),
			"deleteBatchTime_98th_percentile":      newMetric("deleteBatchTime_98th_percentile", "The description of deleteBatchTime_98th_percentile"),
			"deleteBatchTime_99th_percentile":      newMetric("deleteBatchTime_99th_percentile", "The description of deleteBatchTime_99th_percentile"),
			"deleteBatchTime_99.9th_percentile":    newMetric("deleteBatchTime_99.9th_percentile", "The description of deleteBatchTime_99.9th_percentile"),
			"appendTime_num_ops":                   newMetric("appendTime_num_ops", "The description of appendTime_num_ops"),
			"appendTime_min":                       newMetric("appendTime_min", "The description of appendTime_min"),
			"appendTime_max":                       newMetric("appendTime_max", "The description of appendTime_max"),
			"appendTime_mean":                      newMetric("appendTime_mean", "The description of appendTime_mean"),
			"appendTime_25th_percentile":           newMetric("appendTime_25th_percentile", "The description of appendTime_25th_percentile"),
			"appendTime_median":                    newMetric("appendTime_median", "The description of appendTime_median"),
			"appendTime_75th_percentile":           newMetric("appendTime_75th_percentile", "The description of appendTime_75th_percentile"),
			"appendTime_90th_percentile":           newMetric("appendTime_90th_percentile", "The description of appendTime_90th_percentile"),
			"appendTime_95th_percentile":           newMetric("appendTime_95th_percentile", "The description of appendTime_95th_percentile"),
			"appendTime_98th_percentile":           newMetric("appendTime_98th_percentile", "The description of appendTime_98th_percentile"),
			"appendTime_99th_percentile":           newMetric("appendTime_99th_percentile", "The description of appendTime_99th_percentile"),
			"appendTime_99.9th_percentile":         newMetric("appendTime_99.9th_percentile", "The description of appendTime_99.9th_percentile"),
			"getTime_num_ops":                      newMetric("getTime_num_ops", "The description of getTime_num_ops"),
			"getTime_min":                          newMetric("getTime_min", "The description of getTime_min"),
			"getTime_max":                          newMetric("getTime_max", "The description of getTime_max"),
			"getTime_mean":                         newMetric("getTime_mean", "The description of getTime_mean"),
			"getTime_25th_percentile":              newMetric("getTime_25th_percentile", "The description of getTime_25th_percentile"),
			"getTime_median":                       newMetric("getTime_median", "The description of getTime_median"),
			"getTime_75th_percentile":              newMetric("getTime_75th_percentile", "The description of getTime_75th_percentile"),
			"getTime_90th_percentile":              newMetric("getTime_90th_percentile", "The description of getTime_90th_percentile"),
			"getTime_95th_percentile":              newMetric("getTime_95th_percentile", "The description of getTime_95th_percentile"),
			"getTime_98th_percentile":              newMetric("getTime_98th_percentile", "The description of getTime_98th_percentile"),
			"getTime_99th_percentile":              newMetric("getTime_99th_percentile", "The description of getTime_99th_percentile"),
			"getTime_99.9th_percentile":            newMetric("getTime_99.9th_percentile", "The description of getTime_99.9th_percentile"),
			"incrementTime_num_ops":                newMetric("incrementTime_num_ops", "The description of incrementTime_num_ops"),
			"incrementTime_min":                    newMetric("incrementTime_min", "The description of incrementTime_min"),
			"incrementTime_max":                    newMetric("incrementTime_max", "The description of incrementTime_max"),
			"incrementTime_mean":                   newMetric("incrementTime_mean", "The description of incrementTime_mean"),
			"incrementTime_25th_percentile":        newMetric("incrementTime_25th_percentile", "The description of incrementTime_25th_percentile"),
			"incrementTime_median":                 newMetric("incrementTime_median", "The description of incrementTime_median"),
			"incrementTime_75th_percentile":        newMetric("incrementTime_75th_percentile", "The description of incrementTime_75th_percentile"),
			"incrementTime_90th_percentile":        newMetric("incrementTime_90th_percentile", "The description of incrementTime_90th_percentile"),
			"incrementTime_95th_percentile":        newMetric("incrementTime_95th_percentile", "The description of incrementTime_95th_percentile"),
			"incrementTime_98th_percentile":        newMetric("incrementTime_98th_percentile", "The description of incrementTime_98th_percentile"),
			"incrementTime_99th_percentile":        newMetric("incrementTime_99th_percentile", "The description of incrementTime_99th_percentile"),
			"incrementTime_99.9th_percentile":      newMetric("incrementTime_99.9th_percentile", "The description of incrementTime_99.9th_percentile"),
			"putBatchTime_num_ops":                 newMetric("putBatchTime_num_ops", "The description of putBatchTime_num_ops"),
			"putBatchTime_min":                     newMetric("putBatchTime_min", "The description of putBatchTime_min"),
			"putBatchTime_max":                     newMetric("putBatchTime_max", "The description of putBatchTime_max"),
			"putBatchTime_mean":                    newMetric("putBatchTime_mean", "The description of putBatchTime_mean"),
			"putBatchTime_25th_percentile":         newMetric("putBatchTime_25th_percentile", "The description of putBatchTime_25th_percentile"),
			"putBatchTime_median":                  newMetric("putBatchTime_median", "The description of putBatchTime_median"),
			"putBatchTime_75th_percentile":         newMetric("putBatchTime_75th_percentile", "The description of putBatchTime_75th_percentile"),
			"putBatchTime_90th_percentile":         newMetric("putBatchTime_90th_percentile", "The description of putBatchTime_90th_percentile"),
			"putBatchTime_95th_percentile":         newMetric("putBatchTime_95th_percentile", "The description of putBatchTime_95th_percentile"),
			"putBatchTime_98th_percentile":         newMetric("putBatchTime_98th_percentile", "The description of putBatchTime_98th_percentile"),
			"putBatchTime_99th_percentile":         newMetric("putBatchTime_99th_percentile", "The description of putBatchTime_99th_percentile"),
			"putBatchTime_99.9th_percentile":       newMetric("putBatchTime_99.9th_percentile", "The description of putBatchTime_99.9th_percentile"),
			"putTime_num_ops":                      newMetric("putTime_num_ops", "The description of putTime_num_ops"),
			"putTime_min":                          newMetric("putTime_min", "The description of putTime_min"),
			"putTime_max":                          newMetric("putTime_max", "The description of putTime_max"),
			"putTime_mean":                         newMetric("putTime_mean", "The description of putTime_mean"),
			"putTime_25th_percentile":              newMetric("putTime_25th_percentile", "The description of putTime_25th_percentile"),
			"putTime_median":                       newMetric("putTime_median", "The description of putTime_median"),
			"putTime_75th_percentile":              newMetric("putTime_75th_percentile", "The description of putTime_75th_percentile"),
			"putTime_90th_percentile":              newMetric("putTime_90th_percentile", "The description of putTime_90th_percentile"),
			"putTime_95th_percentile":              newMetric("putTime_95th_percentile", "The description of putTime_95th_percentile"),
			"putTime_98th_percentile":              newMetric("putTime_98th_percentile", "The description of putTime_98th_percentile"),
			"putTime_99th_percentile":              newMetric("putTime_99th_percentile", "The description of putTime_99th_percentile"),
			"putTime_99.9th_percentile":            newMetric("putTime_99.9th_percentile", "The description of putTime_99.9th_percentile"),
			"scanSize_num_ops":                     newMetric("scanSize_num_ops", "The description of scanSize_num_ops"),
			"scanSize_min":                         newMetric("scanSize_min", "The description of scanSize_min"),
			"scanSize_max":                         newMetric("scanSize_max", "The description of scanSize_max"),
			"scanSize_mean":                        newMetric("scanSize_mean", "The description of scanSize_mean"),
			"scanSize_25th_percentile":             newMetric("scanSize_25th_percentile", "The description of scanSize_25th_percentile"),
			"scanSize_median":                      newMetric("scanSize_median", "The description of scanSize_median"),
			"scanSize_75th_percentile":             newMetric("scanSize_75th_percentile", "The description of scanSize_75th_percentile"),
			"scanSize_90th_percentile":             newMetric("scanSize_90th_percentile", "The description of scanSize_90th_percentile"),
			"scanSize_95th_percentile":             newMetric("scanSize_95th_percentile", "The description of scanSize_95th_percentile"),
			"scanSize_98th_percentile":             newMetric("scanSize_98th_percentile", "The description of scanSize_98th_percentile"),
			"scanSize_99th_percentile":             newMetric("scanSize_99th_percentile", "The description of scanSize_99th_percentile"),
			"scanSize_99.9th_percentile":           newMetric("scanSize_99.9th_percentile", "The description of scanSize_99.9th_percentile"),
			"checkAndMutateTime_num_ops":           newMetric("checkAndMutateTime_num_ops", "The description of checkAndMutateTime_num_ops"),
			"checkAndMutateTime_min":               newMetric("checkAndMutateTime_min", "The description of checkAndMutateTime_min"),
			"checkAndMutateTime_max":               newMetric("checkAndMutateTime_max", "The description of checkAndMutateTime_max"),
			"checkAndMutateTime_mean":              newMetric("checkAndMutateTime_mean", "The description of checkAndMutateTime_mean"),
			"checkAndMutateTime_25th_percentile":   newMetric("checkAndMutateTime_25th_percentile", "The description of checkAndMutateTime_25th_percentile"),
			"checkAndMutateTime_median":            newMetric("checkAndMutateTime_median", "The description of checkAndMutateTime_median"),
			"checkAndMutateTime_75th_percentile":   newMetric("checkAndMutateTime_75th_percentile", "The description of checkAndMutateTime_75th_percentile"),
			"checkAndMutateTime_90th_percentile":   newMetric("checkAndMutateTime_90th_percentile", "The description of checkAndMutateTime_90th_percentile"),
			"checkAndMutateTime_95th_percentile":   newMetric("checkAndMutateTime_95th_percentile", "The description of checkAndMutateTime_95th_percentile"),
			"checkAndMutateTime_98th_percentile":   newMetric("checkAndMutateTime_98th_percentile", "The description of checkAndMutateTime_98th_percentile"),
			"checkAndMutateTime_99th_percentile":   newMetric("checkAndMutateTime_99th_percentile", "The description of checkAndMutateTime_99th_percentile"),
			"checkAndMutateTime_99.9th_percentile": newMetric("checkAndMutateTime_99.9th_percentile", "The description of checkAndMutateTime_99.9th_percentile"),
			"checkAndPutTime_num_ops":              newMetric("checkAndPutTime_num_ops", "The description of checkAndPutTime_num_ops"),
			"checkAndPutTime_min":                  newMetric("checkAndPutTime_min", "The description of checkAndPutTime_min"),
			"checkAndPutTime_max":                  newMetric("checkAndPutTime_max", "The description of checkAndPutTime_max"),
			"checkAndPutTime_mean":                 newMetric("checkAndPutTime_mean", "The description of checkAndPutTime_mean"),
			"checkAndPutTime_25th_percentile":      newMetric("checkAndPutTime_25th_percentile", "The description of checkAndPutTime_25th_percentile"),
			"checkAndPutTime_median":               newMetric("checkAndPutTime_median", "The description of checkAndPutTime_median"),
			"checkAndPutTime_75th_percentile":      newMetric("checkAndPutTime_75th_percentile", "The description of checkAndPutTime_75th_percentile"),
			"checkAndPutTime_90th_percentile":      newMetric("checkAndPutTime_90th_percentile", "The description of checkAndPutTime_90th_percentile"),
			"checkAndPutTime_95th_percentile":      newMetric("checkAndPutTime_95th_percentile", "The description of checkAndPutTime_95th_percentile"),
			"checkAndPutTime_98th_percentile":      newMetric("checkAndPutTime_98th_percentile", "The description of checkAndPutTime_98th_percentile"),
			"checkAndPutTime_99th_percentile":      newMetric("checkAndPutTime_99th_percentile", "The description of checkAndPutTime_99th_percentile"),
			"checkAndPutTime_99.9th_percentile":    newMetric("checkAndPutTime_99.9th_percentile", "The description of checkAndPutTime_99.9th_percentile"),
			"checkAndDeleteTime_num_ops":           newMetric("checkAndDeleteTime_num_ops", "The description of checkAndDeleteTime_num_ops"),
			"checkAndDeleteTime_min":               newMetric("checkAndDeleteTime_min", "The description of checkAndDeleteTime_min"),
			"checkAndDeleteTime_max":               newMetric("checkAndDeleteTime_max", "The description of checkAndDeleteTime_max"),
			"checkAndDeleteTime_mean":              newMetric("checkAndDeleteTime_mean", "The description of checkAndDeleteTime_mean"),
			"checkAndDeleteTime_25th_percentile":   newMetric("checkAndDeleteTime_25th_percentile", "The description of checkAndDeleteTime_25th_percentile"),
			"checkAndDeleteTime_median":            newMetric("checkAndDeleteTime_median", "The description of checkAndDeleteTime_median"),
			"checkAndDeleteTime_75th_percentile":   newMetric("checkAndDeleteTime_75th_percentile", "The description of checkAndDeleteTime_75th_percentile"),
			"checkAndDeleteTime_90th_percentile":   newMetric("checkAndDeleteTime_90th_percentile", "The description of checkAndDeleteTime_90th_percentile"),
			"checkAndDeleteTime_95th_percentile":   newMetric("checkAndDeleteTime_95th_percentile", "The description of checkAndDeleteTime_95th_percentile"),
			"checkAndDeleteTime_98th_percentile":   newMetric("checkAndDeleteTime_98th_percentile", "The description of checkAndDeleteTime_98th_percentile"),
			"checkAndDeleteTime_99th_percentile":   newMetric("checkAndDeleteTime_99th_percentile", "The description of checkAndDeleteTime_99th_percentile"),
			"checkAndDeleteTime_99.9th_percentile": newMetric("checkAndDeleteTime_99.9th_percentile", "The description of checkAndDeleteTime_99.9th_percentile"),
			"scanTime_num_ops":                     newMetric("scanTime_num_ops", "The description of scanTime_num_ops"),
			"scanTime_min":                         newMetric("scanTime_min", "The description of scanTime_min"),
			"scanTime_max":                         newMetric("scanTime_max", "The description of scanTime_max"),
			"scanTime_mean":                        newMetric("scanTime_mean", "The description of scanTime_mean"),
			"scanTime_25th_percentile":             newMetric("scanTime_25th_percentile", "The description of scanTime_25th_percentile"),
			"scanTime_median":                      newMetric("scanTime_median", "The description of scanTime_median"),
			"scanTime_75th_percentile":             newMetric("scanTime_75th_percentile", "The description of scanTime_75th_percentile"),
			"scanTime_90th_percentile":             newMetric("scanTime_90th_percentile", "The description of scanTime_90th_percentile"),
			"scanTime_95th_percentile":             newMetric("scanTime_95th_percentile", "The description of scanTime_95th_percentile"),
			"scanTime_98th_percentile":             newMetric("scanTime_98th_percentile", "The description of scanTime_98th_percentile"),
			"scanTime_99th_percentile":             newMetric("scanTime_99th_percentile", "The description of scanTime_99th_percentile"),
			"scanTime_99.9th_percentile":           newMetric("scanTime_99.9th_percentile", "The description of scanTime_99.9th_percentile"),
			"deleteTime_num_ops":                   newMetric("deleteTime_num_ops", "The description of deleteTime_num_ops"),
			"deleteTime_min":                       newMetric("deleteTime_min", "The description of deleteTime_min"),
			"deleteTime_max":                       newMetric("deleteTime_max", "The description of deleteTime_max"),
			"deleteTime_mean":                      newMetric("deleteTime_mean", "The description of deleteTime_mean"),
			"deleteTime_25th_percentile":           newMetric("deleteTime_25th_percentile", "The description of deleteTime_25th_percentile"),
			"deleteTime_median":                    newMetric("deleteTime_median", "The description of deleteTime_median"),
			"deleteTime_75th_percentile":           newMetric("deleteTime_75th_percentile", "The description of deleteTime_75th_percentile"),
			"deleteTime_90th_percentile":           newMetric("deleteTime_90th_percentile", "The description of deleteTime_90th_percentile"),
			"deleteTime_95th_percentile":           newMetric("deleteTime_95th_percentile", "The description of deleteTime_95th_percentile"),
			"deleteTime_98th_percentile":           newMetric("deleteTime_98th_percentile", "The description of deleteTime_98th_percentile"),
			"deleteTime_99th_percentile":           newMetric("deleteTime_99th_percentile", "The description of deleteTime_99th_percentile"),
			"deleteTime_99.9th_percentile":         newMetric("deleteTime_99.9th_percentile", "The description of deleteTime_99.9th_percentile"),
			"tableWriteQueryPerSecond_count":       newMetric("tableWriteQueryPerSecond_count", "The description of tableWriteQueryPerSecond_count"),
			"tableWriteQueryPerSecond_mean_rate":   newMetric("tableWriteQueryPerSecond_mean_rate", "The description of tableWriteQueryPerSecond_mean_rate"),
			"tableWriteQueryPerSecond_1min_rate":   newMetric("tableWriteQueryPerSecond_1min_rate", "The description of tableWriteQueryPerSecond_1min_rate"),
			"tableWriteQueryPerSecond_5min_rate":   newMetric("tableWriteQueryPerSecond_5min_rate", "The description of tableWriteQueryPerSecond_5min_rate"),
			"tableWriteQueryPerSecond_15min_rate":  newMetric("tableWriteQueryPerSecond_15min_rate", "The description of tableWriteQueryPerSecond_15min_rate"),
			"tableReadQueryPerSecond_count":        newMetric("tableReadQueryPerSecond_count", "The description of tableReadQueryPerSecond_count"),
			"tableReadQueryPerSecond_mean_rate":    newMetric("tableReadQueryPerSecond_mean_rate", "The description of tableReadQueryPerSecond_mean_rate"),
			"tableReadQueryPerSecond_1min_rate":    newMetric("tableReadQueryPerSecond_1min_rate", "The description of tableReadQueryPerSecond_1min_rate"),
			"tableReadQueryPerSecond_5min_rate":    newMetric("tableReadQueryPerSecond_5min_rate", "The description of tableReadQueryPerSecond_5min_rate"),
			"tableReadQueryPerSecond_15min_rate":   newMetric("tableReadQueryPerSecond_15min_rate", "The description of tableReadQueryPerSecond_15min_rate"),
		},

		jmxs: []*hbaseLatencyJmxMetric{},
	}
}

func (m *RsLatency) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric
	}
}

func (r *RsLatency) fetchAndDecodeRsLatency() (string, string, error) {
	u := *r.url
	r.jmxs = r.jmxs[0:0]
	url := u.String() + "?" + "qry=Hadoop:service=HBase,name=RegionServer,sub=TableLatencies"
	res, err := http.Get(url)

	if err != nil {
		return "", "", fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(r.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	data := gjson.Get(string(bts), "beans").Array()[0].Map()

	host := data["tag.Hostname"].String()
	role := data["tag.Context"].String()

	for k, v := range data {
		if strings.HasPrefix(k, "Namespace") {
			keys := utils.SplitHBaseLatencyStr(k)
			values := v.Float()

			r.jmxs = append(r.jmxs, &hbaseLatencyJmxMetric{
				keys[0],
				keys[1],
				keys[2],
				values,
			})

		}
	}

	return host, role, nil

}

func (r *RsLatency) Collect(ch chan<- prometheus.Metric) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var err error

	host, role, err := r.fetchAndDecodeRsLatency()

	if err != nil {
		_ = level.Warn(r.logger).Log(
			"msg", "failed to fetch and decode cluster health",
			"err", err,
		)
		return
	}

	for _, jmx := range r.jmxs {
		Type := prometheus.GaugeValue
		Value := jmx.Value
		Labels := []string{host, role, jmx.Namespace, jmx.Table}

		if jmx.Metric == "deleteBatchTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_num_ops"),
					"The number of deleteBatchTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "deleteBatchTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_min"),
					"The number of deleteBatchTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		}
	}
}
