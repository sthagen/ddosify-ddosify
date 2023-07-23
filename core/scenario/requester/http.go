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

package requester

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.ddosify.com/ddosify/core/scenario/scripting/assertion"
	"go.ddosify.com/ddosify/core/scenario/scripting/assertion/evaluator"
	"go.ddosify.com/ddosify/core/scenario/scripting/extraction"
	"go.ddosify.com/ddosify/core/scenario/scripting/injection"
	"go.ddosify.com/ddosify/core/types"
	"go.ddosify.com/ddosify/core/types/regex"
	"golang.org/x/net/http2"
)

type HttpRequester struct {
	ctx                  context.Context
	proxyAddr            *url.URL
	packet               types.ScenarioStep
	client               *http.Client
	request              *http.Request
	ei                   *injection.EnvironmentInjector
	containsDynamicField map[string]bool
	containsEnvVar       map[string]bool
	debug                bool
	dynamicRgx           *regexp.Regexp
	envRgx               *regexp.Regexp
}

// Init creates a client with the given scenarioItem. HttpRequester uses the same http.Client for all requests
func (h *HttpRequester) Init(ctx context.Context, s types.ScenarioStep, proxyAddr *url.URL, debug bool,
	ei *injection.EnvironmentInjector) (err error) {
	h.ctx = ctx
	h.packet = s
	h.proxyAddr = proxyAddr
	h.ei = ei
	h.containsDynamicField = make(map[string]bool)
	h.containsEnvVar = make(map[string]bool)
	h.debug = debug
	h.dynamicRgx = regexp.MustCompile(regex.DynamicVariableRegex)
	h.envRgx = regexp.MustCompile(regex.EnvironmentVariableRegex)

	// Transport segment
	tr := h.initTransport()
	tr.MaxIdleConnsPerHost = 60000
	tr.MaxIdleConns = 0

	// http client
	h.client = &http.Client{Transport: tr, Timeout: time.Duration(h.packet.Timeout) * time.Second}
	if val, ok := h.packet.Custom["disable-redirect"]; ok {
		val := val.(bool)
		if val {
			h.client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			}
		}
	}

	// Request instance
	err = h.initRequestInstance()
	if err != nil {
		return
	}

	// body
	if h.dynamicRgx.MatchString(h.packet.Payload) {
		_, err = h.ei.InjectDynamic(h.packet.Payload)
		if err != nil {
			return
		}
		h.containsDynamicField["body"] = true
	}

	if h.envRgx.MatchString(h.packet.Payload) {
		h.containsEnvVar["body"] = true
	}

	// url
	if h.dynamicRgx.MatchString(h.packet.URL) {
		_, err = h.ei.InjectDynamic(h.packet.URL)
		if err != nil {
			return
		}
		h.containsDynamicField["url"] = true
	}

	if h.envRgx.MatchString(h.packet.URL) {
		h.containsEnvVar["url"] = true
	}

	// header
	for k, values := range h.request.Header {
		for _, v := range values {
			if h.dynamicRgx.MatchString(k) || h.dynamicRgx.MatchString(v) {
				_, err = h.ei.InjectDynamic(k)
				if err != nil {
					return
				}

				_, err = h.ei.InjectDynamic(v)
				if err != nil {
					return
				}
				h.containsDynamicField["header"] = true
			}
			if h.envRgx.MatchString(k) || h.envRgx.MatchString(v) {
				h.containsEnvVar["header"] = true
			}
		}
	}

	// basicauth
	if h.dynamicRgx.MatchString(h.packet.Auth.Username) || h.dynamicRgx.MatchString(h.packet.Auth.Password) {
		_, err = h.ei.InjectDynamic(h.packet.Auth.Username)
		if err != nil {
			return
		}

		_, err = h.ei.InjectDynamic(h.packet.Auth.Password)
		if err != nil {
			return
		}
		h.containsDynamicField["basicauth"] = true
	}

	if h.envRgx.MatchString(h.packet.Auth.Username) || h.envRgx.MatchString(h.packet.Auth.Password) {
		h.containsEnvVar["basicauth"] = true
	}

	return
}

func (h *HttpRequester) Done() {
	// MaxIdleConnsPerHost and MaxIdleConns at Transport layer configuration
	// let us reuse the connections when keep-alive enabled(default)
	// When the Job is finished, we have to Close idle connections to prevent sockets to lock in at the TIME_WAIT state.
	// Otherwise, the next job can't use these sockets because they are reserved for the current target host.

	h.client.CloseIdleConnections()
}

