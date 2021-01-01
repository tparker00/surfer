// Copyright 2020 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package s33 scrapes status from the ARRIS S33.
package s33

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"

	"github.com/wathiede/surfer/modem"
)

const idURL = "https://192.168.100.1"
const signalURL = "https://192.168.100.1/Cmconnectionstatus.html"
const hnapURL = "https://192.168.100.1/HNAP1/" // This is a bit silly but the trailing slash needs to be there or auth fails
const hnapBase = "http://purenetworks.com/HNAP1"

var (
	password = flag.String("password", "password", "Admin password if needed")
)

// JSON payload needed to send to the HNAP endpoint to for a login request
type login struct {
	Login struct {
		Action        string `json:"Action"`
		Captcha       string `json:"Captcha"`
		LoginPassword string `json:"LoginPassword"`
		PrivateLogin  string `json:"PrivateLogin"`
		Username      string `json:"Username"`
	} `json:"Login"`
}

// Response from HNAP with challenge, publicKey, etc needed to create session
type loginResponse struct {
	LoginResponse struct {
		Challenge   string `json:"Challenge"`
		Cookie      string `json:"Cookie"`
		LoginResult string `json:"LoginResult"`
		PublicKey   string `json:"PublicKey"`
	} `json:"LoginResponse"`
}

// JSON payload for getting downstream/upstream info.
// From poking around there are other HNAP endpoints that
// can be queried but not sure if they are all too useful.
type status struct {
	HNAPs struct {
		DownstreamChannelInfo string `json:"GetCustomerStatusDownstreamChannelInfo"`
		UpstreamChannelInfo   string `json:"GetCustomerStatusUpstreamChannelInfo"`
	} `json:"GetMultipleHNAPs"`
}

// Response containing downstream/upstream info
type statusResponse struct {
	HNAPsResponse struct {
		Downstream struct {
			Info   string `json:"CustomerConnDownstreamChannel"`
			Result string `json:"GetCustomerStatusDownstreamChannelInfoResult"`
		} `json:"GetCustomerStatusDownstreamChannelInfoResponse"`
		Upstream struct {
			Info   string `json:"CustomerConnUpstreamChannel"`
			Result string `json:"GetCustomerStatusUpstreamChannelInfoResult"`
		} `json:"GetCustomerStatusUpstreamChannelInfoResponse"`
		Result string `json:"GetMultipleHNAPsResult"`
	} `json:"GetMultipleHNAPsResponse"`
}

type s33 struct {
	fakeData []byte
}

func (s33) Name() string { return "S33" }

// New returns a modem.Modem that scrapes S33 formatted data at the default
// URL.
func New() modem.Modem {
	return &s33{}
}

// NewFakeData returns a modem.Modem that will parse S33 formatted data
// from the file given in path.
func NewFakeData(path string) (modem.Modem, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &s33{fakeData: b}, nil
}

// Status will return signal data parsed from an HTML status page.  If
// sb.fakeData is not nil, the fake data is parsed.  If it is nil, then an
// HTTP request is made to the default signal URL of a S33.
func (sb *s33) Status(ctx context.Context) (*modem.Signal, error) {
	if sb.fakeData != nil {
		status := statusResponse{}
		json.Unmarshal(sb.fakeData, &status)
		return parseStatus(&status)
	}

	rc, err := getStatus(ctx)
	if err != nil {
		return nil, err
	}
	return parseStatus(rc)
}

func init() {
	modem.Register(probe)
}

func isS33(b []byte) bool {
	return bytes.Contains(b, []byte(`<span id="thisModelNumberIs"> S33 </span>`))
}

