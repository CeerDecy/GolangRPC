package rpc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
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
	//TODO implement me
	panic("implement me")
}

// Deserialize Protobuf反序列化
func (p *Protobuf) Deserialize(data []byte, target any) error {
	//TODO implement me
	panic("implement me")
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
	if rsp.Code == 200 {

	}
	headers := make([]byte, 17)
	//magic number
	headers[0] = MagicNumber
	//version
	headers[17] = Version
	//full length
	//消息类型
	headers[6] = byte(msgResponse)
	headers[7] = byte(rsp.CompressType)
	headers[8] = byte(rsp.SerializerType)
	binary.BigEndian.PutUint64(headers[9:], uint64(rsp.RequestId))
	// 编码 先序列化 再压缩
	serializer := loadSerializer(rsp.SerializerType)
	body, err := serializer.Serialize(rsp.Data)
	if err != nil {
		return err
	}
	compress := loadCompress(rsp.CompressType)
	body, err = compress.Compress(body)
	if err != nil {
		return err
	}
	_, err = c.conn.Write(headers[:])
	if err != nil {
		return err
	}
	_, err = c.conn.Write(body)
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
	// 接收数据
	msg, err := t.decodeFrame(conn)
	if err != nil {
		rsp := &CrRpcResponse{
			Code: 500,
			Msg:  err.Error(),
		}
		conn.rpcChan <- rsp
		return err
	}
	if msg.Header.MessageType == msgRequest {
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
		conn.rpcChan <- rsp
	}
	return err
}

// 发送数据
func (t *TcpRpcServer) writeHandle(conn *TcpConn) {
	select {
	case rsp := <-conn.rpcChan:
		defer conn.conn.Close()
		//发送数据
		err := conn.Send(rsp)
		log.Println(err)
	}
}

func (t *TcpRpcServer) decodeFrame(conn *TcpConn) (*CrRpcMessage, error) {
	headers := make([]byte, 17)
	_, err := io.ReadFull(conn.conn, headers)
	if err != nil {
		return nil, err
	}
	if headers[0] != MagicNumber {
		rsp := &CrRpcResponse{
			Code: 500,
			Msg:  errors.New("magic number error").Error(),
		}
		conn.rpcChan <- rsp
		return nil, errors.New("magic number error")
	}

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
	body := make([]byte, bodyLen)
	_, err = io.ReadFull(conn.conn, body)
	if err != nil {
		rsp := &CrRpcResponse{
			Code: 500,
			Msg:  err.Error(),
		}
		conn.rpcChan <- rsp
		return nil, err
	}
	// 编码 ： 先序列化 后压缩
	// 解码 ： 先解压 后反序列化
	compress := loadCompress(msg.Header.CompressType)
	body, err = compress.Compress(body)
	serializer := loadSerializer(msg.Header.SerializeType)
	if msg.Header.MessageType == msgRequest {
		req := &CrRpcRequest{}
		err = serializer.Deserialize(body, req)
		if err != nil {
			rsp := &CrRpcResponse{
				Code: 500,
				Msg:  err.Error(),
			}
			conn.rpcChan <- rsp
			return nil, err
		}
		msg.Data = req
		return msg, nil
	}
	if msg.Header.MessageType == msgResponse {
		rsp := &CrRpcResponse{}
		err = serializer.Deserialize(body, rsp)
		if err != nil {
			rsp := &CrRpcResponse{
				Code: 500,
				Msg:  err.Error(),
			}
			conn.rpcChan <- rsp
			return nil, err
		}
		msg.Data = rsp
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
	Invoke(ctx context.Context, serviceName, method string, args map[string]any) (any, error)
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

var DefaultTcpClientOption = &TcpClientOption{
	Host:              "127.0.0.1",
	Port:              9000,
	Retries:           3,
	ConnectionTimeout: 5 * time.Second,
	SerializerType:    GOB,
	CompressType:      GZIP,
}

type TcpClient struct {
	conn   net.Conn
	option TcpClientOption
}

func NewTcpClient(option TcpClientOption) *TcpClient {
	return &TcpClient{option: option}
}

// Connect 获取🔗
func (t *TcpClient) Connect() error {
	addr := fmt.Sprintf("%s:%d", t.option.Host, t.option.Port)
	conn, err := net.DialTimeout("tcp", addr, t.option.ConnectionTimeout)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

func (t *TcpClient) Invoke(ctx context.Context, serviceName, method string, args map[string]any) (any, error) {
	//TODO implement me
	panic("implement me")
}

func (t *TcpClient) Close() error {
	//TODO implement me
	panic("implement me")
}