func (h *HttpRequester) Send(client *http.Client, envs map[string]interface{}) (res *types.ScenarioStepResult) {
	var statusCode int
	var contentLength int64
	var requestErr types.RequestError
	var reqStartTime = time.Now()

	// for debug mode
	var copiedReqBody []byte
	var respBody []byte
	var respHeaders http.Header
	var bodyReadErr error
	var extractedVars = make(map[string]interface{})
	var failedCaptures = make(map[string]string, 0)
	var failedAssertions = make([]types.FailedAssertion, 0)

	var usableVars = make(map[string]interface{}, len(envs))
	for k, v := range envs {
		usableVars[k] = v
	}

	if client == nil {
		// engine mode is 'ddosify'
		// if passed client is nil , use requesters client that is dedicated to one step, thereby one transport
		client = h.client
	} else {
		// engine mode is 'distinct-user' or 'repeated-user'
		// passed client is used for multiple steps throughout an iteration, update transport
		if client.Transport == nil {
			client.Transport = h.initTransport()
			client.Transport.(*http.Transport).MaxConnsPerHost = 1 // use same connection per host throughout an iteration
		} else {
			h.updateTransport(client.Transport.(*http.Transport))
		}

		// update client timeout
		client.Timeout = time.Duration(h.packet.Timeout) * time.Second
	}

	durations := &duration{
		serverProcessDurCh:   make(chan time.Duration, 1),
		serverProcessStartCh: make(chan time.Time, 1),
		resDurCh:             make(chan time.Duration, 1),
		resStartCh:           make(chan time.Time, 1),
	}
	headersAddedByClient := make(map[string][]string)
	trace := newTrace(durations, h.proxyAddr, headersAddedByClient)

	httpReq, err := h.prepareReq(usableVars, trace)

	if err != nil { // could not prepare req
		requestErr.Type = types.ErrorInvalidRequest
		requestErr.Reason = fmt.Sprintf("Could not prepare req, %s", err.Error())
		res = &types.ScenarioStepResult{
			StepID:    h.packet.ID,
			StepName:  h.packet.Name,
			RequestID: uuid.New(),
			Err:       requestErr,
		}

		return res
	}

	if httpReq.Body != nil {
		if int64(len(h.packet.Payload)) > 300000 {
			// Don't store req bodies bigger than 300KB
			copiedReqBody = []byte("too long body")
		} else {
			// TODO: make copiedReqBody an io.Reader, and pass same underlying buffer to both httpReq and copiedReqBody
			// copy
			buf := new(bytes.Buffer)
			io.Copy(buf, httpReq.Body)
			copiedReqBody = buf.Bytes()
			httpReq.Body = io.NopCloser(bytes.NewBuffer(copiedReqBody)) // restore body
		}
	}

	// Action
	httpRes, err := client.Do(httpReq)
	if err != nil {
		requestErr = fetchErrType(err)
		failedCaptures = h.captureEnvironmentVariables(nil, nil, nil, extractedVars)
	} else {
		// got response, no timeout or any other error, resStart should be set
		durations.setResDur()
	}

	// From the DOC: If the Body is not both read to EOF and closed,
	// the Client's underlying RoundTripper (typically Transport)
	// may not be able to re-use a persistent TCP connection to the server for a subsequent "keep-alive" request.
	if httpRes != nil {
		// read resp body conditionally
		if h.debug || len(h.packet.EnvsToCapture) > 0 || len(h.packet.Assertions) > 0 {
			respBody, bodyReadErr = io.ReadAll(httpRes.Body)
			if bodyReadErr != nil {
				requestErr = fetchErrType(bodyReadErr)
			}
		} else {
			// do not write into memory, just read
			_, bodyReadErr = io.Copy(io.Discard, httpRes.Body)
			if bodyReadErr != nil {
				requestErr = fetchErrType(bodyReadErr)
			}
		}

		httpRes.Body.Close()
		respHeaders = httpRes.Header
		contentLength = httpRes.ContentLength
		statusCode = httpRes.StatusCode
		cookies := make(map[string]*http.Cookie, len(httpRes.Cookies()))
		for _, cookie := range httpRes.Cookies() {
			cookies[cookie.Name] = &http.Cookie{
				Name:     cookie.Name,
				Value:    cookie.Value,
				Path:     cookie.Path,
				Domain:   cookie.Domain,
				Expires:  cookie.Expires,
				Secure:   cookie.Secure,
				HttpOnly: cookie.HttpOnly,
				SameSite: cookie.SameSite,
				Raw:      cookie.Raw,
				Unparsed: cookie.Unparsed,
			}
		}

		// capture
		if len(h.packet.EnvsToCapture) > 0 {
			failedCaptures = h.captureEnvironmentVariables(httpRes.Header, respBody, cookies, extractedVars)
		}

		// assert
		if len(h.packet.Assertions) > 0 {
			_, failedAssertions = h.applyAssertions(&evaluator.AssertEnv{
				StatusCode:   int64(httpRes.StatusCode),
				ResponseSize: int64(len(respBody)),
				ResponseTime: durations.totalDuration().Milliseconds(), // in ms
				Body:         string(respBody),
				Headers:      httpRes.Header,
				Variables:    concatEnvs(envs, extractedVars),
				Cookies:      cookies,
			})
		}
	}

	var ddResTime time.Duration
	if httpRes != nil && httpRes.Header.Get("x-ddsfy-response-time") != "" {
		resTime, _ := strconv.ParseFloat(httpRes.Header.Get("x-ddsfy-response-time"), 8)
		ddResTime = time.Duration(resTime*1000) * time.Millisecond
	}

	// close duration channels, so that if any goroutine is waiting on them, it can return
	go time.AfterFunc(10*time.Millisecond, durationCloseFunc(durations))

	// Finalize
	res = &types.ScenarioStepResult{
		StepID:        h.packet.ID,
		StepName:      h.packet.Name,
		RequestID:     uuid.New(),
		StatusCode:    statusCode,
		RequestTime:   reqStartTime,
		Duration:      durations.totalDuration(),
		ContentLength: contentLength,
		Err:           requestErr,

		Url:         httpReq.URL.String(),
		Method:      httpReq.Method,
		ReqHeaders:  concatHeaders(httpReq.Header, headersAddedByClient),
		ReqBody:     copiedReqBody,
		RespHeaders: respHeaders,
		RespBody:    respBody,

		Custom: map[string]interface{}{
			"dnsDuration":           durations.getDNSDur(),
			"connDuration":          durations.getConnDur(),
			"reqDuration":           durations.getReqDur(),
			"resDuration":           durations.getResDur(),
			"serverProcessDuration": durations.getServerProcessDur(),
		},
		ExtractedEnvs:    extractedVars,
		UsableEnvs:       usableVars,
		FailedCaptures:   failedCaptures,
		FailedAssertions: failedAssertions,
	}

	if strings.EqualFold(h.request.URL.Scheme, types.ProtocolHTTPS) { // TODOcorr : check here, used URL.scheme instead TODOcorr
		res.Custom["tlsDuration"] = durations.getTLSDur()
	}

	if ddResTime != 0 {
		res.Custom["ddResponseTime"] = ddResTime
	}

	return
}

