/*
*
*	Ddosify - Load testing tool for any web system.
*   Copyright (C) 2021  Ddosify (https://ddosify.com)
*
*   This program is free software: you can redistribute it and/or modify
*   it under the terms of the GNU Affero General Public License as published
*   by the Free Software Foundation, either version 3 of the License, or
*   (at your option) any later version.
*
*   This program is distributed in the hope that it will be useful,
*   but WITHOUT ANY WARRANTY; without even the implied warranty of
*   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
*   GNU Affero General Public License for more details.
*
*   You should have received a copy of the GNU Affero General Public License
*   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*
 */

package core

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ddosify/go-faker/faker"
	"go.ddosify.com/ddosify/core/proxy"
	"go.ddosify.com/ddosify/core/report"
	"go.ddosify.com/ddosify/core/types"
)

//TODO: Engine stop channel close order test

func newDummyHammer() types.Hammer {
	return types.Hammer{
		Proxy:             proxy.Proxy{Strategy: proxy.ProxyTypeSingle},
		ReportDestination: report.OutputTypeStdout,
		LoadType:          types.LoadTypeLinear,
		TestDuration:      1,
		TotalReqCount:     1,
		Scenario: types.Scenario{
			Scenario: []types.ScenarioItem{
				{
					ID:       1,
					Protocol: "HTTP",
					Method:   "GET",
					URL:      "http://127.0.0.1",
				},
			},
		},
	}
}

func TestCreateEngine(t *testing.T) {
	t.Parallel()

	hInvalidProxy := newDummyHammer()
	hInvalidProxy.Proxy = proxy.Proxy{Strategy: "invalidProxy"}

	hInvalidReport := newDummyHammer()
	hInvalidReport.ReportDestination = "invalidReport"

	tests := []struct {
		name      string
		hammer    types.Hammer
		shouldErr bool
	}{
		{"Normal", newDummyHammer(), false},
		{"InvalidProxy", hInvalidProxy, true},
		{"InvalidReport", hInvalidReport, true},
	}

	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			e, err := NewEngine(context.TODO(), test.hammer)

			if test.shouldErr {
				if err == nil {
					t.Errorf("Should be errored")
				}
			} else {
				if err != nil {
					t.Errorf("Error occurred %v", err)
				}

				if e.proxyService == nil {
					t.Errorf("Proxy Service should be created")
				}
				if e.scenarioService == nil {
					t.Errorf("Scenario Service should be created")
				}
				if e.reportService == nil {
					t.Errorf("Report Service should be created")
				}
			}
		})
	}
}

