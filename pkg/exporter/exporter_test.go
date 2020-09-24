package exporter

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/jamesnetherton/homehub-metrics-exporter/pkg/client"

	gomock "github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func TestMetricsScrapeSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := NewMockClient(ctrl)
	exporter := New(client)

	defer ctrl.Finish()
	defer prometheus.Unregister(exporter)

	client.EXPECT().GetSummaryStatistics().Return(createSummaryStatisticsResponse())
	client.EXPECT().GetBandwithStatistics().Return(createBandwidthStatisticsResponse())

	response, err := scrapeMetrics(exporter, t.Name())
	if err != nil {
		t.Fatal("Error occurred scraping metrics")
	}

	if response.StatusCode != 200 {
		t.Fatalf("Error occurred scraping metrics. Got status code %d", response.StatusCode)
	}

	scanner := bufio.NewScanner(response.Body)

	expectedMetrics := `bt_homehub_build_info{firmware="ABC123"} 1
	bt_homehub_device_downloaded_megabytes{host_name="Alias 3",ip_address="192.168.1.3",mac_address="AA:BB:CC:DD:EE:F3"} 1000
	bt_homehub_device_downloaded_megabytes{host_name="Host Name 1",ip_address="192.168.1.1",mac_address="AA:BB:CC:DD:EE:F1"} 600
	bt_homehub_device_downloaded_megabytes{host_name="Host Name 2",ip_address="192.168.1.2",mac_address="AA:BB:CC:DD:EE:F2"} 300
	bt_homehub_device_downloaded_megabytes{host_name="Host Name 4",ip_address="192.168.1.4",mac_address="AA:BB:CC:DD:EE:F4"} 300
	bt_homehub_device_downloaded_megabytes{host_name="User Host Name 5",ip_address="192.168.1.5",mac_address="AA:BB:CC:DD:EE:F5"} 100
	bt_homehub_device_downloaded_megabytes{host_name="User Host Name 6",ip_address="192.168.1.6",mac_address="AA:BB:CC:DD:EE:F6"} 1000
	bt_homehub_device_uploaded_megabytes{host_name="Alias 3",ip_address="192.168.1.3",mac_address="AA:BB:CC:DD:EE:F3"} 100
	bt_homehub_device_uploaded_megabytes{host_name="Host Name 1",ip_address="192.168.1.1",mac_address="AA:BB:CC:DD:EE:F1"} 60
	bt_homehub_device_uploaded_megabytes{host_name="Host Name 2",ip_address="192.168.1.2",mac_address="AA:BB:CC:DD:EE:F2"} 30
	bt_homehub_device_uploaded_megabytes{host_name="Host Name 4",ip_address="192.168.1.4",mac_address="AA:BB:CC:DD:EE:F4"} 30
	bt_homehub_device_uploaded_megabytes{host_name="User Host Name 5",ip_address="192.168.1.5",mac_address="AA:BB:CC:DD:EE:F5"} 10
	bt_homehub_device_uploaded_megabytes{host_name="User Host Name 6",ip_address="192.168.1.6",mac_address="AA:BB:CC:DD:EE:F6"} 100
	bt_homehub_download_bytes_total 654321
	bt_homehub_download_rate_mbps 123.45
	bt_homehub_up 1
	bt_homehub_upload_bytes_total 123456
	bt_homehub_upload_rate_mbps 543.21
	bt_homehub_uptime_seconds 9.8765421e+07`

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "bt_homehub") {
			if !containsLine(expectedMetrics, line) {
				t.Fatalf("Unexpected metric encountered: %s", line)
			}
		}
	}
}

func TestMetricsScrapeFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	client := NewMockClient(ctrl)
	exporter := New(client)

	defer ctrl.Finish()
	defer prometheus.Unregister(exporter)

	bandwidthStatsResponse := createBandwidthStatisticsResponse()
	bandwidthStatsResponse.Error = errors.New("Scrape error")

	client.EXPECT().GetSummaryStatistics().Return(bandwidthStatsResponse)
	client.EXPECT().GetBandwithStatistics().Return(createBandwidthStatisticsResponse())

	response, err := scrapeMetrics(exporter, t.Name())
	if err != nil {
		t.Fatal("Error occurred scraping metrics")
	}

	if response.StatusCode != 200 {
		t.Fatalf("Error occurred scraping metrics. Got status code %d", response.StatusCode)
	}

	scanner := bufio.NewScanner(response.Body)

	expectedMetrics := `bt_homehub_build_info{firmware="ABC123"} 1
	bt_homehub_download_rate_mbps 123.45
	bt_homehub_up 0
	bt_homehub_upload_rate_mbps 543.21
	bt_homehub_uptime_seconds 9.8765421e+07`

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "bt_homehub") {
			if !containsLine(expectedMetrics, line) {
				t.Fatalf("Unexpected metric encountered: %s", line)
			}
		}
	}
}

