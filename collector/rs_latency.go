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

func newLatencyMetric(metric string, doc string) *prometheus.Desc {
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
			"deleteBatchTime_num_ops":              newLatencyMetric("deleteBatchTime_num_ops", "The description of deleteBatchTime_num_ops"),
			"deleteBatchTime_min":                  newLatencyMetric("deleteBatchTime_min", "The description of deleteBatchTime_min"),
			"deleteBatchTime_max":                  newLatencyMetric("deleteBatchTime_max", "The description of deleteBatchTime_max"),
			"deleteBatchTime_mean":                 newLatencyMetric("deleteBatchTime_mean", "The description of deleteBatchTime_mean"),
			"deleteBatchTime_25th_percentile":      newLatencyMetric("deleteBatchTime_25th_percentile", "The description of deleteBatchTime_25th_percentile"),
			"deleteBatchTime_median":               newLatencyMetric("deleteBatchTime_median", "The description of deleteBatchTime_median"),
			"deleteBatchTime_75th_percentile":      newLatencyMetric("deleteBatchTime_75th_percentile", "The description of deleteBatchTime_75th_percentile"),
			"deleteBatchTime_90th_percentile":      newLatencyMetric("deleteBatchTime_90th_percentile", "The description of deleteBatchTime_90th_percentile"),
			"deleteBatchTime_95th_percentile":      newLatencyMetric("deleteBatchTime_95th_percentile", "The description of deleteBatchTime_95th_percentile"),
			"deleteBatchTime_98th_percentile":      newLatencyMetric("deleteBatchTime_98th_percentile", "The description of deleteBatchTime_98th_percentile"),
			"deleteBatchTime_99th_percentile":      newLatencyMetric("deleteBatchTime_99th_percentile", "The description of deleteBatchTime_99th_percentile"),
			"deleteBatchTime_99.9th_percentile":    newLatencyMetric("deleteBatchTime_99.9th_percentile", "The description of deleteBatchTime_99.9th_percentile"),
			"appendTime_num_ops":                   newLatencyMetric("appendTime_num_ops", "The description of appendTime_num_ops"),
			"appendTime_min":                       newLatencyMetric("appendTime_min", "The description of appendTime_min"),
			"appendTime_max":                       newLatencyMetric("appendTime_max", "The description of appendTime_max"),
			"appendTime_mean":                      newLatencyMetric("appendTime_mean", "The description of appendTime_mean"),
			"appendTime_25th_percentile":           newLatencyMetric("appendTime_25th_percentile", "The description of appendTime_25th_percentile"),
			"appendTime_median":                    newLatencyMetric("appendTime_median", "The description of appendTime_median"),
			"appendTime_75th_percentile":           newLatencyMetric("appendTime_75th_percentile", "The description of appendTime_75th_percentile"),
			"appendTime_90th_percentile":           newLatencyMetric("appendTime_90th_percentile", "The description of appendTime_90th_percentile"),
			"appendTime_95th_percentile":           newLatencyMetric("appendTime_95th_percentile", "The description of appendTime_95th_percentile"),
			"appendTime_98th_percentile":           newLatencyMetric("appendTime_98th_percentile", "The description of appendTime_98th_percentile"),
			"appendTime_99th_percentile":           newLatencyMetric("appendTime_99th_percentile", "The description of appendTime_99th_percentile"),
			"appendTime_99.9th_percentile":         newLatencyMetric("appendTime_99.9th_percentile", "The description of appendTime_99.9th_percentile"),
			"getTime_num_ops":                      newLatencyMetric("getTime_num_ops", "The description of getTime_num_ops"),
			"getTime_min":                          newLatencyMetric("getTime_min", "The description of getTime_min"),
			"getTime_max":                          newLatencyMetric("getTime_max", "The description of getTime_max"),
			"getTime_mean":                         newLatencyMetric("getTime_mean", "The description of getTime_mean"),
			"getTime_25th_percentile":              newLatencyMetric("getTime_25th_percentile", "The description of getTime_25th_percentile"),
			"getTime_median":                       newLatencyMetric("getTime_median", "The description of getTime_median"),
			"getTime_75th_percentile":              newLatencyMetric("getTime_75th_percentile", "The description of getTime_75th_percentile"),
			"getTime_90th_percentile":              newLatencyMetric("getTime_90th_percentile", "The description of getTime_90th_percentile"),
			"getTime_95th_percentile":              newLatencyMetric("getTime_95th_percentile", "The description of getTime_95th_percentile"),
			"getTime_98th_percentile":              newLatencyMetric("getTime_98th_percentile", "The description of getTime_98th_percentile"),
			"getTime_99th_percentile":              newLatencyMetric("getTime_99th_percentile", "The description of getTime_99th_percentile"),
			"getTime_99.9th_percentile":            newLatencyMetric("getTime_99.9th_percentile", "The description of getTime_99.9th_percentile"),
			"incrementTime_num_ops":                newLatencyMetric("incrementTime_num_ops", "The description of incrementTime_num_ops"),
			"incrementTime_min":                    newLatencyMetric("incrementTime_min", "The description of incrementTime_min"),
			"incrementTime_max":                    newLatencyMetric("incrementTime_max", "The description of incrementTime_max"),
			"incrementTime_mean":                   newLatencyMetric("incrementTime_mean", "The description of incrementTime_mean"),
			"incrementTime_25th_percentile":        newLatencyMetric("incrementTime_25th_percentile", "The description of incrementTime_25th_percentile"),
			"incrementTime_median":                 newLatencyMetric("incrementTime_median", "The description of incrementTime_median"),
			"incrementTime_75th_percentile":        newLatencyMetric("incrementTime_75th_percentile", "The description of incrementTime_75th_percentile"),
			"incrementTime_90th_percentile":        newLatencyMetric("incrementTime_90th_percentile", "The description of incrementTime_90th_percentile"),
			"incrementTime_95th_percentile":        newLatencyMetric("incrementTime_95th_percentile", "The description of incrementTime_95th_percentile"),
			"incrementTime_98th_percentile":        newLatencyMetric("incrementTime_98th_percentile", "The description of incrementTime_98th_percentile"),
			"incrementTime_99th_percentile":        newLatencyMetric("incrementTime_99th_percentile", "The description of incrementTime_99th_percentile"),
			"incrementTime_99.9th_percentile":      newLatencyMetric("incrementTime_99.9th_percentile", "The description of incrementTime_99.9th_percentile"),
			"putBatchTime_num_ops":                 newLatencyMetric("putBatchTime_num_ops", "The description of putBatchTime_num_ops"),
			"putBatchTime_min":                     newLatencyMetric("putBatchTime_min", "The description of putBatchTime_min"),
			"putBatchTime_max":                     newLatencyMetric("putBatchTime_max", "The description of putBatchTime_max"),
			"putBatchTime_mean":                    newLatencyMetric("putBatchTime_mean", "The description of putBatchTime_mean"),
			"putBatchTime_25th_percentile":         newLatencyMetric("putBatchTime_25th_percentile", "The description of putBatchTime_25th_percentile"),
			"putBatchTime_median":                  newLatencyMetric("putBatchTime_median", "The description of putBatchTime_median"),
			"putBatchTime_75th_percentile":         newLatencyMetric("putBatchTime_75th_percentile", "The description of putBatchTime_75th_percentile"),
			"putBatchTime_90th_percentile":         newLatencyMetric("putBatchTime_90th_percentile", "The description of putBatchTime_90th_percentile"),
			"putBatchTime_95th_percentile":         newLatencyMetric("putBatchTime_95th_percentile", "The description of putBatchTime_95th_percentile"),
			"putBatchTime_98th_percentile":         newLatencyMetric("putBatchTime_98th_percentile", "The description of putBatchTime_98th_percentile"),
			"putBatchTime_99th_percentile":         newLatencyMetric("putBatchTime_99th_percentile", "The description of putBatchTime_99th_percentile"),
			"putBatchTime_99.9th_percentile":       newLatencyMetric("putBatchTime_99.9th_percentile", "The description of putBatchTime_99.9th_percentile"),
			"putTime_num_ops":                      newLatencyMetric("putTime_num_ops", "The description of putTime_num_ops"),
			"putTime_min":                          newLatencyMetric("putTime_min", "The description of putTime_min"),
			"putTime_max":                          newLatencyMetric("putTime_max", "The description of putTime_max"),
			"putTime_mean":                         newLatencyMetric("putTime_mean", "The description of putTime_mean"),
			"putTime_25th_percentile":              newLatencyMetric("putTime_25th_percentile", "The description of putTime_25th_percentile"),
			"putTime_median":                       newLatencyMetric("putTime_median", "The description of putTime_median"),
			"putTime_75th_percentile":              newLatencyMetric("putTime_75th_percentile", "The description of putTime_75th_percentile"),
			"putTime_90th_percentile":              newLatencyMetric("putTime_90th_percentile", "The description of putTime_90th_percentile"),
			"putTime_95th_percentile":              newLatencyMetric("putTime_95th_percentile", "The description of putTime_95th_percentile"),
			"putTime_98th_percentile":              newLatencyMetric("putTime_98th_percentile", "The description of putTime_98th_percentile"),
			"putTime_99th_percentile":              newLatencyMetric("putTime_99th_percentile", "The description of putTime_99th_percentile"),
			"putTime_99.9th_percentile":            newLatencyMetric("putTime_99.9th_percentile", "The description of putTime_99.9th_percentile"),
			"scanSize_num_ops":                     newLatencyMetric("scanSize_num_ops", "The description of scanSize_num_ops"),
			"scanSize_min":                         newLatencyMetric("scanSize_min", "The description of scanSize_min"),
			"scanSize_max":                         newLatencyMetric("scanSize_max", "The description of scanSize_max"),
			"scanSize_mean":                        newLatencyMetric("scanSize_mean", "The description of scanSize_mean"),
			"scanSize_25th_percentile":             newLatencyMetric("scanSize_25th_percentile", "The description of scanSize_25th_percentile"),
			"scanSize_median":                      newLatencyMetric("scanSize_median", "The description of scanSize_median"),
			"scanSize_75th_percentile":             newLatencyMetric("scanSize_75th_percentile", "The description of scanSize_75th_percentile"),
			"scanSize_90th_percentile":             newLatencyMetric("scanSize_90th_percentile", "The description of scanSize_90th_percentile"),
			"scanSize_95th_percentile":             newLatencyMetric("scanSize_95th_percentile", "The description of scanSize_95th_percentile"),
			"scanSize_98th_percentile":             newLatencyMetric("scanSize_98th_percentile", "The description of scanSize_98th_percentile"),
			"scanSize_99th_percentile":             newLatencyMetric("scanSize_99th_percentile", "The description of scanSize_99th_percentile"),
			"scanSize_99.9th_percentile":           newLatencyMetric("scanSize_99.9th_percentile", "The description of scanSize_99.9th_percentile"),
			"checkAndMutateTime_num_ops":           newLatencyMetric("checkAndMutateTime_num_ops", "The description of checkAndMutateTime_num_ops"),
			"checkAndMutateTime_min":               newLatencyMetric("checkAndMutateTime_min", "The description of checkAndMutateTime_min"),
			"checkAndMutateTime_max":               newLatencyMetric("checkAndMutateTime_max", "The description of checkAndMutateTime_max"),
			"checkAndMutateTime_mean":              newLatencyMetric("checkAndMutateTime_mean", "The description of checkAndMutateTime_mean"),
			"checkAndMutateTime_25th_percentile":   newLatencyMetric("checkAndMutateTime_25th_percentile", "The description of checkAndMutateTime_25th_percentile"),
			"checkAndMutateTime_median":            newLatencyMetric("checkAndMutateTime_median", "The description of checkAndMutateTime_median"),
			"checkAndMutateTime_75th_percentile":   newLatencyMetric("checkAndMutateTime_75th_percentile", "The description of checkAndMutateTime_75th_percentile"),
			"checkAndMutateTime_90th_percentile":   newLatencyMetric("checkAndMutateTime_90th_percentile", "The description of checkAndMutateTime_90th_percentile"),
			"checkAndMutateTime_95th_percentile":   newLatencyMetric("checkAndMutateTime_95th_percentile", "The description of checkAndMutateTime_95th_percentile"),
			"checkAndMutateTime_98th_percentile":   newLatencyMetric("checkAndMutateTime_98th_percentile", "The description of checkAndMutateTime_98th_percentile"),
			"checkAndMutateTime_99th_percentile":   newLatencyMetric("checkAndMutateTime_99th_percentile", "The description of checkAndMutateTime_99th_percentile"),
			"checkAndMutateTime_99.9th_percentile": newLatencyMetric("checkAndMutateTime_99.9th_percentile", "The description of checkAndMutateTime_99.9th_percentile"),
			"checkAndPutTime_num_ops":              newLatencyMetric("checkAndPutTime_num_ops", "The description of checkAndPutTime_num_ops"),
			"checkAndPutTime_min":                  newLatencyMetric("checkAndPutTime_min", "The description of checkAndPutTime_min"),
			"checkAndPutTime_max":                  newLatencyMetric("checkAndPutTime_max", "The description of checkAndPutTime_max"),
			"checkAndPutTime_mean":                 newLatencyMetric("checkAndPutTime_mean", "The description of checkAndPutTime_mean"),
			"checkAndPutTime_25th_percentile":      newLatencyMetric("checkAndPutTime_25th_percentile", "The description of checkAndPutTime_25th_percentile"),
			"checkAndPutTime_median":               newLatencyMetric("checkAndPutTime_median", "The description of checkAndPutTime_median"),
			"checkAndPutTime_75th_percentile":      newLatencyMetric("checkAndPutTime_75th_percentile", "The description of checkAndPutTime_75th_percentile"),
			"checkAndPutTime_90th_percentile":      newLatencyMetric("checkAndPutTime_90th_percentile", "The description of checkAndPutTime_90th_percentile"),
			"checkAndPutTime_95th_percentile":      newLatencyMetric("checkAndPutTime_95th_percentile", "The description of checkAndPutTime_95th_percentile"),
			"checkAndPutTime_98th_percentile":      newLatencyMetric("checkAndPutTime_98th_percentile", "The description of checkAndPutTime_98th_percentile"),
			"checkAndPutTime_99th_percentile":      newLatencyMetric("checkAndPutTime_99th_percentile", "The description of checkAndPutTime_99th_percentile"),
			"checkAndPutTime_99.9th_percentile":    newLatencyMetric("checkAndPutTime_99.9th_percentile", "The description of checkAndPutTime_99.9th_percentile"),
			"checkAndDeleteTime_num_ops":           newLatencyMetric("checkAndDeleteTime_num_ops", "The description of checkAndDeleteTime_num_ops"),
			"checkAndDeleteTime_min":               newLatencyMetric("checkAndDeleteTime_min", "The description of checkAndDeleteTime_min"),
			"checkAndDeleteTime_max":               newLatencyMetric("checkAndDeleteTime_max", "The description of checkAndDeleteTime_max"),
			"checkAndDeleteTime_mean":              newLatencyMetric("checkAndDeleteTime_mean", "The description of checkAndDeleteTime_mean"),
			"checkAndDeleteTime_25th_percentile":   newLatencyMetric("checkAndDeleteTime_25th_percentile", "The description of checkAndDeleteTime_25th_percentile"),
			"checkAndDeleteTime_median":            newLatencyMetric("checkAndDeleteTime_median", "The description of checkAndDeleteTime_median"),
			"checkAndDeleteTime_75th_percentile":   newLatencyMetric("checkAndDeleteTime_75th_percentile", "The description of checkAndDeleteTime_75th_percentile"),
			"checkAndDeleteTime_90th_percentile":   newLatencyMetric("checkAndDeleteTime_90th_percentile", "The description of checkAndDeleteTime_90th_percentile"),
			"checkAndDeleteTime_95th_percentile":   newLatencyMetric("checkAndDeleteTime_95th_percentile", "The description of checkAndDeleteTime_95th_percentile"),
			"checkAndDeleteTime_98th_percentile":   newLatencyMetric("checkAndDeleteTime_98th_percentile", "The description of checkAndDeleteTime_98th_percentile"),
			"checkAndDeleteTime_99th_percentile":   newLatencyMetric("checkAndDeleteTime_99th_percentile", "The description of checkAndDeleteTime_99th_percentile"),
			"checkAndDeleteTime_99.9th_percentile": newLatencyMetric("checkAndDeleteTime_99.9th_percentile", "The description of checkAndDeleteTime_99.9th_percentile"),
			"scanTime_num_ops":                     newLatencyMetric("scanTime_num_ops", "The description of scanTime_num_ops"),
			"scanTime_min":                         newLatencyMetric("scanTime_min", "The description of scanTime_min"),
			"scanTime_max":                         newLatencyMetric("scanTime_max", "The description of scanTime_max"),
			"scanTime_mean":                        newLatencyMetric("scanTime_mean", "The description of scanTime_mean"),
			"scanTime_25th_percentile":             newLatencyMetric("scanTime_25th_percentile", "The description of scanTime_25th_percentile"),
			"scanTime_median":                      newLatencyMetric("scanTime_median", "The description of scanTime_median"),
			"scanTime_75th_percentile":             newLatencyMetric("scanTime_75th_percentile", "The description of scanTime_75th_percentile"),
			"scanTime_90th_percentile":             newLatencyMetric("scanTime_90th_percentile", "The description of scanTime_90th_percentile"),
			"scanTime_95th_percentile":             newLatencyMetric("scanTime_95th_percentile", "The description of scanTime_95th_percentile"),
			"scanTime_98th_percentile":             newLatencyMetric("scanTime_98th_percentile", "The description of scanTime_98th_percentile"),
			"scanTime_99th_percentile":             newLatencyMetric("scanTime_99th_percentile", "The description of scanTime_99th_percentile"),
			"scanTime_99.9th_percentile":           newLatencyMetric("scanTime_99.9th_percentile", "The description of scanTime_99.9th_percentile"),
			"deleteTime_num_ops":                   newLatencyMetric("deleteTime_num_ops", "The description of deleteTime_num_ops"),
			"deleteTime_min":                       newLatencyMetric("deleteTime_min", "The description of deleteTime_min"),
			"deleteTime_max":                       newLatencyMetric("deleteTime_max", "The description of deleteTime_max"),
			"deleteTime_mean":                      newLatencyMetric("deleteTime_mean", "The description of deleteTime_mean"),
			"deleteTime_25th_percentile":           newLatencyMetric("deleteTime_25th_percentile", "The description of deleteTime_25th_percentile"),
			"deleteTime_median":                    newLatencyMetric("deleteTime_median", "The description of deleteTime_median"),
			"deleteTime_75th_percentile":           newLatencyMetric("deleteTime_75th_percentile", "The description of deleteTime_75th_percentile"),
			"deleteTime_90th_percentile":           newLatencyMetric("deleteTime_90th_percentile", "The description of deleteTime_90th_percentile"),
			"deleteTime_95th_percentile":           newLatencyMetric("deleteTime_95th_percentile", "The description of deleteTime_95th_percentile"),
			"deleteTime_98th_percentile":           newLatencyMetric("deleteTime_98th_percentile", "The description of deleteTime_98th_percentile"),
			"deleteTime_99th_percentile":           newLatencyMetric("deleteTime_99th_percentile", "The description of deleteTime_99th_percentile"),
			"deleteTime_99.9th_percentile":         newLatencyMetric("deleteTime_99.9th_percentile", "The description of deleteTime_99.9th_percentile"),
			"tableWriteQueryPerSecond_count":       newLatencyMetric("tableWriteQueryPerSecond_count", "The description of tableWriteQueryPerSecond_count"),
			"tableWriteQueryPerSecond_mean_rate":   newLatencyMetric("tableWriteQueryPerSecond_mean_rate", "The description of tableWriteQueryPerSecond_mean_rate"),
			"tableWriteQueryPerSecond_1min_rate":   newLatencyMetric("tableWriteQueryPerSecond_1min_rate", "The description of tableWriteQueryPerSecond_1min_rate"),
			"tableWriteQueryPerSecond_5min_rate":   newLatencyMetric("tableWriteQueryPerSecond_5min_rate", "The description of tableWriteQueryPerSecond_5min_rate"),
			"tableWriteQueryPerSecond_15min_rate":  newLatencyMetric("tableWriteQueryPerSecond_15min_rate", "The description of tableWriteQueryPerSecond_15min_rate"),
			"tableReadQueryPerSecond_count":        newLatencyMetric("tableReadQueryPerSecond_count", "The description of tableReadQueryPerSecond_count"),
			"tableReadQueryPerSecond_mean_rate":    newLatencyMetric("tableReadQueryPerSecond_mean_rate", "The description of tableReadQueryPerSecond_mean_rate"),
			"tableReadQueryPerSecond_1min_rate":    newLatencyMetric("tableReadQueryPerSecond_1min_rate", "The description of tableReadQueryPerSecond_1min_rate"),
			"tableReadQueryPerSecond_5min_rate":    newLatencyMetric("tableReadQueryPerSecond_5min_rate", "The description of tableReadQueryPerSecond_5min_rate"),
			"tableReadQueryPerSecond_15min_rate":   newLatencyMetric("tableReadQueryPerSecond_15min_rate", "The description of tableReadQueryPerSecond_15min_rate"),
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

		} else if jmx.Metric == "deleteBatchTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_max"),
					"The number of deleteBatchTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_mean"),
					"The number of deleteBatchTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_25th_percentile"),
					"The number of deleteBatchTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_median"),
					"The number of deleteBatchTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_75th_percentile"),
					"The number of deleteBatchTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_90th_percentile"),
					"The number of deleteBatchTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_95th_percentile"),
					"The number of deleteBatchTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_98th_percentile"),
					"The number of deleteBatchTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_99th_percentile"),
					"The number of deleteBatchTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteBatchTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteBatchTime_99.9th_percentile"),
					"The number of deleteBatchTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_num_ops"),
					"The number of appendTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "appendTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_min"),
					"The number of appendTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_max"),
					"The number of appendTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_mean"),
					"The number of appendTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_25th_percentile"),
					"The number of appendTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_median"),
					"The number of appendTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_75th_percentile"),
					"The number of appendTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_90th_percentile"),
					"The number of appendTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_95th_percentile"),
					"The number of appendTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_98th_percentile"),
					"The number of appendTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_99th_percentile"),
					"The number of appendTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "appendTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "appendTime_99.9th_percentile"),
					"The number of appendTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_num_ops"),
					"The number of getTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "getTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_min"),
					"The number of getTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_max"),
					"The number of getTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_mean"),
					"The number of getTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_25th_percentile"),
					"The number of getTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_median"),
					"The number of getTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_75th_percentile"),
					"The number of getTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_90th_percentile"),
					"The number of getTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_95th_percentile"),
					"The number of getTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_98th_percentile"),
					"The number of getTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_99th_percentile"),
					"The number of getTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "getTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "getTime_99.9th_percentile"),
					"The number of getTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_num_ops"),
					"The number of incrementTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "incrementTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_min"),
					"The number of incrementTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_max"),
					"The number of incrementTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_mean"),
					"The number of incrementTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_25th_percentile"),
					"The number of incrementTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_median"),
					"The number of incrementTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_75th_percentile"),
					"The number of incrementTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_90th_percentile"),
					"The number of incrementTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_95th_percentile"),
					"The number of incrementTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_98th_percentile"),
					"The number of incrementTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_99th_percentile"),
					"The number of incrementTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "incrementTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "incrementTime_99.9th_percentile"),
					"The number of incrementTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_num_ops"),
					"The number of putBatchTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "putBatchTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_min"),
					"The number of putBatchTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_max"),
					"The number of putBatchTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_mean"),
					"The number of putBatchTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_25th_percentile"),
					"The number of putBatchTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_median"),
					"The number of putBatchTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_75th_percentile"),
					"The number of putBatchTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_90th_percentile"),
					"The number of putBatchTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_95th_percentile"),
					"The number of putBatchTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_98th_percentile"),
					"The number of putBatchTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_99th_percentile"),
					"The number of putBatchTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putBatchTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putBatchTime_99.9th_percentile"),
					"The number of putBatchTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_num_ops"),
					"The number of putTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "putTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_min"),
					"The number of putTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_max"),
					"The number of putTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_mean"),
					"The number of putTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_25th_percentile"),
					"The number of putTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_median"),
					"The number of putTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_75th_percentile"),
					"The number of putTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_90th_percentile"),
					"The number of putTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_95th_percentile"),
					"The number of putTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_98th_percentile"),
					"The number of putTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_99th_percentile"),
					"The number of putTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "putTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "putTime_99.9th_percentile"),
					"The number of putTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_num_ops"),
					"The number of scanSize_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "scanSize_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_min"),
					"The number of scanSize_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_max"),
					"The number of scanSize_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_mean"),
					"The number of scanSize_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_25th_percentile"),
					"The number of scanSize_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_median"),
					"The number of scanSize_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_75th_percentile"),
					"The number of scanSize_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_90th_percentile"),
					"The number of scanSize_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_95th_percentile"),
					"The number of scanSize_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_98th_percentile"),
					"The number of scanSize_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_99th_percentile"),
					"The number of scanSize_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanSize_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanSize_99.9th_percentile"),
					"The number of scanSize_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_num_ops"),
					"The number of checkAndMutateTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "checkAndMutateTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_min"),
					"The number of checkAndMutateTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_max"),
					"The number of checkAndMutateTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_mean"),
					"The number of checkAndMutateTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_25th_percentile"),
					"The number of checkAndMutateTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_median"),
					"The number of checkAndMutateTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_75th_percentile"),
					"The number of checkAndMutateTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_90th_percentile"),
					"The number of checkAndMutateTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_95th_percentile"),
					"The number of checkAndMutateTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_98th_percentile"),
					"The number of checkAndMutateTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_99th_percentile"),
					"The number of checkAndMutateTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndMutateTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndMutateTime_99.9th_percentile"),
					"The number of checkAndMutateTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_num_ops"),
					"The number of checkAndPutTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "checkAndPutTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_min"),
					"The number of checkAndPutTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_max"),
					"The number of checkAndPutTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_mean"),
					"The number of checkAndPutTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_25th_percentile"),
					"The number of checkAndPutTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_median"),
					"The number of checkAndPutTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_75th_percentile"),
					"The number of checkAndPutTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_90th_percentile"),
					"The number of checkAndPutTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_95th_percentile"),
					"The number of checkAndPutTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_98th_percentile"),
					"The number of checkAndPutTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_99th_percentile"),
					"The number of checkAndPutTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndPutTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndPutTime_99.9th_percentile"),
					"The number of checkAndPutTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_num_ops"),
					"The number of checkAndDeleteTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "checkAndDeleteTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_min"),
					"The number of checkAndDeleteTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_max"),
					"The number of checkAndDeleteTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_mean"),
					"The number of checkAndDeleteTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_25th_percentile"),
					"The number of checkAndDeleteTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_median"),
					"The number of checkAndDeleteTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_75th_percentile"),
					"The number of checkAndDeleteTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_90th_percentile"),
					"The number of checkAndDeleteTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_95th_percentile"),
					"The number of checkAndDeleteTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_98th_percentile"),
					"The number of checkAndDeleteTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_99th_percentile"),
					"The number of checkAndDeleteTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "checkAndDeleteTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "checkAndDeleteTime_99.9th_percentile"),
					"The number of checkAndDeleteTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_num_ops"),
					"The number of scanTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "scanTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_min"),
					"The number of scanTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_max"),
					"The number of scanTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_mean"),
					"The number of scanTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_25th_percentile"),
					"The number of scanTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_median"),
					"The number of scanTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_75th_percentile"),
					"The number of scanTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_90th_percentile"),
					"The number of scanTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_95th_percentile"),
					"The number of scanTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_98th_percentile"),
					"The number of scanTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_99th_percentile"),
					"The number of scanTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "scanTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "scanTime_99.9th_percentile"),
					"The number of scanTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_num_ops" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_num_ops"),
					"The number of deleteTime_num_ops.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "deleteTime_min" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_min"),
					"The number of deleteTime_min.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_max" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_max"),
					"The number of deleteTime_max.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_mean" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_mean"),
					"The number of deleteTime_mean.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_25th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_25th_percentile"),
					"The number of deleteTime_25th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_median" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_median"),
					"The number of deleteTime_median.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_75th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_75th_percentile"),
					"The number of deleteTime_75th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_90th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_90th_percentile"),
					"The number of deleteTime_90th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_95th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_95th_percentile"),
					"The number of deleteTime_95th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_98th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_98th_percentile"),
					"The number of deleteTime_98th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_99th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_99th_percentile"),
					"The number of deleteTime_99th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "deleteTime_99.9th_percentile" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "deleteTime_99.9th_percentile"),
					"The number of deleteTime_99.9th_percentile.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableWriteQueryPerSecond_count" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableWriteQueryPerSecond_count"),
					"The number of tableWriteQueryPerSecond_count.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableWriteQueryPerSecond_mean_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableWriteQueryPerSecond_mean_rate"),
					"The number of tableWriteQueryPerSecond_mean_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableWriteQueryPerSecond_1min_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableWriteQueryPerSecond_1min_rate"),
					"The number of tableWriteQueryPerSecond_1min_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableWriteQueryPerSecond_5min_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableWriteQueryPerSecond_5min_rate"),
					"The number of tableWriteQueryPerSecond_5min_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableWriteQueryPerSecond_15min_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableWriteQueryPerSecond_15min_rate"),
					"The number of tableWriteQueryPerSecond_15min_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableReadQueryPerSecond_count" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableReadQueryPerSecond_count"),
					"The number of tableReadQueryPerSecond_count.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableReadQueryPerSecond_mean_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableReadQueryPerSecond_mean_rate"),
					"The number of tableReadQueryPerSecond_mean_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableReadQueryPerSecond_1min_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableReadQueryPerSecond_1min_rate"),
					"The number of tableReadQueryPerSecond_1min_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableReadQueryPerSecond_5min_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableReadQueryPerSecond_5min_rate"),
					"The number of tableReadQueryPerSecond_5min_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "tableReadQueryPerSecond_15min_rate" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "tableReadQueryPerSecond_15min_rate"),
					"The number of tableReadQueryPerSecond_15min_rate.",
					defaultHBaseRsLatencyLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		}
	}
}
