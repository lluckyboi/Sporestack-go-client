package gosporestack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	version      = "1.0.0"
	userAgent    = "gosporestack/" + version
	maxRateLimit = 900 * time.Millisecond
	retryLimit   = 3
	baseURI      = "https://api.sporestack.com"
	baseURITor   = "https://api.spore64i5sofqlfz5gq2ju4msgzojjwifls7rok2cti624zyq3fcelad.onion"
)

type Client struct {
	token   string
	hclient *retryablehttp.Client

	TokenInfo *TokenInfoService
}

func NewClient(token string) *Client {
	proxyURL := getTorProxy()
	var proxyFunc func(*http.Request) (*url.URL, error)

	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err != nil {
			log.Fatalf("parase URL err: %v", err)
		}
		proxyFunc = http.ProxyURL(parsedProxyURL)
	} else {
		proxyFunc = http.ProxyFromEnvironment
	}
	transport := &http.Transport{
		Proxy: proxyFunc,
	}

	c := Client{
		token:   token,
		hclient: retryablehttp.NewClient(),
	}

	c.hclient.Logger = nil
	c.hclient.ErrorHandler = c.errorHandler
	c.hclient.RetryMax = retryLimit
	c.hclient.RetryWaitMin = maxRateLimit / 3
	c.hclient.RetryWaitMax = maxRateLimit
	c.hclient.HTTPClient.Transport = transport

	c.TokenInfo = &TokenInfoService{&c}

	return &c
}

// getTorProxy retrieves the Tor proxy URL from environment variables
func getTorProxy() string {
	return os.Getenv("TOR_PROXY")
}

func (c *Client) NewRequest(method, path string, body []byte) (*http.Request, error) {
	fullURI := baseURI + path

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, fullURI, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer: "+c.token)
	req.Header.Add("User-Agent", userAgent)

	return req, nil
}

// DoRequest performs a http request
func (c *Client) DoRequest(r *http.Request, data interface{}) error {
	rreq, err := retryablehttp.FromRequest(r)
	if err != nil {
		return err
	}

	res, err := c.hclient.Do(rreq)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		if data != nil {
			if err := json.Unmarshal(body, data); err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("error %d %s", res.StatusCode, string(body))
}

func (c *Client) errorHandler(resp *http.Response, err error, numTries int) (*http.Response, error) {
	if resp == nil {
		return nil, fmt.Errorf("gave up after %d attempts, last error unavailable (resp == nil)", retryLimit+1)
	}

	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gave up after %d attempts, last error unavailable (error reading response body: %v)", retryLimit+1, err)
	}

	return nil, fmt.Errorf("gave up after %d attempts, last error: %#v", retryLimit+1, strings.TrimSpace(string(buf)))
}
