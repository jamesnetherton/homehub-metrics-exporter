package client

const (
	// BandwidthMonitoring string constant for the BandwidthMonitoring request XPath expression
	BandwidthMonitoring string = "Device/Services/BandwidthMonitoring"
	// ConnectedDevices string constant for the Hosts request XPath expression
	ConnectedDevices string = "Device/Hosts/Hosts"
	// DownloadRate string constant for the DownstreamCurrRate request XPath expression
	DownloadRate string = "Device/DSL/Channels/Channel[@uid='1']/DownstreamCurrRate"
	// FirmwareVersion string constant for the ExternalFirmwareVersion request XPath expression
	FirmwareVersion string = "Device/DeviceInfo/ExternalFirmwareVersion"
	// UploadRate string constant for the UpstreamCurrRate request XPath expression
	UploadRate string = "Device/DSL/Channels/Channel[@uid='1']/UpstreamCurrRate"
	// UpTime string constant for the UpTime request XPath expression
	UpTime string = "Device/DeviceInfo/UpTime"
)
