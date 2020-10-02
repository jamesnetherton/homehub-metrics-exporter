package exporter

import (
	"strconv"
	"strings"
)

type device struct {
	macAddress          string
	ipAddress           string
	hostName            string
	deviceType          string
	active              bool
	bandwidthStatistics *deviceBandwidthStatistics
}

func newDevice(deviceDetails map[string]interface{}) *device {
	active := deviceDetails["Active"].(bool)
	deviceType := deviceDetails["InterfaceType"].(string)
	ipaddress := deviceDetails["IPAddress"].(string)
	mac := deviceDetails["PhysAddress"].(string)
	hostName := ""
	switch {
	case deviceDetails["UserHostName"] != "":
		hostName = deviceDetails["UserHostName"].(string)
	case deviceDetails["HostName"] != "":
		hostName = deviceDetails["HostName"].(string)
	default:
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

type deviceBandwidthStatistics struct {
	macAddress string
	uploaded   float64
	downloaded float64
}

func newDeviceBandwidthStatistics(statistics string) *deviceBandwidthStatistics {
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

		return &deviceBandwidthStatistics{
			macAddress: strings.ToUpper(statistic[1]),
			uploaded:   uploaded,
			downloaded: downloaded,
		}
	}
	return nil
}
