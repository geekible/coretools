package stream

import (
	"bytes"
	"encoding/gob"
	"net"
	"net/http"
)

type TcpStream struct {
	connection net.Conn
	Data       interface{}
	statusCode int
	payload    interface{}
}

type tcpResponse struct {
	ResponseCode int         `json:"response_code"`
	Message      string      `json:"message,omitempty"`
	Data         interface{} `json:"data,omitempty"`
}

type TcpStreamUtil struct {
	stream *TcpStream
}

func NewTcpStreamUtil() *TcpStreamUtil {
	return &TcpStreamUtil{&TcpStream{}}
}

func (ts *TcpStreamUtil) SetConn(conn net.Conn) *TcpStreamUtil {
	ts.stream.connection = conn
	return ts
}

func (ts *TcpStreamUtil) SetStatusCode(statusCode int) *TcpStreamUtil {
	ts.stream.statusCode = statusCode
	return ts
}

func (ts *TcpStreamUtil) SetDataModel(data interface{}) *TcpStreamUtil {
	ts.stream.Data = &data
	return ts
}

func (ts *TcpStreamUtil) SetPayload(payload interface{}) *TcpStreamUtil {
	ts.stream.payload = payload
	return ts
}

func (ts *TcpStreamUtil) Read() error {
	tmp := make([]byte, 1024)
	_, err := ts.stream.connection.Read(tmp)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(tmp)
	gobobj := gob.NewDecoder(buf)
	err = gobobj.Decode(&ts.stream.Data)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TcpStreamUtil) WriteResponse() error {
	defer ts.stream.connection.Close()

	if ts.stream.statusCode == 0 {
		ts.stream.statusCode = http.StatusOK
	}

	response := tcpResponse{
		ResponseCode: ts.stream.statusCode,
		Message:      "Ok",
		Data:         ts.stream.payload,
	}

	buf := new(bytes.Buffer)
	gobobj := gob.NewEncoder(buf)
	err := gobobj.Encode(&response)
	if err != nil {
		return err
	}

	_, err = ts.stream.connection.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (ts *TcpStreamUtil) WriteErrorResponse() error {
	defer ts.stream.connection.Close()

	if ts.stream.statusCode == 0 {
		ts.stream.statusCode = http.StatusInternalServerError
	}

	response := tcpResponse{
		ResponseCode: ts.stream.statusCode,
		Message:      "Error",
	}

	buf := new(bytes.Buffer)
	gobobj := gob.NewEncoder(buf)
	err := gobobj.Encode(&response)
	if err != nil {
		return err
	}

	_, err = ts.stream.connection.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
