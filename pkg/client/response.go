package client

// Response represents a HTTP response from the Home Hub
type Response struct {
	Body         string
	ResponseBody responseBody
	Error        error
}

type responseBody struct {
	Reply *reply `json:"reply"`
}

type reply struct {
	UID             int              `json:"uid"`
	ID              int              `json:"id"`
	ReplyError      replyError       `json:"error"`
	ResponseActions []ResponseAction `json:"actions"`
}

type replyError struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// ResponseAction represents a response action from the Home Hub
type ResponseAction struct {
	UID               int                `json:"uid"`
	ID                int                `json:"id"`
	ReplyError        replyError         `json:"error"`
	ResponseCallbacks []responseCallback `json:"callbacks"`
	ResponseEvents    []responseEvent    `json:"events"`
}

type responseCallback struct {
	UID        int        `json:"uid"`
	Result     result     `json:"result"`
	XPath      string     `json:"xpath"`
	Parameters parameters `json:"parameters"`
}

type result struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type responseEvent struct {
}

// DeviceDetail represents a device that has connected to the Home Hub at some point in time
type DeviceDetail struct {
	UID                int    `json:"uid,omitempty"`
	Alias              string `json:"Alias,omitempty"`
	PhysicalAddress    string `json:"PhysAddress,omitempty"`
	IPAddress          string `json:"IPAddress,omitempty"`
	HostName           string `json:"HostName,omitempty"`
	Active             bool   `json:"Active,omitempty"`
	InterfaceType      string `json:"InterfaceType,omitempty"`
	DetectedDeviceType string `json:"DetectedDeviceType,omitempty"`
	UserFriendlyName   string `json:"UserFriendlyName,omitempty"`
	UserHostName       string `json:"UserHostName,omitempty"`
	UserDeviceType     string `json:"UserDeviceType,omitempty"`
}