// TODO: Add other load types as you implement
func TestRequestCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		loadType       string
		duration       int
		reqCount       int
		timeRunCount   types.TimeRunCount
		expectedReqArr []int
		delta          int
	}{
		{"Linear1", types.LoadTypeLinear, 1, 100, nil, []int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, 1},
		{"Linear2", types.LoadTypeLinear, 1, 5, nil, []int{1, 1, 1, 1, 1, 0, 0, 0, 0, 0}, 0},
		{"Linear3", types.LoadTypeLinear, 2, 4, nil,
			[]int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{"Linear4", types.LoadTypeLinear, 2, 23, nil,
			[]int{2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 0},
		{"Incremental1", types.LoadTypeIncremental, 1, 5, nil,
			[]int{1, 1, 1, 1, 1, 0, 0, 0, 0, 0}, 2},
		{"Incremental2", types.LoadTypeIncremental, 3, 1022, nil,
			[]int{17, 17, 17, 17, 17, 17, 17, 17, 17, 17, 35, 34, 34, 34,
				34, 34, 34, 34, 34, 34, 52, 51, 51, 51, 51, 51, 51, 51, 51, 51}, 2},
		{"Incremental3", types.LoadTypeIncremental, 5, 10, nil,
			[]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1,
				0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0}, 0},
		{"Incremental4", types.LoadTypeIncremental, 4, 10, nil,
			[]int{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1,
				0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0}, 0},
		{"Waved1", types.LoadTypeWaved, 1, 5, nil,
			[]int{1, 1, 1, 1, 1, 0, 0, 0, 0, 0}, 0},
		{"Waved2", types.LoadTypeWaved, 4, 32, nil,
			[]int{1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 1, 1,
				1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0}, 0},
		{"Waved3", types.LoadTypeWaved, 5, 10, nil,
			[]int{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1,
				0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{"Waved4", types.LoadTypeWaved, 9, 1000, nil,
			[]int{6, 6, 6, 6, 6, 6, 5, 5, 5, 5, 12, 11, 11, 11, 11, 11, 11, 11, 11, 11, 17, 17, 17, 17,
				17, 17, 16, 16, 16, 16, 17, 17, 17, 17, 17, 17, 16, 16, 16, 16, 12, 11, 11, 11, 11, 11,
				11, 11, 11, 11, 6, 6, 6, 6, 6, 6, 5, 5, 5, 5, 6, 6, 6, 6, 6, 6, 5, 5, 5, 5, 12, 11, 11,
				11, 11, 11, 11, 11, 11, 11, 17, 17, 17, 17, 17, 17, 17, 16, 16, 16}, 1},
		{"TimeRunCount1", "", 1, 100, types.TimeRunCount{{Duration: 1, Count: 100}},
			[]int{10, 10, 10, 10, 10, 10, 10, 10, 10, 10}, 1},
		{"TimeRunCount2", "", 1, 5, types.TimeRunCount{{Duration: 1, Count: 5}},
			[]int{1, 1, 1, 1, 1, 0, 0, 0, 0, 0}, 0},
		{"TimeRunCount3", "", 6, 55,
			types.TimeRunCount{{Duration: 1, Count: 20}, {Duration: 2, Count: 30}, {Duration: 3, Count: 5}},
			[]int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1,
				1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 0},
		{"TimeRunCount4", "", 5, 40,
			types.TimeRunCount{{Duration: 1, Count: 20}, {Duration: 2, Count: 0}, {Duration: 2, Count: 20}},
			[]int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
				0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, 0},
	}

	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			var timeReqMap map[int]int
			var now time.Time
			var m sync.Mutex

			// Test server
			handler := func(w http.ResponseWriter, r *http.Request) {
				m.Lock()
				i := time.Since(now).Milliseconds()/tickerInterval - 1
				timeReqMap[int(i)]++
				m.Unlock()
			}
			server := httptest.NewServer(http.HandlerFunc(handler))
			defer server.Close()

			// Prepare
			h := newDummyHammer()
			h.LoadType = test.loadType
			h.TestDuration = test.duration
			h.TimeRunCountMap = test.timeRunCount
			h.TotalReqCount = test.reqCount
			h.Scenario.Scenario[0].URL = server.URL

			now = time.Now()
			timeReqMap = make(map[int]int, 0)

			e, err := NewEngine(context.TODO(), h)
			if err != nil {
				t.Errorf("TestRequestCount error occurred %v", err)
			}

			// Act
			err = e.Init()
			if err != nil {
				t.Errorf("TestRequestCount error occurred %v", err)
			}

			e.Start()

			m.Lock()
			// Assert create reqCountArr
			if !reflect.DeepEqual(e.reqCountArr, test.expectedReqArr) {
				t.Errorf("Expected: %v, Found: %v", test.expectedReqArr, e.reqCountArr)
			}

			// Assert sent request count
			if testing.Short() {
				// Poor machine's test case assertions are special since they can't run the test fast.
				totalRecieved := 0
				for _, v := range timeReqMap {
					totalRecieved += v
				}
				expected := arraySum(test.expectedReqArr)
				if totalRecieved != expected {
					t.Errorf("Poor Machine Expected: %v, Received: %v", totalRecieved, expected)
				}
			} else {
				for i, v := range test.expectedReqArr {
					if timeReqMap[i] > v+test.delta || timeReqMap[i] < v-test.delta {
						t.Errorf("Expected: %v, Received: %v, Tick: %v", v, timeReqMap[i], i)
					}
				}
			}

			m.Unlock()
		})
	}
}

func TestRequestData(t *testing.T) {
	t.Parallel()

	var uri, header1, header2, body, protocol, method string

	// Test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		protocol = r.Proto
		method = r.Method
		uri = r.RequestURI
		header1 = r.Header.Get("Test1")
		header2 = r.Header.Get("Test2")

		bodyByte, _ := ioutil.ReadAll(r.Body)
		body = string(bodyByte)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// Prepare
	h := newDummyHammer()
	h.Scenario.Scenario[0] = types.ScenarioItem{
		ID:       1,
		Protocol: "HTTP",
		Method:   "GET",
		URL:      server.URL + "/get_test_data",
		Headers:  map[string]string{"Test1": "Test1Value", "Test2": "Test2Value"},
		Payload:  "Body content",
	}

	// Act
	e, err := NewEngine(context.TODO(), h)
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	err = e.Init()
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	e.Start()

	// Assert
	if uri != "/get_test_data" {
		t.Errorf("invalid uri received: %s", uri)
	}

	if protocol != "HTTP/1.1" {
		t.Errorf("invalid protocol received: %v", protocol)
	}

	if method != "GET" {
		t.Errorf("invalid method received: %v", method)
	}

	if header1 != "Test1Value" {
		t.Errorf("invalid header1 receieved: %s", header1)
	}

	if header2 != "Test2Value" {
		t.Errorf("invalid header2 receieved: %s", header2)
	}

	if body != "Body content" {
		t.Errorf("invalid body received: %v", body)
	}
}