var durationCloseFunc = func(d *duration) func() {
	return func() {
		d.close()
	}
}

func concatEnvs(envs1, envs2 map[string]interface{}) map[string]interface{} {
	total := make(map[string]interface{})

	for k, v := range envs1 {
		total[k] = v
	}

	for k, v := range envs2 {
		total[k] = v
	}

	return total
}

func concatHeaders(envs1, envs2 map[string][]string) map[string][]string {
	total := make(map[string][]string)

	for k, v := range envs1 {
		total[k] = v
	}

	for k, v := range envs2 {
		total[k] = v
	}

	return total
}

func (h *HttpRequester) prepareReq(envs map[string]interface{}, trace *httptrace.ClientTrace) (*http.Request, error) {
	re := regexp.MustCompile(regex.DynamicVariableRegex)
	httpReq := h.request.Clone(h.ctx)

	body := h.packet.Payload
	if h.containsDynamicField["body"] || h.containsEnvVar["body"] {
		pieces := h.ei.GenerateBodyPieces(body, envs)
		customReader := injection.DdosifyBodyReader{
			Body:   body,
			Pieces: pieces,
		}
		httpReq.Body = &customReader
		httpReq.ContentLength = int64(injection.GetContentLength(pieces))
	} else {
		// if body is constant, we can just set it
		httpReq.Body = io.NopCloser(bytes.NewReader(injection.StringToBytes(body)))
		httpReq.ContentLength = int64(len(body))
	}

	// url
	hostURL := h.packet.URL
	var errURL error

	if h.containsDynamicField["url"] {
		hostURL, _ = h.ei.InjectDynamic(hostURL)
	}
	if h.containsEnvVar["url"] {
		hostURL, errURL = h.ei.InjectEnv(hostURL, envs)
		if errURL != nil {
			return nil, errURL
		}
	}

	httpReq.URL, errURL = url.Parse(hostURL)
	if errURL != nil {
		return nil, errURL
	}

	// If Host is not given in the header, set it from the original URL
	// Note that a temporary url used in initRequest
	if httpReq.Header.Get("Host") == "" {
		httpReq.Host = httpReq.URL.Host
	}

	// header
	if h.containsDynamicField["header"] {
		for k, values := range httpReq.Header {
			for _, v := range values {
				kk := k
				vv := v
				if re.MatchString(v) {
					vv, _ = h.ei.InjectDynamic(v)
				}
				if re.MatchString(k) {
					kk, _ = h.ei.InjectDynamic(k)
					httpReq.Header.Del(k)
				}
				httpReq.Header.Set(kk, vv)
			}
		}
	}

	if h.containsEnvVar["header"] {
		for k, v := range httpReq.Header {
			// check vals
			for i, vv := range v {
				if h.envRgx.MatchString(vv) {
					vvv, err := h.ei.InjectEnv(vv, envs)
					if err != nil {
						return nil, err
					}
					v[i] = vvv
				}
			}
			httpReq.Header.Set(k, strings.Join(v, ","))

			// check keys
			if h.envRgx.MatchString(k) {
				kk, err := h.ei.InjectEnv(k, envs)
				if err != nil {
					return nil, err
				}
				httpReq.Header.Del(k)
				httpReq.Header.Set(kk, strings.Join(v, ","))
			}
		}
	}

	username, password := h.packet.Auth.Username, h.packet.Auth.Password
	if h.containsDynamicField["basicauth"] {
		username, _ = h.ei.InjectDynamic(username)
		password, _ = h.ei.InjectDynamic(password)
	}
	if h.containsEnvVar["basicauth"] {
		var err error
		username, err = h.ei.InjectEnv(username, envs)
		if err != nil {
			return nil, err
		}

		password, err = h.ei.InjectEnv(password, envs)
		if err != nil {
			return nil, err
		}
	}
	if username != "" && password != "" {
		httpReq.SetBasicAuth(username, password)
	}

	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))
	return httpReq, nil
}

