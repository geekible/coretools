package stream

import (
	"encoding/json"
	"net/http"
)

type HttpStream struct {
	w              http.ResponseWriter
	r              *http.Request
	DataStruct     *interface{}
	payload        *any
	httpStatusCode int
}

type httpErrorResponse struct {
	ResponseCode int
	Message      string
}

type HttpStreamUtil struct {
	stream *HttpStream
}

func NewHttpStreamUtil() *HttpStreamUtil {
	return &HttpStreamUtil{&HttpStream{}}
}

func (hs *HttpStreamUtil) SetResponseWriter(w http.ResponseWriter) *HttpStreamUtil {
	hs.stream.w = w
	return hs
}

func (hs *HttpStreamUtil) SetRequest(r *http.Request) *HttpStreamUtil {
	hs.stream.r = r
	return hs
}

func (hs *HttpStreamUtil) SetDataModel(data interface{}) *HttpStreamUtil {
	hs.stream.DataStruct = &data
	return hs
}

func (hs *HttpStreamUtil) SetStatusCode(statusCode int) *HttpStreamUtil {
	hs.stream.httpStatusCode = statusCode
	return hs
}

func (hs *HttpStreamUtil) SetPayload(payload any) *HttpStreamUtil {
	hs.stream.payload = &payload
	return hs
}

func (hs *HttpStreamUtil) Read() error {
	maxBytes := 1048576
	hs.stream.r.Body = http.MaxBytesReader(hs.stream.w, hs.stream.r.Body, int64(maxBytes))
	defer hs.stream.r.Body.Close()

	decoder := json.NewDecoder(hs.stream.r.Body)
	err := decoder.Decode(&hs.stream.DataStruct)
	if err != nil {
		return err
	}
	return nil
}

func (hs *HttpStreamUtil) WriteResponse() error {
	if hs.stream.httpStatusCode == 0 {
		hs.stream.httpStatusCode = http.StatusOK
	}

	out, err := json.Marshal(hs.stream.payload)
	if err != nil {
		return err
	}

	hs.stream.w.Header().Set("Content-Type", "application/json")
	hs.stream.w.WriteHeader(hs.stream.httpStatusCode)

	_, err = hs.stream.w.Write(out)
	if err != nil {
		return err
	}

	return nil
}

func (hs *HttpStreamUtil) WriteErrorResponse(message string) error {
	if hs.stream.httpStatusCode == 0 {
		hs.stream.httpStatusCode = http.StatusInternalServerError
	}
	response := httpErrorResponse{
		ResponseCode: hs.stream.httpStatusCode,
		Message:      message,
	}

	out, err := json.Marshal(response)
	if err != nil {
		return err
	}

	hs.stream.w.Header().Set("Content-Type", "application/json")
	hs.stream.w.WriteHeader(hs.stream.httpStatusCode)

	_, err = hs.stream.w.Write(out)
	if err != nil {
		return err
	}

	return nil
}
