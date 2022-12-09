package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

var (
	defaultHBaseRsServerLabels            = []string{"host", "role"}
	defaultHBaseRsServerLabelServerValues = func(rsServer rsServerResponse) []string {
		return []string{
			rsServer.Host,
			strings.ToLower(rsServer.Role),
		}
	}
)

type rsServerMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(rsServer rsServerResponse) float64
	Labels func(rsServer rsServerResponse) []string
}

type RsServer struct {
	logger log.Logger
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	metrics []*rsServerMetric
}

func NewRsServer(logger log.Logger, url *url.URL) *RsServer {
	subsystem := "server"

	return &RsServer{
		logger: logger,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "up"),
			Help: "Was the last scrape of the ElasticSearch cluster health endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "total_scrapes"),
			Help: "Current total ElasticSearch cluster health scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),

		metrics: []*rsServerMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "mem_store_size"),
					"The number of mem_store_size.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.MemStoreSize)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "region_count"),
					"The number of region_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.RegionCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "store_count"),
					"The number of store_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.StoreCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "store_file_count"),
					"The number of store_file_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.StoreFileCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "store_file_size"),
					"The number of store_file_size.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.StoreFileSize)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_request_count"),
					"The number of total_request_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.TotalRequestCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "split_queue_length"),
					"The number of split_queue_length.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.SplitQueueLength)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "compaction_queue_length"),
					"The number of compaction_queue_length.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.CompactionQueueLength)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "flush_queue_length"),
					"The number of flush_queue_length.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.FlushQueueLength)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_count_hit_percent"),
					"The number of block_count_hit_percent.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCountHitPercent)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "slow_append_count"),
					"The number of slow_append_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.SlowAppendCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "slow_delete_count"),
					"The number of slow_delete_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.SlowDeleteCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "slow_get_count"),
					"The number of slow_get_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.SlowGetCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "slow_put_count"),
					"The number of slow_put_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.SlowPutCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "slow_increment_count"),
					"The number of slow_increment_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.SlowIncrementCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheFreeSize"),
					"The number of BlockCacheFreeSize.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheFreeSize)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheCount"),
					"The number of BlockCacheCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheSize"),
					"The number of BlockCacheSize.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheSize)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheCountHitPercent"),
					"The number of BlockCacheCountHitPercent.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheCountHitPercent)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheExpressHitPercent"),
					"The number of BlockCacheExpressHitPercent.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheExpressHitPercent)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheHitCount"),
					"The number of BlockCacheHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheHitCountPrimary"),
					"The number of BlockCacheHitCountPrimary.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheHitCountPrimary)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheMissCount"),
					"The number of BlockCacheMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheMissCountPrimary"),
					"The number of BlockCacheMissCountPrimary.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheMissCountPrimary)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheEvictionCount"),
					"The number of BlockCacheEvictionCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheEvictionCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheEvictionCountPrimary"),
					"The number of BlockCacheEvictionCountPrimary.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheEvictionCountPrimary)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheFailedInsertionCount"),
					"The number of BlockCacheFailedInsertionCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheFailedInsertionCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheDataMissCount"),
					"The number of BlockCacheDataMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheDataMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheLeafIndexMissCount"),
					"The number of BlockCacheLeafIndexMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheLeafIndexMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheBloomChunkMissCount"),
					"The number of BlockCacheBloomChunkMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheBloomChunkMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheMetaMissCount"),
					"The number of BlockCacheMetaMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheMetaMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheRootIndexMissCount"),
					"The number of BlockCacheRootIndexMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheRootIndexMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheIntermediateIndexMissCount"),
					"The number of BlockCacheIntermediateIndexMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheIntermediateIndexMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheFileInfoMissCount"),
					"The number of BlockCacheFileInfoMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheFileInfoMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheGeneralBloomMetaMissCount"),
					"The number of BlockCacheGeneralBloomMetaMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheGeneralBloomMetaMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheDeleteFamilyBloomMissCount"),
					"The number of BlockCacheDeleteFamilyBloomMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheDeleteFamilyBloomMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheTrailerMissCount"),
					"The number of BlockCacheTrailerMissCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheTrailerMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheDataHitCount"),
					"The number of BlockCacheDataHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheDataHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheLeafIndexHitCount"),
					"The number of BlockCacheLeafIndexHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheLeafIndexHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheBloomChunkHitCount"),
					"The number of BlockCacheBloomChunkHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheBloomChunkHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheMetaHitCount"),
					"The number of BlockCacheMetaHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheMetaHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheRootIndexHitCount"),
					"The number of BlockCacheRootIndexHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheRootIndexHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheIntermediateIndexHitCount"),
					"The number of BlockCacheIntermediateIndexHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheIntermediateIndexHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheFileInfoHitCount"),
					"The number of BlockCacheFileInfoHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheFileInfoHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheGeneralBloomMetaHitCount"),
					"The number of BlockCacheGeneralBloomMetaHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheGeneralBloomMetaHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheDeleteFamilyBloomHitCount"),
					"The number of BlockCacheDeleteFamilyBloomHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheDeleteFamilyBloomHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "BlockCacheTrailerHitCount"),
					"The number of BlockCacheTrailerHitCount.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheTrailerHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
		},
	}
}

func (m *RsServer) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric.Desc
	}

	ch <- m.up.Desc()
	ch <- m.totalScrapes.Desc()
	ch <- m.jsonParseFailures.Desc()
}

func (r *RsServer) fetchAndDecodeRsServer() (rsServerResponse, error) {
	var rsr rsServerResponse

	u := *r.url
	url := u.String() + "?" + "qry=Hadoop:service=HBase,name=RegionServer,sub=Server"
	res, err := http.Get(url)

	if err != nil {
		return rsr, fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
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
		return rsr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		r.jsonParseFailures.Inc()
		return rsr, err
	}

	data := gjson.Get(string(bts), "beans")
	rsServerStr := data.Array()[0].String()

	if err := json.Unmarshal([]byte(rsServerStr), &rsr); err != nil {
		r.jsonParseFailures.Inc()
		return rsr, err
	}

	return rsr, nil

}

func (r *RsServer) Collect(ch chan<- prometheus.Metric) {
	var err error

	r.totalScrapes.Inc()
	defer func() {
		ch <- r.up
		ch <- r.totalScrapes
		ch <- r.jsonParseFailures
	}()

	rsServerResp, err := r.fetchAndDecodeRsServer()

	if err != nil {
		r.up.Set(0)
		_ = level.Warn(r.logger).Log(
			"msg", "failed to fetch and decode cluster health",
			"err", err,
		)
		return
	}
	r.up.Set(1)

	for _, metric := range r.metrics {

		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			metric.Type,
			metric.Value(rsServerResp),
			metric.Labels(rsServerResp)...,
		)
	}
}