// Currently we can't detect exact error type by returned err.
// But we need to find an elegant way instead of this.
func fetchErrType(err error) types.RequestError {
	var requestErr types.RequestError = types.RequestError{
		Type:   types.ErrorUnkown,
		Reason: err.Error()}

	ue, ok := err.(*url.Error)
	if ok {
		errString := ue.Error()
		if strings.Contains(errString, "proxyconnect") {
			if strings.Contains(errString, "connection refused") {
				requestErr = types.RequestError{Type: types.ErrorProxy, Reason: types.ReasonProxyFailed}
			} else if strings.Contains(errString, "Client.Timeout") {
				requestErr = types.RequestError{Type: types.ErrorProxy, Reason: types.ReasonProxyTimeout}
			} else {
				requestErr = types.RequestError{Type: types.ErrorProxy, Reason: errString}
			}
		} else if strings.Contains(errString, context.DeadlineExceeded.Error()) {
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: types.ReasonConnTimeout}
		} else if strings.Contains(errString, "Client.Timeout exceeded while awaiting headers") {
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: types.ReasonConnTimeout}
		} else if strings.Contains(errString, "i/o timeout") {
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: types.ReasonReadTimeout}
		} else if strings.Contains(errString, "connection refused") {
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: types.ReasonConnRefused}
		} else if strings.Contains(errString, context.Canceled.Error()) {
			requestErr = types.RequestError{Type: types.ErrorIntented, Reason: types.ReasonCtxCanceled}
		} else if strings.Contains(errString, "connection reset by peer") {
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: "connection reset by peer"}
		} else {
			requestErr = types.RequestError{Type: types.ErrorConn, Reason: errString}
		}
	}

	return requestErr
}

