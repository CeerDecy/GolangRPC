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
	"reflect"
	"strings"
	"time"
)

//const tag = "HttpClient"

type RequestParam map[string]any

type HttpClient struct {
	client     http.Client
	serviceMap map[string]CrService
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
		serviceMap: make(map[string]CrService),
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
	//defer request.Body.Close()
	response, err := c.client.Do(request)
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
	response.Body.Close()
	return body, nil
}

type HttpConfig struct {
	Protocol string
	Host     string
	Port     int
}

const HTTP = "http"
const HTTPS = "https"

func (h *HttpConfig) Url() string {
	switch h.Protocol {
	case HTTP, "":
		return fmt.Sprintf("http://%s:%d", h.Host, h.Port)
	case HTTPS:
		return fmt.Sprintf("https://%s:%d", h.Host, h.Port)
	default:
		panic(errors.New("unknown protocol"))
	}
}

type CrService interface {
	Evn() *HttpConfig
}

func (c *HttpClient) RegisterHttpService(name string, service CrService) {
	c.serviceMap[name] = service
}

func (c *HttpClient) Do(serviceName string, method string) CrService {
	service, ok := c.serviceMap[serviceName]
	if !ok {
		panic(errors.New("service not found"))
	}
	t := reflect.TypeOf(service)
	v := reflect.ValueOf(service)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("type is not a pointer"))
	}
	tVar := t.Elem()
	vVar := v.Elem()
	fieldIndex := -1
	for i := 0; i < tVar.NumField(); i++ {
		name := tVar.Field(i).Name
		if name == method {
			fieldIndex = i
			break
		}
	}
	if fieldIndex == -1 {
		panic(errors.New("method is not exist"))
	}
	tag := tVar.Field(fieldIndex).Tag
	if tag == "" {
		panic(errors.New("crpc tag not found"))
	}
	rpcInfo := tag.Get("crpc")
	split := strings.Split(rpcInfo, ",")
	config := service.Evn()
	if len(split) != 2 {
		panic(errors.New("crpc tag not valid"))
	}
	f := func(args map[string]any) ([]byte, error) {
		if split[0] == "GET" {
			return c.Get(config.Url()+split[1], args)
		}
		if split[0] == "POST_FORM" {
			return c.PostForm(config.Url()+split[1], args)
		}
		if split[0] == "POST_JSON" || split[0] == "POST" {
			return c.PostJson(config.Url()+split[1], args)
		}
		return nil, errors.New("unknown method")
	}
	vVar.Field(fieldIndex).Set(reflect.ValueOf(f))
	return service
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
