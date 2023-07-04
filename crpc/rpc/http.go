package rpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const tag = "HttpClient"

type RequestParam map[string]any

type HttpClient struct {
	client http.Client
}

// NewHttpClient 获取一个Http客户端
func NewHttpClient() *HttpClient {
	return &HttpClient{
		client: http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{ // 请求分发 协程安全 支持连接池
				MaxIdleConnsPerHost:   5,
				MaxConnsPerHost:       100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

// GetRequest 获取一个Get请求
func (c *HttpClient) GetRequest(method, url string, args map[string]any) (*http.Request, error) {
	if args != nil && len(args) > 0 {
		url = url + "?" + toValues(args)
	}
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	return request, err
}

func (c *HttpClient) FormRequest(method, url string, args map[string]any) (*http.Request, error) {
	request, err := http.NewRequest(method, url, strings.NewReader(toValues(args)))
	if err != nil {
		return nil, err
	}
	return request, err
}

func (c *HttpClient) JsonRequest(method, url string, args map[string]any) (*http.Request, error) {
	marshal, _ := json.Marshal(args)
	request, err := http.NewRequest(method, url, bytes.NewReader(marshal))
	if err != nil {
		return nil, err
	}
	return request, err
}

// Get 发出一个Get请求
func (c *HttpClient) Get(url string, args map[string]any) ([]byte, error) {
	if args != nil && len(args) > 0 {
		url = url + "?" + toValues(args)
	}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return c.responseHandle(request)
}

// PostForm 发出一个表单形式的Post请求
func (c *HttpClient) PostForm(url string, args map[string]any) ([]byte, error) {
	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(toValues(args)))
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return c.responseHandle(request)
}

// PostJson 发出一个Json形式的Post请求
func (c *HttpClient) PostJson(url string, args map[string]any) ([]byte, error) {
	marshal, _ := json.Marshal(args)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(marshal))
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return c.responseHandle(request)
}

// 处理请求并读取响应体Body返回数据
func (c *HttpClient) responseHandle(request *http.Request) (res []byte, err error) {
	defer request.Body.Close()
	response, err := c.client.Do(request)
	defer response.Body.Close()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("response state is not 200")
	}
	reader := bufio.NewReader(response.Body)
	buf := make([]byte, 127)
	var body []byte
	for {
		n, err := reader.Read(buf)
		if err == io.EOF || n == 0 {
			break
		} else if err != nil {
			return nil, err
		}
		body = append(body, buf[:n]...)
		if n < len(buf) {
			break
		}
	}
	return body, nil
}

// 解析Map中的参数
func toValues(args map[string]any) string {
	if args != nil && len(args) > 0 {
		param := url.Values{}
		for k, v := range args {
			param.Set(k, fmt.Sprintf("%v", v))
		}
		return param.Encode()
	}
	return ""
}
