package client

// Response represents a HTTP response from the Home Hub
type Response struct {
	Body         string
	ResponseBody ResponseBody
	Error        error
}

// ResponseBody represents the body response reply from the Home Hub json-req endpoint
type ResponseBody struct {
	Reply *Reply `json:"reply"`
}

// Reply represents a reply from the Home Hub json-req endpoint
type Reply struct {
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
	ResponseCallbacks []ResponseCallback `json:"callbacks"`
	ResponseEvents    []responseEvent    `json:"events"`
}

// ResponseCallback represents the response details associated with a ResponseAction from the Home Hub json-req endpoint
type ResponseCallback struct {
	UID        int        `json:"uid"`
	Result     result     `json:"result"`
	XPath      string     `json:"xpath"`
	Parameters Parameters `json:"parameters"`
}

type result struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type responseEvent struct {
}

// DeviceDetail represents a device that has connected to the Home Hub at some point in time
type DeviceDetail struct {
	UID             int    `json:"uid,omitempty"`
	Alias           string `json:"Alias,omitempty"`
	PhysicalAddress string `json:"PhysAddress,omitempty"`
	IPAddress       string `json:"IPAddress,omitempty"`
	HostName        string `json:"HostName,omitempty"`
	Active          bool   `json:"Active,omitempty"`
	InterfaceType   string `json:"InterfaceType,omitempty"`
	UserHostName    string `json:"UserHostName,omitempty"`
}
