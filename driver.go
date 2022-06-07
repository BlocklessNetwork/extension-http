package main

import "C"
import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"unsafe"
)

type ErrType uint32

const (
	SUCCESS uint32 = iota
	INVALID_HANDLE
	MEMORY_ACCESS_ERROR
	BUFFER_TOO_SMALL
	HEADER_NOT_FOUND
	UTF8_ERROR
	DESTINATION_NOT_ALLOWED
	INVALID_METHOD
	INVALID_ENCODING
	INVALID_URL
	REQUEST_ERROR
	RUNTIME_ERROR
	TOO_MANY_SESSIONS
	INVALID_DRIVER
)

type InnerContext struct {
	req  *http.Request
	resp *http.Response
}

var Context = make(map[uint32]*InnerContext)

var MaxSeq uint32 = 1

type Options struct {
	Method         string `json:"method"`
	ConnectTimeout int32  `json:"connectTimeout"`
	ReadTimeout    int32  `json:"readTimeout"`
}

//export http_req
func http_req(f_ptr *byte, f_len uint32, opt_ptr *byte, o_len uint32, fd *uint32) uint32 {
	var url_slice = unsafe.Slice(f_ptr, f_len)
	var loc_url = string(url_slice)
	var req *http.Request
	var resp *http.Response
	var err error

	var opts_slice = unsafe.Slice(opt_ptr, o_len)
	var options Options
	if err := json.Unmarshal(opts_slice, &options); err != nil {
		fmt.Fprintf(os.Stderr, "error format params: %s", string(opts_slice))
		return REQUEST_ERROR
	}
	if req, err = http.NewRequest(options.Method, loc_url, nil); err != nil {
		fmt.Fprintf(os.Stderr, "new request error: %s\n", err)
		return REQUEST_ERROR
	} else {
		if resp, err = http.DefaultClient.Do(req); err != nil {
			fmt.Fprintf(os.Stderr, "do request error: %s\n", err)
			return REQUEST_ERROR
		}
	}
	if len(Context) > 0 {
		MaxSeq++
	}
	Context[MaxSeq] = &InnerContext{req, resp}
	*fd = MaxSeq
	return SUCCESS
}

//export http_read_body
func http_read_body(fd uint32, p *byte, l uint32, retn *uint32) uint32 {
	if l == 0 {
		return BUFFER_TOO_SMALL
	}
	var bs []byte = unsafe.Slice(p, l)
	var ctx *InnerContext = Context[fd]
	if ctx == nil {
		return INVALID_HANDLE
	}
	if n, err := ctx.resp.Body.Read(bs); err != nil {
		if err == io.EOF {
			if n > 0 {
				*retn = uint32(n)
			} else {
				*retn = uint32(0)
			}
			return SUCCESS
		}
		fmt.Fprintf(os.Stderr, "read body error: %s\n", err)
		return RUNTIME_ERROR
	} else {
		*retn = uint32(n)
		return SUCCESS
	}
}

//export http_read_header
func http_read_header(fd uint32, h_ptr *byte, h_len uint32, buf_ptr *byte, buf_len uint32, retn *uint32) uint32 {
	var header = string(unsafe.Slice(h_ptr, h_len))
	var buf = unsafe.Slice(buf_ptr, buf_len)
	var ctx *InnerContext = Context[fd]
	if ctx == nil {
		return INVALID_HANDLE
	}
	headVal := ctx.resp.Request.Header.Get(header)
	if headVal == "" {
		return HEADER_NOT_FOUND
	}
	headValBuf := []byte(headVal)
	n := copy(buf, headValBuf)
	*retn = uint32(n)
	return SUCCESS
}

//export http_close
func http_close(fd uint32) uint32 {
	var ctx *InnerContext = Context[fd]
	if ctx == nil {
		return INVALID_HANDLE
	}
	if ctx.resp != nil {
		ctx.resp.Body.Close()
	}
	delete(Context, fd)
	return SUCCESS
}

func main() {

}