func TestRequestDataForMultiScenarioStep(t *testing.T) {
	t.Parallel()

	var uri, header, body, protocol, method []string

	var m sync.Mutex

	// Test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		m.Lock()
		protocol = append(protocol, r.Proto)
		method = append(method, r.Method)
		uri = append(uri, r.RequestURI)
		header = append(header, r.Header.Get("Test"))

		bodyByte, _ := ioutil.ReadAll(r.Body)
		body = append(body, string(bodyByte))
		m.Unlock()
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// Prepare
	h := newDummyHammer()
	h.Scenario = types.Scenario{
		Scenario: []types.ScenarioItem{
			{
				ID:       1,
				Protocol: "HTTP",
				Method:   "GET",
				URL:      server.URL + "/api_get",
				Headers:  map[string]string{"Test": "h1"},
				Payload:  "Body 1",
			},
			{
				ID:       2,
				Protocol: "HTTP",
				Method:   "POST",
				URL:      server.URL + "/api_post",
				Headers:  map[string]string{"Test": "h2"},
				Payload:  "Body 2",
			},
		}}

	// Act
	e, err := NewEngine(context.TODO(), h)
	if err != nil {
		t.Errorf("TestRequestDataForMultiScenarioStep error occurred %v", err)
	}

	err = e.Init()
	if err != nil {
		t.Errorf("TestRequestDataForMultiScenarioStep error occurred %v", err)
	}

	e.Start()

	// Assert
	expected := []string{"/api_get", "/api_post"}
	if !reflect.DeepEqual(uri, expected) {
		t.Logf("%#v - %#v", uri, expected)
		t.Errorf("invalid uri receieved: %#v expected %#v", uri, expected)
	}

	expected = []string{"HTTP/1.1", "HTTP/1.1"}
	if !reflect.DeepEqual(protocol, expected) {
		t.Errorf("invalid protocol receieved: %#v expected %#v", protocol, expected)
	}

	expected = []string{"GET", "POST"}
	if !reflect.DeepEqual(method, expected) {
		t.Errorf("invalid method receieved: %#v expected %#v", method, expected)
	}

	expected = []string{"h1", "h2"}
	if !reflect.DeepEqual(header, expected) {
		t.Errorf("invalid header receieved: %#v expected %#v", header, expected)
	}

	expected = []string{"Body 1", "Body 2"}
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("invalid body receieved: %#v expected %#v", body, expected)
	}
}

func TestRequestTimeout(t *testing.T) {
	t.Parallel()

	// Prepare
	tests := []struct {
		name     string
		timeout  int
		expected bool
	}{
		{"Timeout", 1, false},
		{"NotTimeout", 3, true},
	}

	// Act
	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := false
			var m sync.Mutex

			// Test server
			handler := func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(time.Duration(2) * time.Second)

				m.Lock()
				result = true
				m.Unlock()
			}
			server := httptest.NewServer(http.HandlerFunc(handler))
			defer server.Close()

			h := newDummyHammer()
			h.Scenario.Scenario[0].Timeout = test.timeout
			h.Scenario.Scenario[0].URL = server.URL

			e, err := NewEngine(context.TODO(), h)
			if err != nil {
				t.Errorf("TestRequestTimeout error occurred %v", err)
			}

			err = e.Init()
			if err != nil {
				t.Errorf("TestRequestTimeout error occurred %v", err)
			}

			e.Start()

			// Assert
			m.Lock()
			if result != test.expected {
				t.Errorf("Expected %v, Found :%v", test.expected, result)
			}
			m.Unlock()
		})
	}
}

