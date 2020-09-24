package client

import (
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"
	"time"
)

type session struct {
	url          string
	apiURL       string
	userName     string
	password     string
	sessionID    string
	nonce        string
	requestCount int32
}

// Client represents an interface to the Home Hub router
type Client interface {
	Login() *Response
	GetSummaryStatistics() *Response
	GetBandwithStatistics() *Response
}

// HubClient is an instance of client
type HubClient struct {
	session session
}

// New creates a new client
func New(url string, userName string, password string) Client {
	return &HubClient{
		session: session{
			url:          url,
			apiURL:       url + "/" + homeHubAPIPath,
			userName:     userName,
			password:     hexmd5(password),
			sessionID:    "0",
			requestCount: 0,
		},
	}
}

// Login authenticates a user against the Home Hub
func (client *HubClient) Login() *Response {
	newNss := newNss()
	var nssOptions []nss
	nssOptions = append(nssOptions, *newNss)

	contextFlags := &contextFlags{
		GetContentName: true,
		LocalTime:      true,
	}

	capabilityFlags := &capabilityFlags{
		Name:         true,
		DefaultValue: false,
		Restriction:  true,
		Description:  false,
	}

	sessionOptions := &sessionOptions{
		Nss:             nssOptions,
		Language:        "ident",
		ContextFlags:    *contextFlags,
		CapabilityDepth: 2,
		CapabilityFlags: *capabilityFlags,
		TimeFormat:      "ISO_8601",
	}

	parameters := &Parameters{
		User:           client.session.userName,
		Persistent:     "true",
		SessionOptions: sessionOptions,
	}

	loginAction := action{
		ID:         0,
		Method:     "logIn",
		Parameters: parameters,
	}

	var actions []action
	actions = append(actions, loginAction)

	request := request{
		Body:    newRequestBody(client.session, actions),
		session: client.session,
		method:  "POST",
		url:     client.session.apiURL,
	}

	response := request.send()
	if response.Error == nil {
		responseParams := response.ResponseBody.Reply.ResponseActions[0].ResponseCallbacks[0].Parameters
		client.session.sessionID = strconv.Itoa(responseParams.ID)
		client.session.nonce = responseParams.Nonce
	}

	return response
}

// GetSummaryStatistics returns a composite response for various Home Hub metrics
func (client *HubClient) GetSummaryStatistics() *Response {
	var (
		flags   *capabilityFlags
		options *interfaceOptions
	)

	client.session.requestCount++

	flags = &capabilityFlags{
		Interface: true,
	}

	options = &interfaceOptions{
		CapabilityFlags: *flags,
	}

	var actions []action
	xpaths := []string{ConnectedDevices, DownloadedBytes, DownloadRate, FirmwareVersion, UploadedBytes, UploadRate, UpTime}

	for i, xpath := range xpaths {
		getValueAction := action{
			ID:               i,
			Method:           "getValue",
			XPath:            xpath,
			InterfaceOptions: options,
			Parameters:       nil,
		}
		actions = append(actions, getValueAction)
	}

	request := request{
		Body:    newRequestBody(client.session, actions),
		session: client.session,
		method:  "POST",
		url:     client.session.apiURL,
	}

	return request.send()
}

// GetBandwithStatistics returns a response containing a summary of bandwidth statistics for any devices
// that have connected to the Home Hub
func (client *HubClient) GetBandwithStatistics() *Response {

	var (
		options *interfaceOptions
		params  *Parameters
	)

	atomic.AddInt32(&client.session.requestCount, 1)

	var actions []action

	now := time.Now()
	params = &Parameters{
		StartDate: "20000101",
		EndDate:   fmt.Sprintf("%d%02d%02d", now.Year(), now.Month(), now.Day()),
	}

	getValueAction := action{
		ID:               0,
		Method:           "uploadBMStatisticsFile",
		XPath:            BandwidthMonitoring,
		InterfaceOptions: options,
		Parameters:       params,
	}
	actions = append(actions, getValueAction)

	statisticsFileRequest := request{
		Body:    newRequestBody(client.session, actions),
		session: client.session,
		method:  "POST",
		url:     client.session.apiURL,
	}

	response := statisticsFileRequest.send()
	vo := reflect.ValueOf(response.ResponseBody.Reply.ResponseActions[0].ResponseCallbacks[0].Parameters.Data)

	statisticsDownloadRequest := request{
		session: client.session,
		method:  "GET",
		url:     fmt.Sprintf("%s/%s", client.session.url, vo.String()),
	}

	return statisticsDownloadRequest.send()
}