func probe(ctx context.Context, path string) modem.Modem {
	if path != "" {
		b, err := ioutil.ReadFile(path)
		if err != nil {
			glog.Errorf("Failed to read %q: %v", path, err)
			return nil
		}
		if isS33(b) {
			m, err := NewFakeData(path)
			if err != nil {
				glog.Errorf("Failed to create fake S33: %v", err)
				return nil
			}
			return m
		}
		return nil
	}
	glog.Infof("Probing %q", signalURL)
	rc, err := getID(ctx)
	if err != nil {
		glog.Errorf("Failed to get status page: %v", err)
		return nil
	}
	defer rc.Close()
	b, err := ioutil.ReadAll(io.LimitReader(rc, 1<<20))
	if err != nil {
		glog.Errorf("Failed to read status page: %v", err)
		return nil
	}
	if isS33(b) {
		return New()
	}
	return nil
}

func getID(ctx context.Context) (io.ReadCloser, error) {
	client := httpClient()

	req, err := http.NewRequest("GET", idURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func getStatus(ctx context.Context) (*statusResponse, error) {
	// Cookies are used for auth after login completes
	// TODO: Store the cookies and only re-auth if we need to
	cookies, err := auth(ctx)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := httpClient()
	client.Jar = jar
	urlPath, _ := url.Parse((hnapURL))
	client.Jar.SetCookies(urlPath, cookies)

	body, _ := json.Marshal(status{})
	req, err := http.NewRequest("POST", hnapURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	var pKey string
	for _, i := range cookies {
		if i.Name == "PrivateKey" {
			pKey = i.Value
		}
	}
	hnap := hnapAuth(pKey, "GetMultipleHNAPs")

	req = req.WithContext(ctx)
	req.Header.Add("SOAPAction", fmt.Sprintf("%s/GetMultipleHNAPs", hnapBase))
	req.Header.Add("HNAP_AUTH", hnap)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	response := &statusResponse{}
	b, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(b, response)
	return response, err
}

func privateKey(l loginResponse) string {
	return encrypt(l.LoginResponse.PublicKey+*password, l.LoginResponse.Challenge)
}

func encryptedPass(l loginResponse, privateKey string) string {
	return encrypt(privateKey, l.LoginResponse.Challenge)
}

func hnapAuth(privateKey string, action string) string {
	t := (time.Now().UnixNano() / 1000000) % 2000000000000
	return fmt.Sprintf("%s %d", encrypt(privateKey, fmt.Sprintf("%d%s", t, fmt.Sprintf("%s/%s", hnapBase, action))), t)
}

func auth(ctx context.Context) ([]*http.Cookie, error) {
	// The S33 forces https via a redirect but also uses a self-signed
	// certificates from Arris.
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Transport: transport,
	}

	auth := login{}
	auth.Login.Action = "request"
	auth.Login.Username = "admin"
	auth.Login.PrivateLogin = "LoginPassword"
	authJSON, err := json.Marshal(auth)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", hnapURL, bytes.NewBuffer(authJSON))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("SOAPAction", fmt.Sprintf("%s/Login", hnapBase))
	req.Header.Add("Content-Type", "application/json")

	body, err := httpCall(client, req)
	if err != nil {
		return nil, err
	}

	parsedResponse := loginResponse{}
	err = json.Unmarshal(body, &parsedResponse)
	if err != nil {
		return nil, err
	}

	privateKey := privateKey(parsedResponse)
	encryptedPass := encryptedPass(parsedResponse, privateKey)

	hnap := hnapAuth(privateKey, "Login")

	// Once we auth the S33 uses cookies to keep things going
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client.Jar = jar

	cookies := []*http.Cookie{
		{
			Name:   "uid",
			Value:  parsedResponse.LoginResponse.Cookie,
			Path:   "/",
			Secure: true,
		},
		{
			Name:   "PrivateKey",
			Value:  privateKey,
			Path:   "/",
			Secure: true,
		},
	}

	urlPath, err := url.Parse(hnapURL)
	client.Jar.SetCookies(urlPath, cookies)
	auth.Login.Action = "login"
	auth.Login.LoginPassword = encryptedPass

	authJSON, err = json.Marshal(auth)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequest("POST", hnapURL, bytes.NewReader(authJSON))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add("SOAPAction", fmt.Sprintf("%s/Login", hnapBase))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("HNAP_AUTH", hnap)

	body, err = httpCall(client, req)
	if err != nil {
		return nil, err
	}

	if strings.Contains(string(body), "OK") {
		// Success!
		return cookies, nil
	}

	return nil, errors.New("Login Failed")
}

func encrypt(key string, value string) string {
	// The S33 uses this hash function for all its encryption dance
	encVal := hmac.New(md5.New, []byte(key))
	io.WriteString(encVal, value)
	return fmt.Sprintf("%X", encVal.Sum(nil))
}

func httpClient() *http.Client {
	// S33 uses built in certificate from Arris which isn't trusted by default
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &http.Client{Transport: transport}
}

func httpCall(client *http.Client, req *http.Request) ([]byte, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func parseStatus(s *statusResponse) (*modem.Signal, error) {
	d, err := parseDownstreamTable(s.HNAPsResponse.Downstream.Info)
	if err != nil {
		return nil, err
	}
	u, err := parseUpstreamTable(s.HNAPsResponse.Upstream.Info)
	if err != nil {
		return nil, err
	}
	return &modem.Signal{
		Downstream: d,
		Upstream:   u,
	}, nil
}

func parseDownstreamTable(t string) (map[modem.Channel]*modem.Downstream, error) {
	m := map[modem.Channel]*modem.Downstream{}
	rows := strings.Split(t, "|+|")
	if len(rows) < 0 {
		return nil, fmt.Errorf("No channels returned")
	}
	for _, row := range rows {
		d := &modem.Downstream{}
		var ch modem.Channel
		cols := strings.Split(row, "^")
		// There's a trailing ^ that we don't want to process
		for i, col := range cols[:len(cols)-1] {
			fv := col
			f, _ := strconv.ParseFloat(fv, 64)
			switch i {
			case 0:
				// Channel
			case 1:
				// Lock Status
			case 2:
				// Modulation
				d.Modulation = col
			case 3:
				// Channel ID
				ch = modem.Channel(col)
			case 4:
				// Frequency (Hz)
				d.Frequency = fmt.Sprintf("%s Hz", col)
			case 5:
				// Power (dBmV)
				d.PowerLevel = f
			case 6:
				// SNR (dB)
				d.SNR = f
			case 7:
				// Corrected
				d.Correctable = f
			case 8:
				// Uncorrectables
				d.Uncorrectable = f
			default:
				glog.Errorf("Unexpected %dth column in downstream table", i)
			}
		}
		m[ch] = d
	}
	return m, nil
}

func parseUpstreamTable(t string) (map[modem.Channel]*modem.Upstream, error) {
	m := map[modem.Channel]*modem.Upstream{}
	rows := strings.Split(t, "|+|")
	if len(rows) <= 2 {
		return nil, fmt.Errorf("Expected more than channels, got %d", len(rows))
	}
	for _, row := range rows {
		u := &modem.Upstream{}
		var ch modem.Channel
		cols := strings.Split(row, "^")
		// There's a trailing ^ that we don't want to process
		for i, col := range cols[:len(cols)-1] {
			fv := col
			f, _ := strconv.ParseFloat(fv, 64)
			switch i {
			case 0:
				// Channel Entry
			case 1:
				// Lock Status
				u.Status = col
			case 2:
				// US Channel Type
				u.Modulation = col
			case 3:
				// Channel ID
				ch = modem.Channel(col)
			case 4:
				// Width (Hz)
			case 5:
				// Frequency (Hz)
				u.Frequency = fmt.Sprintf("%s Hz", col)
			case 6:
				// Power (dBmV)
				u.PowerLevel = f
			default:
				glog.Errorf("Unexpected %dth column in upstream table", i)
			}
		}
		m[ch] = u
	}
	return m, nil
}
