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
	defaultHBaseSystemLabels      = []string{"metricsName", "modelerType"}
	defaultHBaseLabelSystemValues = func(hbaseSystem hbaseSystemResponse) []string {
		return []string{
			hbaseSystem.Name,
			strings.ToLower(hbaseSystem.ModelerType),
		}
	}
)

type hbaseSystemMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(hbaseSystem hbaseSystemResponse) float64
	Labels func(hbaseSystem hbaseSystemResponse) []string
}

type HBaseSystem struct {
	logger log.Logger
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	metrics []*hbaseSystemMetric
}

func NewHBaseSystem(logger log.Logger, url *url.URL) *HBaseSystem {
	subsystem := "operating_system"

	return &HBaseSystem{
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

		metrics: []*hbaseSystemMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "open_file_descriptor_count"),
					"The number of open_file_descriptor_count.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.OpenFileDescriptorCount)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "max_file_descriptor_count"),
					"The number of max_file_descriptor_count.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.MaxFileDescriptorCount)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "committed_virtual_memory_size"),
					"The number of committed_virtual_memory_size.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.CommittedVirtualMemorySize)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_swap_space_size"),
					"The number of total_swap_space_size.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.TotalSwapSpaceSize)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "free_swap_space_size"),
					"The number of free_swap_space_size.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.FreeSwapSpaceSize)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "process_cpu_time"),
					"The number of process_cpu_time.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.ProcessCpuTime)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "total_physical_memory_size"),
					"The number of total_physical_memory_size.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.TotalPhysicalMemorySize)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "system_cpu_load"),
					"The number of system_cpu_load.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.SystemCpuLoad)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "process_cpu_load"),
					"The number of process_cpu_load.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.ProcessCpuLoad)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "free_physical_memory_size"),
					"The number of free_physical_memory_size.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.FreePhysicalMemorySize)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "available_processors"),
					"The number of available_processors.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.AvailableProcessors)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "system_load_average"),
					"The number of system_load_average.",
					defaultHBaseSystemLabels, nil,
				),
				Value: func(hbaseSystem hbaseSystemResponse) float64 {
					return float64(hbaseSystem.SystemLoadAverage)
				},
				Labels: defaultHBaseLabelSystemValues,
			},
		},
	}
}

func (m *HBaseSystem) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric.Desc
	}

	ch <- m.up.Desc()
	ch <- m.totalScrapes.Desc()
	ch <- m.jsonParseFailures.Desc()
}

func (m *HBaseSystem) fetchAndDecodeHBaseSystem() (hbaseSystemResponse, error) {
	var mjr hbaseSystemResponse

	u := *m.url
	// url := u.String() + "?" + "qry=java.lang:type=OperatingSystem"
	url := url.URL{
		Scheme:   u.Scheme,
		Host:     u.Host,
		User:     u.User,
		RawQuery: "qry=java.lang:type=OperatingSystem",
	}

	res, err := http.Get(url.String())

	if err != nil {
		return mjr, fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(m.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return mjr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		m.jsonParseFailures.Inc()
		return mjr, err
	}

	data := gjson.Get(string(bts), "beans")
	hbaseSystemStr := data.Array()[0].String()

	if err := json.Unmarshal([]byte(hbaseSystemStr), &mjr); err != nil {
		m.jsonParseFailures.Inc()
		return mjr, err
	}

	return mjr, nil
}

func (m *HBaseSystem) Collect(ch chan<- prometheus.Metric) {
	var err error

	m.totalScrapes.Inc()
	defer func() {
		ch <- m.up
		ch <- m.totalScrapes
		ch <- m.jsonParseFailures
	}()

	hbaseSystemResp, err := m.fetchAndDecodeHBaseSystem()

	if err != nil {
		m.up.Set(0)
		_ = level.Warn(m.logger).Log(
			"msg", "failed to fetch and decode cluster health",
			"err", err,
		)
		return
	}
	m.up.Set(1)

	for _, metric := range m.metrics {

		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			metric.Type,
			metric.Value(hbaseSystemResp),
			metric.Labels(hbaseSystemResp)...,
		)
	}
}
