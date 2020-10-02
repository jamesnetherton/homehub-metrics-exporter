package client

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type request struct {
	Body    *requestBody `json:"request"`
	session session
	method  string
	url     string
}

type requestBody struct {
	ID                int32    `json:"id"`
	SessionID         string   `json:"session-id"`
	SessionExpiryTime string   `json:"-"`
	Priority          bool     `json:"priority"`
	Actions           []action `json:"actions"`
	CNonce            int      `json:"cnonce"`
	AuthKey           string   `json:"auth-key"`
}

type action struct {
	ID               int               `json:"id"`
	Method           string            `json:"method"`
	XPath            string            `json:"xpath,omitempty"`
	Parameters       *Parameters       `json:"parameters,omitempty"`
	InterfaceOptions *interfaceOptions `json:"options,omitempty"`
}

// Parameters represents a set of Home Hub request action parameters
type Parameters struct {
	ID             int             `json:"id,omitempty"`
	Nonce          string          `json:"nonce,omitempty"`
	Persistent     string          `json:"persistent,omitempty"`
	SessionOptions *sessionOptions `json:"session-options,omitempty"`
	User           string          `json:"user,omitempty"`
	Value          interface{}     `json:"value,omitempty"`
	Capability     *capability     `json:"capability,omitempty"`
	URI            string          `json:"uri,omitempty"`
	Data           string          `json:"data,omitempty"`
	FileName       string          `json:"FileName,omitempty"`
	StartDate      string          `json:"startDate,omitempty"`
	EndDate        string          `json:"endDate,omitempty"`
	Source         string          `json:"source,omitempty"`
}

type interfaceOptions struct {
	CapabilityFlags capabilityFlags `json:"capability-flags"`
}

type sessionOptions struct {
	Nss             []nss           `json:"nss"`
	Language        string          `json:"language"`
	ContextFlags    contextFlags    `json:"context-flags"`
	CapabilityFlags capabilityFlags `json:"capability-flags"`
	CapabilityDepth int             `json:"capability-depth"`
	TimeFormat      string          `json:"time-format"`
}

type nss struct {
	Name string `json:"name"`
	URI  string `json:"uri"`
}

type contextFlags struct {
	GetContentName bool `json:"get-content-name"`
	LocalTime      bool `json:"local-time"`
}

type capabilityFlags struct {
	Name         bool `json:"name,omitempty"`
	DefaultValue bool `json:"default-value,omitempty"`
	Restriction  bool `json:"restriction,omitempty"`
	Description  bool `json:"description,omitempty"`
	Interface    bool `json:"interface,omitempty"`
}

type capability struct {
	Type string `json:"type"`
}

type sessionData struct {
	ID        int32     `json:"req_id"`
	SessionID int       `json:"sess_id"`
	Basic     bool      `json:"basic"`
	User      string    `json:"user"`
	DataModel dataModel `json:"dataModel"`
	Ha1       string    `json:"ha1"`
	Nonce     string    `json:"nonce"`
}

type dataModel struct {
	Name string `json:"name"`
	Nss  []nss  `json:"nss"`
}

const (
	contentType    string = "application/x-www-form-urlencoded; charset=UTF-8"
	accept         string = "application/json, text/javascript, */*; q=0.01"
	encoding       string = "gzip, deflate"
	language       string = "en-GB,en-US;q=0.8,en;q=0.6"
	homeHubAPIPath string = "cgi/json-req"
)

func (req request) send() *Response {
	response := &Response{}

	session, err := getSessionData(req)
	if err != nil {
		response.Error = err
		return response
	}

	httpResponse, err := doHTTPRequest(req, session)
	if err != nil {
		response.Error = err
		return response
	}

	if httpResponse.StatusCode >= 400 {
		response.Error = fmt.Errorf("Error processing request. Hub returned HTTP response code: %d", httpResponse.StatusCode)
		return response
	}

	defer httpResponse.Body.Close()
	bodyBytes, err := ioutil.ReadAll(httpResponse.Body)
	if err != nil {
		response.Error = err
		return response
	}

	var responseBody = &ResponseBody{}

	contentType := httpResponse.Header.Get("Content-type")
	if strings.HasPrefix(contentType, "application/json") {
		json.Unmarshal(bodyBytes, responseBody)
		response.ResponseBody = *responseBody
		if responseBody.Reply != nil && responseBody.Reply.ReplyError.Description != "Ok" {
			response.Error = errors.New(responseBody.Reply.ReplyError.Description)
		}
	} else {
		response.Body = string(bodyBytes)
	}

	return response
}

func getSessionData(req request) ([]byte, error) {
	sessionData := newSessionData(&req.session)
	return json.Marshal(sessionData)
}

func getHTTPRequest(req request) (*http.Request, error) {
	if req.method == "POST" {
		payload, err := json.Marshal(req)
		if err != nil {
			return nil, err
		}

		form := url.Values{}
		form.Add("req", string(payload))
		body := strings.NewReader(form.Encode())
		return http.NewRequest(req.method, req.url, body)
	}

	return http.NewRequest(req.method, req.url, nil)
}

func doHTTPRequest(req request, session []byte) (*http.Response, error) {
	httpRequest, _ := getHTTPRequest(req)
	httpRequest.Header.Set("Content-Type", contentType)
	httpRequest.Header.Set("Accept", accept)
	httpRequest.Header.Set("Accept-Encoding", encoding)
	httpRequest.Header.Set("Accept-Language", language)
	httpRequest.AddCookie(&http.Cookie{Name: "lang", Value: "en"})
	httpRequest.AddCookie(&http.Cookie{Name: "session", Value: url.QueryEscape(string(session))})
	httpClient := &http.Client{}
	return httpClient.Do(httpRequest)
}

func newNss() *nss {
	return &nss{Name: "gtw", URI: "http://sagemcom.com/gateway-data"}
}

func newRequestBody(session session, actions []action) *requestBody {
	cnonce := cnonceGenerate()

	var ha1 string
	if session.nonce != "" {
		ha1 = hexmd5(fmt.Sprintf("%s:%s:%s", session.userName, session.nonce, session.password))
	} else {
		ha1 = hexmd5(fmt.Sprintf("%s::%s", session.userName, session.password))
	}
	authKey := hexmd5(fmt.Sprintf("%s:%d:%d:JSON:/%s", ha1, session.requestCount, cnonce, homeHubAPIPath))

	return &requestBody{
		ID:                session.requestCount,
		SessionID:         session.sessionID,
		SessionExpiryTime: "",
		Priority:          false,
		Actions:           actions,
		CNonce:            cnonce,
		AuthKey:           authKey,
	}
}

func newSessionData(session *session) *sessionData {
	newNss := newNss()
	var nssOptions []nss
	nssOptions = append(nssOptions, *newNss)

	dataModel := &dataModel{
		Name: "Internal",
		Nss:  nssOptions,
	}

	sessionID, _ := strconv.Atoi(session.sessionID)
	authKey := hexmd5(fmt.Sprintf("%s:%s:%s", session.userName, session.nonce, session.password))
	ha1 := authKey[:10] + session.password + authKey[10:]

	return &sessionData{
		ID:        session.requestCount,
		SessionID: sessionID,
		Basic:     false,
		User:      session.userName,
		DataModel: *dataModel,
		Ha1:       ha1,
		Nonce:     session.nonce,
	}
}

func hexmd5(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func cnonceGenerate() int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return random.Intn(math.MaxInt32)
}