func TestEngineResult(t *testing.T) {
	t.Parallel()

	// Prepare
	tests := []struct {
		name           string
		cancelCtx      bool
		expectedStatus string
	}{
		{"CtxCancel", true, "stopped"},
		{"Normal", false, "done"},
	}

	// Act
	for _, tc := range tests {
		test := tc
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var m sync.Mutex

			// Test server
			handler := func(w http.ResponseWriter, r *http.Request) {
				return
			}
			server := httptest.NewServer(http.HandlerFunc(handler))
			defer server.Close()

			h := newDummyHammer()
			h.TestDuration = 2
			h.Scenario.Scenario[0].URL = server.URL

			ctx, cancel := context.WithCancel(context.Background())
			e, err := NewEngine(ctx, h)
			if err != nil {
				t.Errorf("TestRequestTimeout error occurred %v", err)
			}

			err = e.Init()
			if err != nil {
				t.Errorf("TestRequestTimeout error occurred %v", err)
			}

			if test.cancelCtx {
				time.AfterFunc(time.Duration(500)*time.Millisecond, func() {
					cancel()
				})
			}

			res := e.Start()
			cancel()

			// Assert
			m.Lock()
			if res != test.expectedStatus {
				t.Errorf("Expected %v, Found %v", test.expectedStatus, res)
			}
			m.Unlock()
		})
	}
}

func TestDynamicData(t *testing.T) {
	t.Parallel()

	var headers http.Header
	var body, uri string

	// Test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header
		uri = r.RequestURI
		bodyByte, _ := ioutil.ReadAll(r.Body)
		body = string(bodyByte)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// Prepare
	h := newDummyHammer()
	h.Scenario.Scenario[0] = types.ScenarioItem{
		ID:       1,
		Protocol: "HTTP",
		Method:   "GET",
		URL:      server.URL + "/get_test_data/{{_randomInt}}",
		Headers: map[string]string{
			"Test1":            "{{_randomInt}}",
			"{{_randomInt}}":   "Test2Value",
			"{{_randomColor}}": "{{_randomInt}}",
			"Test4":            "Test4Value",
		},
		Payload: "{{_randomJobArea}}",
		Auth: types.Auth{
			Type:     types.AuthHttpBasic,
			Username: "testuser",
			Password: "{{_randomBankAccountBic}}",
		},
	}

	// Act
	e, err := NewEngine(context.TODO(), h)
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	err = e.Init()
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	e.Start()

	// Assert
	if i, err := strconv.Atoi(headers.Get("Test1")); err != nil {
		t.Errorf("invalid header received: %v", i)
	}

	if headers.Get("Test4") != "Test4Value" {
		t.Errorf("invalid header received: %v", headers.Get("Test4"))
	}

	for k, v := range headers {
		vFirst := v[0]
		if vFirst == "Test2Value" {
			if i, err := strconv.Atoi(k); err != nil {
				t.Errorf("invalid header received: %v", i)
			}
		}
		fmt.Println(k, v)

	}

	// body
	contains := false
	for _, v := range faker.JobAreas {
		if body == v {
			contains = true
			break
		}
	}
	if contains == false {
		t.Errorf("invalid body received: %v", body)
	}

	// basic auth
	authHeader := strings.ReplaceAll(headers.Get("Authorization"), "Basic ", "")
	d, _ := base64.StdEncoding.DecodeString(authHeader)
	usernamePassword := string(d)
	usernamePasswordSlice := strings.Split(usernamePassword, ":")
	username := usernamePasswordSlice[0]
	password := usernamePasswordSlice[1]

	if username != "testuser" {
		t.Errorf("invalid username received: %v", username)
	}

	contains = false
	for _, v := range faker.BankAccountBics {
		if password == v {
			contains = true
			break
		}
	}
	if contains == false {
		t.Errorf("invalid body received: %v", body)
	}

	// uri
	uriDynamicPart := strings.ReplaceAll(uri, "/get_test_data/", "")
	if i, err := strconv.Atoi(uriDynamicPart); err != nil {
		t.Errorf("invalid uri received: %v", i)
	}
}

