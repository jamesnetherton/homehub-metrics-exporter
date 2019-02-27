package exporter

import (
	"strconv"
	"strings"
)

type metrics struct {
	dowloadRate     float64
	firmwareVersion string
	uploadRate      float64
	uptime          float64
	err             error
}

type device struct {
	macAddress         string
	ipAddress          string
	hostName           string
	deviceType         string
	active             bool
	bandwithStatistics *deviceBandwithStatistics
}

func newDevice(deviceDetails map[string]interface{}) *device {
	active := deviceDetails["Active"].(bool)
	deviceType := deviceDetails["InterfaceType"].(string)
	ipaddress := deviceDetails["IPAddress"].(string)
	mac := deviceDetails["PhysAddress"].(string)
	hostName := ""
	if deviceDetails["UserHostName"] != "" {
		hostName = deviceDetails["UserHostName"].(string)
	} else if deviceDetails["HostName"] != "" {
		hostName = deviceDetails["HostName"].(string)
	} else {
		hostName = deviceDetails["Alias"].(string)
	}
	return &device{
		active:     active,
		deviceType: deviceType,
		macAddress: strings.ToUpper(mac),
		ipAddress:  ipaddress,
		hostName:   hostName,
	}
}

type deviceBandwithStatistics struct {
	macAddress string
	uploaded   float64
	downloaded float64
}

func newDeviceBandwidthStatistics(statistics string) *deviceBandwithStatistics {
	statistic := strings.Split(statistics, ",")

	if len(statistic) >= 5 {
		downloaded, err := strconv.ParseFloat(statistic[3], 64)
		if err != nil {
			downloaded = 0
		}

		uploaded, err := strconv.ParseFloat(statistic[4], 64)
		if err != nil {
			uploaded = 0
		}

		return &deviceBandwithStatistics{
			macAddress: strings.ToUpper(statistic[1]),
			uploaded:   uploaded,
			downloaded: downloaded,
		}
	}
	return nil
}
