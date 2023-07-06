package rpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"io"
	"log"
	"net"
	"reflect"
	"sync/atomic"
	"time"
)

/**
客户端：
1. 连接服务端
2. 发送请求数据（编码） 二进制 -> 网络发送
3. 等待回复，接收响应（解码）
服务端：
1. 启动服务
2. 接受请求（解码），根据请求调用对应服务
3. 将响应数据发送给客户端（编码）
*/

const MagicNumber byte = 0x1d
const Version byte = 0x01

// Serializer 序列化接口
type Serializer interface {
	Serialize(data any) ([]byte, error)
	Deserialize(data []byte, target any) error
}

type SerializerType byte

const (
	GOB SerializerType = iota
	PROTOBUF
)

// Gob 序列化协议
type Gob struct {
}

// Serialize Gob序列化
func (g *Gob) Serialize(data any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// Deserialize Gob反序列化
func (g *Gob) Deserialize(data []byte, target any) error {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(target)
}

// Protobuf 序列化协议
type Protobuf struct {
}

// Serialize Protobuf序列化
func (p *Protobuf) Serialize(data any) ([]byte, error) {
	return proto.Marshal(data.(proto.Message))
}

// Deserialize Protobuf反序列化
func (p *Protobuf) Deserialize(data []byte, target any) error {
	return proto.Unmarshal(data, target.(proto.Message))
}

// Compressor 压缩接口
type Compressor interface {
	Compress(data []byte) ([]byte, error)
	UnCompress(data []byte) ([]byte, error)
}

// CompressType 压缩类型
type CompressType byte

const (
	GZIP CompressType = iota
)

// Gzip 压缩
type Gzip struct {
}

// Compress Gzip压缩函数
func (g *Gzip) Compress(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	writer := gzip.NewWriter(&buffer)
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// UnCompress Gzip解压函数
func (g *Gzip) UnCompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	defer reader.Close()
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(reader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// MessageType 消息类型
type MessageType byte

const (
	msgRequest MessageType = iota
	msgResponse
	msgPing
	msgPong
)

// Header 头部结构
type Header struct {
	MagicNumber   byte
	Version       byte
	FullLength    int32
	MessageType   MessageType
	CompressType  CompressType
	SerializeType SerializerType
	RequestId     int64
}

// CrRpcMessage 消息
type CrRpcMessage struct {
	Header *Header // 消息头
	Data   any     // 消息体
}

// CrRpcRequest Request请求体
type CrRpcRequest struct {
	RequestId   int64
	ServiceName string
	MethodName  string
	Args        []any
}

// CrRpcResponse Response响应体
type CrRpcResponse struct {
	RequestId      int64
	Code           int16
	Msg            string
	CompressType   CompressType
	SerializerType SerializerType
	Data           any
}

func errorResponse(err error) *CrRpcResponse {
	return &CrRpcResponse{
		Code: 500,
		Msg:  err.Error(),
	}
}

// CrRpcServer RPCServer接口
type CrRpcServer interface {
	Register(name string, service any)
	Run()
	Stop()
}

type TcpRpcServer struct {
	listen     net.Listener
	serviceMap map[string]any
}

// NewTcpRpcServer TcpRpcServer构造器
func NewTcpRpcServer(addr string) *TcpRpcServer {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	return &TcpRpcServer{
		listen:     listen,
		serviceMap: make(map[string]any),
	}
}

// Register 注册服务
func (t *TcpRpcServer) Register(name string, service any) {
	typeOf := reflect.TypeOf(service)
	if typeOf.Kind() != reflect.Pointer {
		panic(errors.New("this service is not a pointer"))
	}
	t.serviceMap[name] = service
}

type TcpConn struct {
	conn    net.Conn
	rpcChan chan *CrRpcResponse
}

// Send 发送响应体中数据
func (c *TcpConn) Send(rsp *CrRpcResponse) error {
	if rsp.Code != 200 {

	}
	headers := make([]byte, 17)
	//magic number
	headers[0] = MagicNumber
	//version
	headers[1] = Version
	//full length
	//消息类型
	headers[6] = byte(msgResponse)
	headers[7] = byte(rsp.CompressType)
	headers[8] = byte(rsp.SerializerType)
	binary.BigEndian.PutUint64(headers[9:], uint64(rsp.RequestId))
	// 编码 先序列化 再压缩
	serializer := loadSerializer(rsp.SerializerType)
	var body []byte
	var err error
	if rsp.SerializerType == PROTOBUF {
		pRsp := &Response{}
		pRsp.SerializerType = int32(rsp.SerializerType)
		pRsp.CompressType = int32(rsp.CompressType)
		pRsp.Code = int32(rsp.Code)
		pRsp.Msg = rsp.Msg
		pRsp.RequestId = rsp.RequestId
		m := make(map[string]any)
		fmt.Printf("%+v", rsp)
		data, _ := json.Marshal(rsp.Data)
		_ = json.Unmarshal(data, &m)
		value, err := structpb.NewStruct(m)
		if err != nil {
			return err
		}
		pRsp.Data = structpb.NewStructValue(value)
		fmt.Printf("%+v\n", pRsp)
		body, err = serializer.Serialize(pRsp)
		_ = serializer.Deserialize(body, pRsp)
		fmt.Printf("%+v\n", pRsp)
	} else {
		body, err = serializer.Serialize(rsp)
	}
	if err != nil {
		return err
	}
	fmt.Println(body)
	compress := loadCompress(rsp.CompressType)
	body, err = compress.Compress(body)
	if err != nil {
		return err
	}
	binary.BigEndian.PutUint32(headers[2:6], uint32(len(headers)+len(body)))
	_, err = c.conn.Write(headers[:])
	if err != nil {
		return err
	}
	_, err = c.conn.Write(body[:])
	return err
}

// Run 运行TcpServer
func (t *TcpRpcServer) Run() {
	for {
		conn, err := t.listen.Accept()
		if err != nil {
			panic(err)
		}
		tcpConn := &TcpConn{conn: conn, rpcChan: make(chan *CrRpcResponse, 1)}
		// 1、一直接受数据 解码工作 请求业务获取结果 发送到rpcChan
		// 2、编码发送数据
		go t.readHandle(tcpConn)
		go t.writeHandle(tcpConn)
	}
}

// Stop 关闭TcpServer
func (t *TcpRpcServer) Stop() error {
	return t.listen.Close()
}

// 读取请求
func (t *TcpRpcServer) readHandle(conn *TcpConn) error {
	defer func() {
		if err := recover(); err != nil {
			log.Println("TcpRpcServer", err)
			_ = conn.conn.Close()
		}
	}()
	// 接收数据
	msg, err := decodeFrame(conn.conn)
	if err != nil {
		log.Println("server readHandle", err)
		return err
	}
	if msg.Header.MessageType == msgRequest {
		if msg.Header.SerializeType == PROTOBUF {
			request := msg.Data.(*Request)
			rsp := &CrRpcResponse{
				RequestId:      request.RequestId,
				CompressType:   msg.Header.CompressType,
				SerializerType: msg.Header.SerializeType,
			}
			fmt.Printf("%+v\n", rsp)
			serviceName := request.ServiceName
			methodName := request.MethodName
			service, ok := t.serviceMap[serviceName]
			if !ok {
				err := errors.New(`service has not been registered`)
				rsp := errorResponse(err)
				conn.rpcChan <- rsp
				return err
			}
			method := reflect.ValueOf(service).MethodByName(methodName)
			if method.IsNil() {
				err := errors.New(fmt.Sprintf("no method found by this name [%s]", methodName))
				rsp := errorResponse(err)
				conn.rpcChan <- rsp
				return err
			}
			param := make([]reflect.Value, len(request.Args))
			for i := range request.Args {
				of := reflect.ValueOf(request.Args[i].AsInterface())
				param[i] = of.Convert(method.Type().In(i))
			}
			res := method.Call(param)
			results := make([]any, len(res))
			for i, v := range res {
				results[i] = v.Interface()
			}
			err, ok := results[len(results)-1].(error)
			if ok {
				rsp := errorResponse(err)
				conn.rpcChan <- rsp
				return err
			}
			rsp.Code = 200
			rsp.Msg = "success"
			rsp.Data = results[0]
			conn.rpcChan <- rsp
		} else {
			request := msg.Data.(*CrRpcRequest)
			rsp := &CrRpcResponse{
				RequestId:      request.RequestId,
				CompressType:   msg.Header.CompressType,
				SerializerType: msg.Header.SerializeType,
			}
			serviceName := request.ServiceName
			methodName := request.MethodName
			service, ok := t.serviceMap[serviceName]
			if !ok {
				err := errors.New(`service has not been registered`)
				rsp := errorResponse(err)
				conn.rpcChan <- rsp
				return err
			}
			method := reflect.ValueOf(service).MethodByName(methodName)
			if method.IsNil() {
				err := errors.New(fmt.Sprintf("no method found by this name [%s]", methodName))
				rsp := errorResponse(err)
				conn.rpcChan <- rsp
				return err
			}
			param := make([]reflect.Value, len(request.Args))
			for i, v := range request.Args {
				param[i] = reflect.ValueOf(v)
			}
			res := method.Call(param)
			results := make([]any, len(res))
			for i, v := range res {
				results[i] = v.Interface()
			}
			err, ok := results[len(results)-1].(error)
			if ok {
				rsp := errorResponse(err)
				conn.rpcChan <- rsp
				return err
			}
			rsp.Code = 200
			rsp.Msg = "success"
			rsp.Data = results[0]
			fmt.Printf("Server ReadHandle %+v", rsp)
			conn.rpcChan <- rsp
		}
	}
	return err
}

// 发送数据
func (t *TcpRpcServer) writeHandle(conn *TcpConn) {
	select {
	case rsp := <-conn.rpcChan:
		defer conn.conn.Close()
		fmt.Printf("%+v\n", rsp)
		//发送数据
		err := conn.Send(rsp)
		log.Println("writeHandle", err)
	}
}

func decodeFrame(conn net.Conn) (*CrRpcMessage, error) {
	headers := make([]byte, 17)
	_, err := io.ReadFull(conn, headers)
	if err != nil {
		return nil, err
	}
	if headers[0] != MagicNumber {
		return nil, errors.New(fmt.Sprintf("magic number error %v %v", headers, MagicNumber))
	}
	fmt.Println(headers)
	fullLength := int32(binary.BigEndian.Uint32(headers[2:6]))
	msg := &CrRpcMessage{
		Header: &Header{
			MagicNumber:   MagicNumber,
			Version:       headers[1],
			FullLength:    fullLength,
			MessageType:   MessageType(headers[6]),
			CompressType:  CompressType(headers[7]),
			SerializeType: SerializerType(headers[8]),
			RequestId:     int64(binary.BigEndian.Uint64(headers[9:])),
		},
	}
	bodyLen := fullLength - 17
	fmt.Println(bodyLen)
	//body := make([]byte, 1024)
	body := make([]byte, bodyLen)
	_, err = io.ReadFull(conn, body)
	if err != nil {
		return nil, err
	}
	// 编码 ： 先序列化 后压缩
	// 解码 ： 先解压 后反序列化
	compress := loadCompress(msg.Header.CompressType)
	fmt.Println("Get", body)
	body, err = compress.UnCompress(body)
	serializer := loadSerializer(msg.Header.SerializeType)
	if msg.Header.MessageType == msgRequest {
		if msg.Header.SerializeType == PROTOBUF {
			pReq := &Request{}
			fmt.Println(body)
			err = serializer.Deserialize(body, pReq)
			if err != nil {
				return nil, err
			}
			msg.Data = pReq
		} else {
			req := &CrRpcRequest{}
			err = serializer.Deserialize(body, req)
			if err != nil {
				return nil, err
			}
			msg.Data = req
		}
		return msg, nil
	}
	if msg.Header.MessageType == msgResponse {
		if msg.Header.SerializeType == PROTOBUF {
			fmt.Println(body)
			rsp := &Response{}
			if err = serializer.Deserialize(body, rsp); err != nil {
				return nil, err
			}
			msg.Data = rsp
		} else {
			rsp := &CrRpcResponse{}
			if err = serializer.Deserialize(body, rsp); err != nil {
				return nil, err
			}
			msg.Data = rsp
		}
		return msg, nil
	}
	return nil, errors.New("no message type")
}

// 加载序列化协议
func loadSerializer(serializeType SerializerType) Serializer {
	switch serializeType {
	case GOB:
		return &Gob{}
	case PROTOBUF:
		return &Protobuf{}
	default:
		panic("unknown serialize type")
	}
}

// 加载解压缩协议
func loadCompress(compressType CompressType) Compressor {
	switch compressType {
	case GZIP:
		return &Gzip{}
	default:
		panic("unknown compress type")
	}
}

type CrRpcClient interface {
	Connect() error
	Invoke(ctx context.Context, serviceName, method string, args []any) (any, error)
	Close() error
}

type TcpClientOption struct {
	Retries           int
	ConnectionTimeout time.Duration
	SerializerType    SerializerType
	CompressType      CompressType
	Host              string
	Port              int
}

var DefaultTcpClientOption = TcpClientOption{
	Host:              "127.0.0.1",
	Port:              9000,
	Retries:           3,
	ConnectionTimeout: 5 * time.Second,
	SerializerType:    GOB,
	CompressType:      GZIP,
}

// Protobuf 将协议设置为protobuf
func (c *TcpClientOption) Protobuf() TcpClientOption {
	c.SerializerType = PROTOBUF
	return *c
}

type TcpClient struct {
	conn   net.Conn
	option TcpClientOption
}

func NewTcpClient(option TcpClientOption) *TcpClient {
	return &TcpClient{option: option}
}

// Connect 获取链接
func (t *TcpClient) Connect() error {
	addr := fmt.Sprintf("%s:%d", t.option.Host, t.option.Port)
	conn, err := net.DialTimeout("tcp", addr, t.option.ConnectionTimeout)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

var reqId int64

// Invoke 调用RPC
func (t *TcpClient) Invoke(ctx context.Context, serviceName, method string, args []any) (any, error) {
	// 设置请求体
	req := &CrRpcRequest{}
	req.RequestId = atomic.AddInt64(&reqId, 1)
	req.ServiceName = serviceName
	req.MethodName = method
	req.Args = args
	// 设置请求头
	header := make([]byte, 17)
	header[0] = MagicNumber
	header[1] = Version
	header[6] = byte(msgRequest)
	header[7] = byte(t.option.CompressType)
	header[8] = byte(t.option.SerializerType)
	binary.BigEndian.PutUint64(header[9:], uint64(req.RequestId))
	//
	serializer := loadSerializer(t.option.SerializerType)
	var body []byte
	var err error
	if t.option.SerializerType == PROTOBUF {
		list, err := structpb.NewList(args)
		if err != nil {
			return nil, err
		}
		pReq := &Request{
			RequestId:   req.RequestId,
			ServiceName: req.ServiceName,
			MethodName:  req.MethodName,
			Args:        list.Values,
		}
		body, err = serializer.Serialize(pReq)
		fmt.Println(body)
		err = serializer.Deserialize(body, pReq)
		log.Println(err)
	} else {
		body, err = serializer.Serialize(req)
	}
	if err != nil {
		return nil, err
	}
	compressor := loadCompress(t.option.CompressType)
	body, err = compressor.Compress(body)
	if err != nil {
		return nil, err
	}
	fullLength := 17 + len(body)
	binary.BigEndian.PutUint32(header[2:6], uint32(fullLength))

	_, err = t.conn.Write(header[:])
	if err != nil {
		return nil, err
	}
	_, err = t.conn.Write(body[:])
	if err != nil {
		return nil, err
	}
	rspChan := make(chan *CrRpcResponse)
	go t.readHandle(rspChan)
	rsp := <-rspChan
	return rsp, nil
}

// 等待响应并通过通道返回数据
func (t *TcpClient) readHandle(rspChan chan *CrRpcResponse) {
	defer func() {
		if err := recover(); err != nil {
			log.Println("client readHandle", err)
			rspChan <- errorResponse(errors.New(fmt.Sprintf("%v", err)))
			_ = t.conn.Close()
		}
	}()
	for {
		msg, err := decodeFrame(t.conn)
		if err != nil {
			log.Println("no data been decode", err)
			rspChan <- errorResponse(err)
			return
		}
		if msg.Header.MessageType == msgResponse {
			if msg.Header.SerializeType == PROTOBUF {
				fmt.Printf("%+v\n", msg.Data)
				rsp := msg.Data.(*Response)
				asInterface := rsp.Data.AsInterface()
				marshal, err := json.Marshal(asInterface)
				if err != nil {
					log.Println(err)
				}
				response := &CrRpcResponse{}
				_ = json.Unmarshal(marshal, response)
				rspChan <- response
			} else {
				response := msg.Data.(*CrRpcResponse)
				rspChan <- response
			}
			return
		}
	}
}

// Close 关闭连接
func (t *TcpClient) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

type TcpClientProxy struct {
	client *TcpClient
	option TcpClientOption
}

func NewTcpClientProxy(option TcpClientOption) *TcpClientProxy {
	return &TcpClientProxy{option: option}
}

func (c *TcpClientProxy) Call(ctx context.Context, serviceName, method string, args []any) (any, error) {
	client := NewTcpClient(c.option)
	c.client = client
	err := client.Connect()
	if err != nil {
		return nil, err
	}
	for i := 0; i < c.option.Retries; i++ {
		result, err := client.Invoke(ctx, serviceName, method, args)
		if err != nil {
			if i >= c.option.Retries-1 {
				log.Println("already retry all time")
				_ = client.Close()
				return nil, err
			}
			continue
		}
		_ = client.Close()
		return result, nil
	}
	return nil, errors.New("retry time")
}