// The test creates a web server with Certificate auth,
// then it spawns an Engine and verifies that the auth was successfully passsed.
func TestTLSMutualAuth(t *testing.T) {
	t.Parallel()

	handlerCalls := 0

	// Test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalls += 1
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	defer server.Close()

	// prepare TLS files
	cert, certKey := generateCerts()
	certFile, keyFile, err := createCertPairFiles(cert, certKey)
	if err != nil {
		t.Errorf("Failed to prepare certs %v", err)
	}
	defer os.Remove(certFile.Name())
	defer os.Remove(keyFile.Name())

	// Prepare
	h := newDummyHammer()
	h.Scenario.Scenario[0] = types.ScenarioItem{
		ID:       1,
		Protocol: "HTTPS",
		Method:   "GET",
		URL:      "",
	}

	certVal, poolVal, err := types.ParseTLS(certFile.Name(), keyFile.Name())
	if err != nil {
		t.Errorf("Failed to parse certs %v", err)
	}

	h.Scenario.Scenario[0].Cert = certVal
	h.Scenario.Scenario[0].CertPool = poolVal

	server.TLS = &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    h.Scenario.Scenario[0].CertPool,
		Certificates: []tls.Certificate{h.Scenario.Scenario[0].Cert},
	}

	server.StartTLS()

	h.Scenario.Scenario[0].URL = server.URL

	// Act
	e, err := NewEngine(context.TODO(), h)
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	err = e.Init()
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	e.Start()

	// Assert
	if handlerCalls == 0 {
		t.Errorf("handler was not called at all: %#v", handlerCalls)
	}
}

// The test creates a web server with Certificate auth,
// then it spawns an Engine, but the engine doesn't have a certificate therefore it's expected that no handler is called.
func TestTLSMutualAuthButWeHaveNoCerts(t *testing.T) {
	t.Parallel()

	handlerCalls := 0

	// Test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalls += 1
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	defer server.Close()

	// prepare TLS files
	cert, certKey := generateCerts()
	certFile, keyFile, err := createCertPairFiles(cert, certKey)
	if err != nil {
		t.Errorf("Failed to prepare certs %v", err)
	}
	defer os.Remove(certFile.Name())
	defer os.Remove(keyFile.Name())

	// Prepare
	h := newDummyHammer()
	h.Scenario.Scenario[0] = types.ScenarioItem{
		ID:       1,
		Protocol: "HTTPS",
		Method:   "GET",
		URL:      "",
	}

	certVal, poolVal, err := types.ParseTLS(certFile.Name(), keyFile.Name())
	if err != nil {
		t.Errorf("Failed to parse certs %v", err)
	}

	h.Scenario.Scenario[0].Cert = certVal
	h.Scenario.Scenario[0].CertPool = poolVal

	server.TLS = &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    h.Scenario.Scenario[0].CertPool,
		Certificates: []tls.Certificate{h.Scenario.Scenario[0].Cert},
	}

	server.StartTLS()

	h.Scenario.Scenario[0].URL = server.URL

	// invalidate the certs
	h.Scenario.Scenario[0].CertPool = nil
	h.Scenario.Scenario[0].Cert = tls.Certificate{}

	e, err := NewEngine(context.TODO(), h)
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	err = e.Init()
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	e.Start()

	if handlerCalls != 0 {
		t.Errorf("handler was called unexpectedly: %#v", handlerCalls)
	}
}

// The test creates a web server with Certificate auth,
// then it spawns an Engine, but the engine have a different certificate therefore it's expected that no handler is called.
func TestTLSMutualAuthButServerAndClientHasDifferentCerts(t *testing.T) {
	t.Parallel()

	handlerCalls := 0

	// Test server
	handler := func(w http.ResponseWriter, r *http.Request) {
		handlerCalls += 1
	}
	server := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	defer server.Close()

	// prepare TLS files
	cert, certKey := generateCerts()
	certFile, keyFile, err := createCertPairFiles(cert, certKey)
	if err != nil {
		t.Errorf("Failed to prepare certs %v", err)
	}
	defer os.Remove(certFile.Name())
	defer os.Remove(keyFile.Name())

	// prepare server TLS files
	cert, certKey = generateCerts2()
	certFile2, keyFile2, err := createCertPairFiles(cert, certKey)
	if err != nil {
		t.Errorf("Failed to prepare certs %v", err)
	}
	defer os.Remove(certFile2.Name())
	defer os.Remove(keyFile2.Name())

	// Prepare
	h := newDummyHammer()
	h.Scenario.Scenario[0] = types.ScenarioItem{ID: 1, Protocol: "HTTPS", Method: "GET", URL: ""}

	// here we use server certs first
	certVal, poolVal, err := types.ParseTLS(certFile.Name(), keyFile.Name())
	if err != nil {
		t.Errorf("Failed to parse certs %v", err)
	}

	h.Scenario.Scenario[0].Cert = certVal
	h.Scenario.Scenario[0].CertPool = poolVal

	server.TLS = &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    h.Scenario.Scenario[0].CertPool,
		Certificates: []tls.Certificate{h.Scenario.Scenario[0].Cert},
	}

	server.StartTLS()

	h.Scenario.Scenario[0].URL = server.URL

	// here we use different certs
	// so the server and client has different pairs
	certVal, poolVal, err = types.ParseTLS(certFile2.Name(), keyFile2.Name())
	if err != nil {
		t.Errorf("Failed to parse certs %v", err)
	}

	h.Scenario.Scenario[0].Cert = certVal
	h.Scenario.Scenario[0].CertPool = poolVal

	e, err := NewEngine(context.TODO(), h)
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	err = e.Init()
	if err != nil {
		t.Errorf("TestRequestData error occurred %v", err)
	}

	e.Start()

	if handlerCalls != 0 {
		t.Errorf("handler was called unexpectedly: %#v", handlerCalls)
	}
}

