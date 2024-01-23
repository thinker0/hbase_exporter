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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_size"),
					"The number of blockCacheSize.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_free_size"),
					"The number of block_cache_free_size.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_count"),
					"The number of block_cache_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_size"),
					"The number of block_cache_size.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_count_hit_percent"),
					"The number of block_cache_count_hit_percent.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheCountHitPercent)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_express_hit_percent"),
					"The number of block_cache_express_hit_percent.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheExpressHitPercent)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_hit_count"),
					"The number of block_cache_hit_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_hit_count_primary"),
					"The number of block_cache_hit_count_primary.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_miss_count"),
					"The number of block_cache_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_miss_count_primary"),
					"The number of block_cache_miss_count_primary.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheMissCountPrimary)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_eviction_count"),
					"The number of block_cache_eviction_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheEvictionCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_eviction_count_primary"),
					"The number of block_cache_eviction_count_primary.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheEvictionCountPrimary)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_failed_insertion_count"),
					"The number of block_cache_failed_insertion_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheFailedInsertionCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_data_miss_count"),
					"The number of block_cache_data_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheDataMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_leaf_index_miss_count"),
					"The number of block_cache_leaf_index_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheLeafIndexMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_bloom_chunk_miss_count"),
					"The number of block_cache_bloom_chunk_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheBloomChunkMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_meta_miss_count"),
					"The number of block_cache_meta_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheMetaMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_root_index_miss_count"),
					"The number of block_cache_root_index_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheRootIndexMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_intermediate_index_miss_count"),
					"The number of block_cache_intermediate_index_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheIntermediateIndexMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_file_info_miss_count"),
					"The number of block_cache_file_info_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheFileInfoMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_general_bloom_meta_miss_count"),
					"The number of block_cache_general_bloom_meta_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheGeneralBloomMetaMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_delete_family_bloom_miss_count"),
					"The number of block_cache_delete_family_bloom_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheDeleteFamilyBloomMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_trailer_miss_count"),
					"The number of block_cache_trailer_miss_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheTrailerMissCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "block_cache_data_hit_count"),
					"The number of block_cache_data_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_leaf_index_hit_count"),
					"The number of block_cache_leaf_index_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_bloom_chunk_hit_count"),
					"The number of block_cache_bloom_chunk_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_meta_hit_count"),
					"The number of block_cache_meta_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_root_index_hit_count"),
					"The number of block_cache_root_index_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_intermediate_index_hit_count"),
					"The number of block_cache_intermediate_index_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_file_info_hit_count"),
					"The number of block_cache_file_info_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_general_bloom_meta_hit_count"),
					"The number of block_cache_general_bloom_meta_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_delete_family_bloom_hit_count"),
					"The number of block_cache_delete_family_bloom_hit_count.",
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
					prometheus.BuildFQName(namespace, subsystem, "block_cache_trailer_hit_count"),
					"The number of block_cache_trailer_hit_count.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.BlockCacheTrailerHitCount)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "l1_cache_hit_ratio"),
					"The number of l1_cache_hit_ratio.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.L1CacheHitRatio)
				},
				Labels: defaultHBaseRsServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "l2_cache_hit_ratio"),
					"The number of l2_cache_hit_ratio.",
					defaultHBaseRsServerLabels, nil,
				),
				Value: func(rsServer rsServerResponse) float64 {
					return float64(rsServer.L2CacheHitRatio)
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