func containsLine(s string, match string) bool {
	for _, line := range strings.Split(s, "\n") {
		if strings.TrimSpace(line) == match {
			return true
		}
	}
	return false
}

func scrapeMetrics(exporter *Exporter, name string) (*http.Response, error) {
	prometheus.MustRegister(exporter)

	http.Handle("/"+name, promhttp.Handler())

	go func() {
		http.ListenAndServe(":19092", nil)
	}()

	return http.Get("http://localhost:19092/" + name)
}

func createSummaryStatisticsResponse() *client.Response {
	return &client.Response{
		ResponseBody: client.ResponseBody{
			Reply: &client.Reply{
				ResponseActions: createResponseActions(),
			},
		},
	}
}

func createBandwidthStatisticsResponse() *client.Response {
	bandwidthStatistics := `FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F1,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F1,2016-12-30,200,20
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F1,2016-12-30,300,30
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F2,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F2,2016-12-30,200,20
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F3,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F3,2016-12-30,200,20
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F3,2016-12-30,300,30
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F3,2016-12-30,400,40
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F4,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F4,2016-12-30,200,20
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F5,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F6,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F6,2016-12-30,200,20
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F6,2016-12-30,300,30
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F6,2016-12-30,400,40
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F7,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F7,2016-12-30,200,20
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F7,2016-12-30,300,30
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F8,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F8,2016-12-30,200,20
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F9,2016-12-30,100,10
	FAKE+SERIAL+NUMBER,AA:BB:CC:DD:EE:F10,2016-12-30,100,10`
	return &client.Response{
		Body: bandwidthStatistics,
	}
}

func createResponseActions() []client.ResponseAction {
	var responseActions []client.ResponseAction
	responseActions = append(responseActions, newResponseAction(client.FirmwareVersion, "ABC123"))
	responseActions = append(responseActions, newResponseAction(client.DownloadRate, 123.45))
	responseActions = append(responseActions, newResponseAction(client.UploadRate, 543.21))
	responseActions = append(responseActions, newResponseAction(client.UpTime, 98765421.0))
	responseActions = append(responseActions, newResponseAction(client.ConnectedDevices, createDevices()))
	responseActions = append(responseActions, newResponseAction(client.DownloadedBytes, "654321"))
	responseActions = append(responseActions, newResponseAction(client.UploadedBytes, "123456"))
	return responseActions
}

func createDevices() []interface{} {
	var deviceDetails []interface{}

	for i := 1; i <= 10; i++ {
		active := true
		interfaceType := "Ethernet"
		hostName := fmt.Sprintf("Host Name %d", i)
		userHostName := ""
		alias := ""

		if i == 3 {
			hostName = ""
			alias = fmt.Sprintf("Alias %d", i)
		} else if i >= 5 && i < 7 {
			interfaceType = "WiFi"
			hostName = ""
			userHostName = fmt.Sprintf("User Host Name %d", i)
		} else if i == 7 {
			interfaceType = "Invalid"
		} else if i > 7 {
			active = false
		}

		device := make(map[string]interface{})
		device["UID"] = i
		device["Alias"] = alias
		device["PhysAddress"] = fmt.Sprintf("AA:BB:CC:DD:EE:F%d", i)
		device["IPAddress"] = fmt.Sprintf("192.168.1.%d", i)
		device["HostName"] = hostName
		device["Active"] = active
		device["InterfaceType"] = interfaceType
		device["UserHostName"] = userHostName

		deviceDetails = append(deviceDetails, device)
	}

	return deviceDetails
}

func newResponseAction(xpath string, value interface{}) client.ResponseAction {
	var responseCallbacks []client.ResponseCallback

	responseCallback := &client.ResponseCallback{
		XPath: xpath,
		Parameters: *&client.Parameters{
			Value: value,
		},
	}

	responseCallbacks = append(responseCallbacks, *responseCallback)
	action := &client.ResponseAction{
		ResponseCallbacks: responseCallbacks,
	}
	return *action
}