func createCertPairFiles(cert string, certKey string) (*os.File, *os.File, error) {
	certFile, err := os.CreateTemp("", ".pem")
	if err != nil {
		return nil, nil, err
	}

	_, err = io.WriteString(certFile, cert)
	if err != nil {
		return nil, nil, err
	}

	keyFile, err := os.CreateTemp("", ".pem")
	if err != nil {
		return nil, nil, err
	}

	_, err = io.WriteString(keyFile, certKey)
	if err != nil {
		return nil, nil, err
	}

	return certFile, keyFile, nil
}

func generateCerts() (string, string) {
	cert := `-----BEGIN CERTIFICATE-----
MIIDazCCAlOgAwIBAgIUS4UhTks8aRCQ1k9IGn437ZyP3MgwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjEwMDUyMjM5MDVaFw0zMjEw
MDIyMjM5MDVaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDMbZctKXBx8v63TXIhM/OB7S6VfPqpzfHufhs6kAHu
jfC2ooCUqzqdg0T8bM1bjahYuAbQA1cWKYBsqfd01Po1ltWmbMf7ZvmSB6VN7kC2
Y670zee91dGDQ2yzmorJuIZAtOBVZesYLg8UHSGzSC/smJOrjYidtlbvzOcX0pv3
RCIUrNMed60EpSch/rzAJLzJmwNSQZ4vJHNlNetSkvTi7cxMWfwpcM/rN1hEmP1X
J43hJp/TNRZVnEsvs/yggP/FwUjG74mU3KfnWiv91AkkarNTNquEMJ+f4OFqMcnF
p0wqg47JTqcAAT0n1B0VB+z0hGXEFMN+IJXsHETZNG+JAgMBAAGjUzBRMB0GA1Ud
DgQWBBSIw+qUKQJjXWti5x/Cnn2GueuX5zAfBgNVHSMEGDAWgBSIw+qUKQJjXWti
5x/Cnn2GueuX5zAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQAA
DXzf8VXi4s2GScNfHf0BzMjpyrtRZ0Wbp2Vfh7OwVR6xcx+pqXNjlydM/vu2LvOK
hh7Jbo+JS+o7O24UJ9lLFkCRsZVF+NFqJf+2rdHCaOiZSdZmtjBU0dFuAGS7+lU3
M8P7WCNOm6NAKbs7VZHVcZPzp81SCPQgQIS19xRf4Irbvsijv4YdyL4Qv7aWcclb
MdZX9AH9Fx8tJq4VKvUYsCXAD0kuywMLjh+yj5O/2hMvs5rvaQvm2daQNRDNp884
uTLrNF7W7QaKEL06ZpXJoBqdKsiwn577XTDKvzN0XxQrT+xV9VHO7OXblF+Od3/Y
SzBR+QiQKy3x+LkOxhkk
-----END CERTIFICATE-----`

	certKey := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDMbZctKXBx8v63
TXIhM/OB7S6VfPqpzfHufhs6kAHujfC2ooCUqzqdg0T8bM1bjahYuAbQA1cWKYBs
qfd01Po1ltWmbMf7ZvmSB6VN7kC2Y670zee91dGDQ2yzmorJuIZAtOBVZesYLg8U
HSGzSC/smJOrjYidtlbvzOcX0pv3RCIUrNMed60EpSch/rzAJLzJmwNSQZ4vJHNl
NetSkvTi7cxMWfwpcM/rN1hEmP1XJ43hJp/TNRZVnEsvs/yggP/FwUjG74mU3Kfn
Wiv91AkkarNTNquEMJ+f4OFqMcnFp0wqg47JTqcAAT0n1B0VB+z0hGXEFMN+IJXs
HETZNG+JAgMBAAECggEAM+U6NHfJmNPD/8qER5OFpJ0Ob1qL06F5Yj7XMLWwF9wm
mGaGV7dkKOpTD/Wa6Dv82ZDWAeZnLDQa6vr228zZO9Nvp1EEL3kDsCOKvk7WVLbX
ikPfKZznE/iA1tNLmkvioPiJ3oQB+2Bt6YA/tuCDcf+FtU43uTm5tiSBIdYQS+Om
xN9OEXihk1svxHXQKa/a3nKPVLvdp3P90hDJ0PcRslXSy1V8az+A94JFEnCvnKsK
nF2rItCcXkInL0lYHZKgLHQMXGWkNl8e3PA1GZk3yF6LPNtPI1T5Ek9GwkHNw4JZ
BL/xEWLKB1qR2Z4I3UbWGVyi418kANv1eISb+49egQKBgQDraSRWB8nM5O3Zl9kT
8S5K924o1oXrO17eqQlVtQVmtUdoVvIBc6uHQZOmV1eHYpr6c95h8apNLexI22AY
SWkq9smpCnxLUsdkplwzie0F4bAzD6MCR8WIJxapUSPlyCA+8st1hquYBchKGQhd
6mMY1gzMDacYV/WhtG4E5d0nMQKBgQDeTr793n00VtpKuquFJe6Stu7Ujf64dL0s
3opLovyI0TmtMz5oCqIezwrjqc0Vy0UksWXaz0AboinDP+5n60cTEIt/6H0kryDc
dxfSHEA9BBDoQtxOFi3QGcxXbwu0i9QSoexrKY7FhA2xPji6bCcPycthhIrCpUiZ
s5gVkjHn2QKBgQCGklxLMbiSgGvXb46Qb9be1AMNJVT427+n2UmUzR6BUC+53boK
Sm1LrJkTBerrYdrmQUZnBxcrd40TORT9zTlpbhppn6zeAjwptVAPxlDQg+uNxOqS
ayToaC/0KoYy3OxSD8lvLcT56pRMh3LY/RwZHoPCQiu7Js0r21DpS93YgQKBgAuc
c09RMprsOmSS0WiX7ZkOIvVJIVfDCSpxySlgLu56dxe7yHOosoUHbVsswEB2KHtd
JKPEFWYcFzBSg4I8AK9XOuIIY5jp6L57Hexke1p0fumSrG0LrYLkBg8/Bo58iywZ
9v414nYgipKKXG4oPfYOJShHwvOdrGgSwEvIIgEpAoGAZz0yC9+x+JaoTnyUIRyI
+Aj5a4KhYjFtsZhcn/yCZHDqzJNDz6gAu579ey+J2CVOhjtgB5lowsDrHu32Hqnn
SEfyTru/ynQ8obwaRzdDYml+On86YWOw+brpMXkN+KB6bs2okE2N68v0qGPakxjt
OLDW6kKz5pI4T8lQJhdqjCU=
-----END PRIVATE KEY-----`

	return cert, certKey
}

func generateCerts2() (string, string) {
	cert := `-----BEGIN CERTIFICATE-----
MIIDazCCAlOgAwIBAgIUSun8oI56ArKxfhqNLLfEmteRHRUwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjEwMDYyMTE1NDdaFw0zMjEw
MDMyMTE1NDdaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQC63U6N03rm4I8yFmYK27DUlVdMUGSRQt6UIdPT5F2c
fv5mBRLwEANoqscNenNajGHiIqBiFQ3pG+p7BIIq11d87Of24XSll4MoK+6R9SFF
6lTdGt9HSzuCXQtMf5g6/MbgH240xrBXmwwJNkqpUzXVOeQBPzxplf1b/0ircf8n
fE81wnCtWyiu8BtlWvs/yJBTvSiIQ6w2Tp+K5oFZLCUwgQZdUcqzXp5nbWZkdO+D
hOGdiY7G+fC19GX7lVt+kw+xB/uAqmXw2WoR/Db/M8tJDzTw810ZbWp0tAw7Pga+
ybvIYN9mTFr4Tm052r2jVXAYejf8z4kdr4mCDKlSQTIlAgMBAAGjUzBRMB0GA1Ud
DgQWBBRWchX65rXlT+/xlgxhKMTX5/FdtTAfBgNVHSMEGDAWgBRWchX65rXlT+/x
lgxhKMTX5/FdtTAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQCo
I3MOkAULOaa4Vr80lVn/kZi8HIwQ1NenqyoykqO/FDS7q5o5vaeNquqOgC4scTdb
WJEgzBNpbIOxEM6ou5Q7IUlX6YZaTMK/Z0QbqjZuHA5ny8uaUERDLoDit318yNe+
0TOY5m5n+pRkFPvjnqoNNxvYabUqQ7NpgKTv277eecfGdFPi971EiT9HSUM8n7tU
1C1FNr7P1WGmng2EO1UCG3SQi1JpMGUYyFLSOP6F7wWhflO1JqdF57nmTtv8lKJ9
O4ACJ5BuWUqUyDLYjMK+oHh/c6xLHxfQKs62HuLqfaobqUPyE0kS7LXN2G7adjrs
2vBHv2U/QrjmLLF8CSdh
-----END CERTIFICATE-----`

	certKey := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC63U6N03rm4I8y
FmYK27DUlVdMUGSRQt6UIdPT5F2cfv5mBRLwEANoqscNenNajGHiIqBiFQ3pG+p7
BIIq11d87Of24XSll4MoK+6R9SFF6lTdGt9HSzuCXQtMf5g6/MbgH240xrBXmwwJ
NkqpUzXVOeQBPzxplf1b/0ircf8nfE81wnCtWyiu8BtlWvs/yJBTvSiIQ6w2Tp+K
5oFZLCUwgQZdUcqzXp5nbWZkdO+DhOGdiY7G+fC19GX7lVt+kw+xB/uAqmXw2WoR
/Db/M8tJDzTw810ZbWp0tAw7Pga+ybvIYN9mTFr4Tm052r2jVXAYejf8z4kdr4mC
DKlSQTIlAgMBAAECggEANaE4X1n3pvWCA3UMOkeM+6YU1PEpu8r+SHNg8SpUd4q3
Bp6kLcPaxppk4IhpPO6XVShs8VlrkaCSblX/6b29/Tuc420XZkMSwF/Da553uzIi
wwZoWHTOEn8TtBPWo+9SQJaksX7os2vrS2WKjgg0pgqkVntIomEKwvGEcLgZ68Gy
aCYgrJfvzS38+XhOJB00YOoq6vgqHj8YnTGtYAwwW+nI7oHGJS7H09eQV51cmQ2j
NSmc0SsGJ/IYrCMfJp0W8Ho9z66qRiFLb7vFS1050r5r3+slHCZPQwYXY6ovo2EJ
2Y5mKdem70dP8JZx6siVlOCKh/2fHOFNnegcQ/ADgQKBgQDx1ueRb7w9a/lh0PPN
8tLvclN/BJCqVoaF31f+Ah9Q7bfagkI7kmaQfYChWPLM5mXwr8YCPM1jysQOUTJp
ExBkGbngv/M0JeXSyt2Z9kbreFSll+ILnImAME+0KKjHTy1gDSvqX/a4NiZdDOaK
44r4CZSeVrpH2YY4tq/huL68xQKBgQDFzlhPEYOxTnQytPuXWRTtB5is1WNs7cU0
AKVGkqgNKj5++Jl+IT3/pDhcJXe06E1V9ldHFpwAorkbIvAEE45aqzp5ZrrlrAjJ
06wmEEgP5tQxmBj+hx6jitzDoEmqHvyN5Dm8/Kxu2VF2n4yTGEeSX+ep1ojLCeAj
heJuuO614QKBgGV+O1DeA7IDTnWuq6MS9VNoN4Jm+A+EoJAuW09OtLXSDga2A/Xc
Sw74nLMaEUvMpZuNKRxnSAtJXV5k1TMjvQ1FfqzD4d1QylLcsIOcx8aqiVu1kjgt
ScdyfwCsz6hVokVdQcDq5TAKCa+jal1/gSL3YlfRLfxZXesPQGEKl4HBAoGBALOw
BMye7nDNAgVmHv6Xr8i6k9i9Z7p2LCRXScxYQUzkSS1yi4zmibmG5qPebWXreQVT
6Gjtgv2Y1GpwTHSHh1OaJF5QEgu9QaaGIOXa+Htphu0ea+YbvJt385/KJeDikS4c
Ws7xAXsY80W9HigpcCrp8Dp6Zn17FR9v6ggG+uJBAoGAFGo7X1bpEA1bKAA04wJL
gq6wwKgTUjqnvHSo1CqPqoWeX8MM0VU9Jw2n0bxfD5He/snYO4pQUatD90kcgQch
BmvE1yTn4kzC0ZO3++qPulpXpAp4QJLIdKeAE9cPhKqe4lBboJRbJqoXCaoIxNeg
z0xcfR+tEmGlvxaHqXlQg9o=
-----END PRIVATE KEY-----`

	return cert, certKey
}