func (h *HttpRequester) initTransport() *http.Transport {
	tr := &http.Transport{
		TLSClientConfig: h.initTLSConfig(),
		Proxy:           http.ProxyURL(h.proxyAddr),
	}

	tr.DisableKeepAlives = false
	if h.packet.Headers["Connection"] == "close" {
		tr.DisableKeepAlives = true
	}
	if val, ok := h.packet.Custom["disable-compression"]; ok {
		tr.DisableCompression = val.(bool)
	}
	if val, ok := h.packet.Custom["h2"]; ok {
		val := val.(bool)
		if val {
			http2.ConfigureTransport(tr)
		}
	}
	return tr
}

func (h *HttpRequester) updateTransport(tr *http.Transport) {
	tr.TLSClientConfig = h.initTLSConfig()
	tr.Proxy = http.ProxyURL(h.proxyAddr)

	tr.DisableKeepAlives = false
	if h.packet.Headers["Connection"] == "close" {
		tr.DisableKeepAlives = true
	}
	if val, ok := h.packet.Custom["disable-compression"]; ok {
		tr.DisableCompression = val.(bool)
	}
	if val, ok := h.packet.Custom["h2"]; ok {
		val := val.(bool)
		if val {
			http2.ConfigureTransport(tr)
		}
	}
}

func (h *HttpRequester) initTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	if h.packet.CertPool != nil && h.packet.Cert.Certificate != nil {
		tlsConfig.RootCAs = h.packet.CertPool
		tlsConfig.Certificates = []tls.Certificate{h.packet.Cert}
	}

	if val, ok := h.packet.Custom["hostname"]; ok {
		tlsConfig.ServerName = val.(string)
	}
	return tlsConfig
}

func (h *HttpRequester) initRequestInstance() (err error) {
	// TODOcorr: https://{{TARGET_URL}} or http://{{TARGET_URL}} could not be parsed, invalidHost
	// give a basic url for now here to avoid initiating request every time
	// override later on prepareReq
	tempValidUrl := "app.ddosify.com"
	if strings.HasPrefix(h.packet.URL, "https://") {
		tempValidUrl = "https://" + "app.ddosify.com"
	}
	h.request, err = http.NewRequest(h.packet.Method, tempValidUrl, bytes.NewBufferString(h.packet.Payload))
	if err != nil {
		return
	}

	// Headers
	header := make(http.Header)
	for k, v := range h.packet.Headers {
		header.Set(k, v)
		// Since we use a temp url, we need to override the request.Host either
		// it will be app.ddosify.com
		// or it will be the host from the headers
		// later on prepareReq, we will override the host if it is set in the headers
		if strings.EqualFold(k, "Host") {
			h.request.Host = v
		}
	}

	h.request.Header = header

	// Auth should be set after header assignment.
	if h.packet.Auth != (types.Auth{}) {
		h.request.SetBasicAuth(h.packet.Auth.Username, h.packet.Auth.Password)
	}

	// If keep-alive is false, prevent the reuse of the previous TCP connection at the request layer also.
	h.request.Close = false
	if h.packet.Headers["Connection"] == "close" {
		h.request.Close = true
	}
	return
}

func (h *HttpRequester) Type() string {
	return "HTTP"
}

