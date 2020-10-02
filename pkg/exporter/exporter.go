package exporter

import (
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/jamesnetherton/homehub-metrics-exporter/pkg/client"

	"github.com/prometheus/client_golang/prometheus"
)

// Exporter is an implementation of a Prometheus Exporter
type Exporter struct {
	client             client.Client
	metricDescriptions map[string]*prometheus.Desc
}

// New creates an instance of a Home Hub exporter
func New(client client.Client) *Exporter {
	return &Exporter{
		client:             client,
		metricDescriptions: createMetricDescriptions(),
	}
}

// Describe - loops through the API metrics and passes them to prometheus.Describe
func (e *Exporter) Describe(channel chan<- *prometheus.Desc) {
	for _, metricDescription := range e.metricDescriptions {
		channel <- metricDescription
	}
}

// Collect function, called on by Prometheus Client library
// This function is called when a scrape is performed on the /metrics page
func (e *Exporter) Collect(channel chan<- prometheus.Metric) {
	var devices = make(map[string]*device)

	summaryStatistics := e.client.GetSummaryStatistics()
	bandwidthStatistics := e.client.GetBandwithStatistics()

	if summaryStatistics.Error != nil || bandwidthStatistics.Error != nil {
		log.Println("Error fetching metrics from Home Hub")
		channel <- prometheus.MustNewConstMetric(e.metricDescriptions["up"], prometheus.GaugeValue, 0)
		return
	}

	for _, action := range summaryStatistics.ResponseBody.Reply.ResponseActions {
		value := reflect.ValueOf(action.ResponseCallbacks[0].Parameters.Value)

		switch action.ResponseCallbacks[0].XPath {
		case client.ConnectedDevices:
			deviceDetails := value.Interface().([]interface{})
			for _, v := range deviceDetails {
				device := newDevice(v.(map[string]interface{}))
				if device.active && (device.deviceType == "WiFi" || device.deviceType == "Ethernet") {
					devices[device.macAddress] = device
				}
			}
		case client.DownloadedBytes:
			floatValue, err := strconv.ParseFloat(value.String(), 64)
			if err == nil {
				channel <- prometheus.MustNewConstMetric(e.metricDescriptions["downloadBytes"], prometheus.GaugeValue, floatValue)
			}
		case client.DownloadRate:
			channel <- prometheus.MustNewConstMetric(e.metricDescriptions["downloadRateMbps"], prometheus.GaugeValue, value.Float())
		case client.FirmwareVersion:
			channel <- prometheus.MustNewConstMetric(e.metricDescriptions["build"], prometheus.GaugeValue, 1, value.String())
		case client.UploadedBytes:
			floatValue, err := strconv.ParseFloat(value.String(), 64)
			if err == nil {
				channel <- prometheus.MustNewConstMetric(e.metricDescriptions["uploadBytes"], prometheus.GaugeValue, floatValue)
			}
		case client.UploadRate:
			channel <- prometheus.MustNewConstMetric(e.metricDescriptions["uploadRateMbps"], prometheus.GaugeValue, value.Float())
		case client.UpTime:
			channel <- prometheus.MustNewConstMetric(e.metricDescriptions["uptime"], prometheus.GaugeValue, value.Float())
		}
	}

	for _, line := range strings.Split(bandwidthStatistics.Body, "\n") {
		statistics := newDeviceBandwidthStatistics(line)
		if statistics == nil {
			continue
		}

		device := devices[statistics.macAddress]
		if device == nil {
			continue
		}

		if device.bandwithStatistics == nil {
			if device.active {
				device.bandwithStatistics = statistics
			}
		} else {
			device.bandwithStatistics.downloaded += statistics.downloaded
			device.bandwithStatistics.uploaded += statistics.uploaded
		}
	}

	for _, device := range devices {
		if device.bandwithStatistics != nil {
			channel <- prometheus.MustNewConstMetric(e.metricDescriptions["deviceUploadedMegabytes"], prometheus.GaugeValue, device.bandwithStatistics.uploaded, device.hostName, device.ipAddress, device.macAddress)
			channel <- prometheus.MustNewConstMetric(e.metricDescriptions["deviceDownloadedMegabytes"], prometheus.GaugeValue, device.bandwithStatistics.downloaded, device.hostName, device.ipAddress, device.macAddress)
		}
	}

	channel <- prometheus.MustNewConstMetric(e.metricDescriptions["up"], prometheus.GaugeValue, 1)
}

func createMetricDescriptions() map[string]*prometheus.Desc {
	deviceLabels := []string{"host_name", "ip_address", "mac_address"}

	metricDescriptions := make(map[string]*prometheus.Desc)
	metricDescriptions["uptime"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "uptime_seconds"), "Uptime of the router", nil, nil)
	metricDescriptions["uploadRateMbps"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "upload_rate_mbps"), "Upload rate of the router", nil, nil)
	metricDescriptions["downloadRateMbps"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "download_rate_mbps"), "Download rate of the router", nil, nil)
	metricDescriptions["deviceUploadedMegabytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "device_uploaded_megabytes"), "Total megabytes downloaded by the device", deviceLabels, nil)
	metricDescriptions["deviceDownloadedMegabytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "device_downloaded_megabytes"), "Total megabytes uploaded by the device", deviceLabels, nil)
	metricDescriptions["build"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "build_info"), "Route build information", []string{"firmware"}, nil)
	metricDescriptions["up"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "up"), "Whether the router is up", nil, nil)
	metricDescriptions["downloadBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "download_bytes_total"), "Bytes downloaded from the internet", nil, nil)
	metricDescriptions["uploadBytes"] = prometheus.NewDesc(
		prometheus.BuildFQName("bt", "homehub", "upload_bytes_total"), "Bytes uploaded to the internet", nil, nil)
	return metricDescriptions
}