func newTrace(duration *duration, proxyAddr *url.URL, headersByClient map[string][]string) *httptrace.ClientTrace {
	var dnsStart, connStart, tlsStart, reqStart time.Time

	// According to the doc in the trace.go;
	// Some of the hooks below can be triggered multiple times in case of retried connections, "Happy Eyeballs" etc..
	// Also, some of the hooks can be triggered after the TCP roundtrip if the request is not successfully finished.
	// To fetch the time only at the first trigger and prevent data race we need to use the mutex mechanism.
	// For start times, except resStart, this mutex is been using.
	// For duration calculations, "duration" struct internally uses another mutex.
	var m sync.Mutex

	return &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			m.Lock()
			if dnsStart.IsZero() {
				dnsStart = time.Now()
			}
			m.Unlock()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			m.Lock()
			// no need to handle error in here. We can detect it at http.Client.Do return.
			if dnsInfo.Err == nil {
				duration.setDNSDur(time.Since(dnsStart))
			}
			m.Unlock()
		},
		ConnectStart: func(network, addr string) {
			m.Lock()
			if connStart.IsZero() {
				connStart = time.Now()
			}
			m.Unlock()
		},
		ConnectDone: func(network, addr string, err error) {
			m.Lock()
			// no need to handle error in here. We can detect it at http.Client.Do return.
			if err == nil {
				duration.setConnDur(time.Since(connStart))
			}
			m.Unlock()
		},
		TLSHandshakeStart: func() {
			m.Lock()
			// This hook can be hit 2 times;
			// If both proxy and target are HTTPS
			//	First hit is for proxy, second is for target.
			//  To catch the second TLS start time (for target), we can't perform tlsStart.IsZero() check here.
			tlsStart = time.Now()
			m.Unlock()
		},
		TLSHandshakeDone: func(cs tls.ConnectionState, e error) {
			m.Lock()
			// This hook can be hit 2 times;
			// If proxy: HTTPS, target: HTTPS
			//	First hit is for proxy, second is for target TLS
			//  We need to calculate TLS duration if and only if the TLS handshake process is for the target.

			if e == nil {
				if proxyAddr == nil || proxyAddr.Hostname() != cs.ServerName {
					duration.setTLSDur(time.Since(tlsStart))
				}
			}
			m.Unlock()
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			m.Lock()
			if reqStart.IsZero() {
				reqStart = time.Now()
			}
			m.Unlock()
		},
		WroteRequest: func(w httptrace.WroteRequestInfo) {
			// no need to handle error in here. We can detect it at http.Client.Do return.
			if w.Err == nil {
				go duration.setReqDur(time.Since(reqStart))
				go duration.setServerProcessStart(time.Now())
			}
		},
		GotFirstResponseByte: func() {
			go duration.setServerProcessDur()
			go duration.setResStartTime(time.Now())
		},
		WroteHeaderField: func(key string, value []string) {
			headersByClient[key] = value
		},
	}
}

func (h *HttpRequester) applyAssertions(assertEnv *evaluator.AssertEnv) (bool, []types.FailedAssertion) {
	// result, failedAssertionIndex, assertionError
	assertions := h.packet.Assertions
	assertionsSuccess := true
	failedAssertions := []types.FailedAssertion{}
	for _, rule := range assertions {
		boolVal, err := assertion.Assert(rule, assertEnv)

		if err != nil {
			assertErr := err.(assertion.AssertionError)
			failedAssertions = append(failedAssertions, types.FailedAssertion{
				Rule:     assertErr.Rule(),
				Received: assertErr.Received(),
				Reason:   assertErr.Unwrap().Error(),
			})
			assertionsSuccess = false
		}
		if !boolVal {
			assertionsSuccess = false
		}
	}

	if assertionsSuccess {
		return true, nil
	}

	return false, failedAssertions

}

func (h *HttpRequester) captureEnvironmentVariables(header http.Header, respBody []byte,
	cookies map[string]*http.Cookie, extractedVars map[string]interface{}) map[string]string {
	var err error
	failedCaptures := make(map[string]string, 0)
	var captureError extraction.ExtractionError

	// request failed, only set default value for later steps
	if header == nil && respBody == nil {
		for _, ce := range h.packet.EnvsToCapture {
			extractedVars[ce.Name] = "" // default value for not extracted envs
			failedCaptures[ce.Name] = "request failed"
		}
		return failedCaptures
	}

	// extract from response
	for _, ce := range h.packet.EnvsToCapture {
		var val interface{}
		switch ce.From {
		case types.Header:
			val, err = extraction.Extract(header, ce)
		case types.Body:
			val, err = extraction.Extract(respBody, ce)
		case types.Cookie:
			val, err = extraction.Extract(cookies, ce)
		}
		if err != nil && errors.As(err, &captureError) {
			// do not terminate in case of a capture error, continue capturing
			extractedVars[ce.Name] = "" // default value for not extracted envs
			failedCaptures[ce.Name] = captureError.Error()
			continue
		}
		extractedVars[ce.Name] = val
	}

	return failedCaptures
}

type duration struct {

	// DNS lookup duration. If IP:Port porvided instead of domain, this will be 0
	dnsDur time.Duration

	// TCP connection setup duration
	connDur time.Duration

	// TLS handshake duration. For HTTP this will be 0
	tlsDur time.Duration

	// Request write duration
	reqDur time.Duration

	// Response read duration
	resDur time.Duration

	// Duration between full request write to first response. AKA Time To First Byte (TTFB)
	serverProcessDur time.Duration

	// Time at response reading start
	resStart         time.Time
	resStartCh       chan time.Time
	resStartChClosed bool

	serverProcessDurCh       chan time.Duration
	serverProcessDurChClosed bool

	serverProcessStartCh       chan time.Time
	serverProcessStartChClosed bool

	resDurCh       chan time.Duration
	resDurChClosed bool

	mu        sync.Mutex
	chMu      sync.Mutex
	getChLock sync.Mutex
}

func (d *duration) setResStartTime(t time.Time) {
	d.chMu.Lock()
	defer d.chMu.Unlock()

	if !d.resStartChClosed {
		d.resStartCh <- t
		d.resStartChClosed = true
		close(d.resStartCh)
	}
}

// this maybe called multiple times in case of retried requests by WroteRequest hook
func (d *duration) setServerProcessStart(t time.Time) {
	d.chMu.Lock()
	defer d.chMu.Unlock()

	if !d.serverProcessStartChClosed {
		d.serverProcessStartCh <- t
		d.serverProcessStartChClosed = true
		close(d.serverProcessStartCh)
	}
}

func (d *duration) setDNSDur(t time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.dnsDur == 0 {
		d.dnsDur = t
	}
}

func (d *duration) getDNSDur() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.dnsDur
}

func (d *duration) setTLSDur(t time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.tlsDur == 0 {
		d.tlsDur = t
	}
}

func (d *duration) getTLSDur() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.tlsDur
}

func (d *duration) setConnDur(t time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.connDur == 0 {
		d.connDur = t
	}
}

func (d *duration) getConnDur() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.connDur
}

func (d *duration) setReqDur(t time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.reqDur == 0 {
		d.reqDur = t
	}
}

func (d *duration) getReqDur() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.reqDur
}

func (d *duration) setServerProcessDur() {
	serverProcessStart := <-d.serverProcessStartCh
	d.chMu.Lock()
	defer d.chMu.Unlock()

	if !d.serverProcessDurChClosed {
		d.serverProcessDurCh <- time.Since(serverProcessStart)
		d.serverProcessDurChClosed = true
		close(d.serverProcessDurCh)
	}

}

func (d *duration) getServerProcessDur() time.Duration {
	d.getChLock.Lock()
	defer d.getChLock.Unlock()

	serverProcessDur, ok := <-d.serverProcessDurCh

	if !ok { // channel closed, dur already set or closed by timer
		return d.serverProcessDur
	}

	d.serverProcessDur = serverProcessDur
	return d.serverProcessDur
}

func (d *duration) setResDur() {
	resStart := <-d.resStartCh

	d.chMu.Lock()
	defer d.chMu.Unlock()

	if !d.resDurChClosed {
		d.resDurCh <- time.Since(resStart)
		d.resDurChClosed = true
		close(d.resDurCh)
	}

}

func (d *duration) getResDur() time.Duration {
	d.getChLock.Lock()
	defer d.getChLock.Unlock()

	resDur, ok := <-d.resDurCh

	if !ok { // channel closed, probably resDur already set and chan closed by sender
		return d.resDur
	}

	d.resDur = resDur
	return d.resDur

}

func (d *duration) totalDuration() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.dnsDur + d.connDur + d.tlsDur + d.reqDur + d.getServerProcessDur() + d.getResDur()
}

// normally channels are closed by sender, but in case of senders are not called, we close them here
func (d *duration) close() {
	d.chMu.Lock()
	defer d.chMu.Unlock()

	// close channels
	func() {
		defer func() {
			if r := recover(); r != nil {
				// channel already closed
			}
		}()
		d.serverProcessStartChClosed = true
		close(d.serverProcessStartCh)
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// channel already closed
			}
		}()
		d.serverProcessDurChClosed = true
		close(d.serverProcessDurCh)
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// channel already closed
			}
		}()
		d.resStartChClosed = true
		close(d.resStartCh)
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				// channel already closed
			}
		}()
		d.resDurChClosed = true
		close(d.resDurCh)
	}()

}
