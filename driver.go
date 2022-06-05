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
	SUCCESS uint32 = 0
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
func http_req(f_ptr *C.char, f_len uint32, opt_ptr *C.char, o_len uint32, fd *uint32) uint32 {
	var slice = (*byte)(unsafe.Pointer(f_ptr))
	var url_slice = unsafe.Slice(slice, f_len)
	var loc_url = string(url_slice)
	var req *http.Request
	var resp *http.Response
	var err error

	slice = (*byte)(unsafe.Pointer(opt_ptr))
	var opts_slice = unsafe.Slice(slice, o_len)
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
func http_read_body(fd uint32, p *C.char, len uint32, retn *uint32) uint32 {
	var slice = (*byte)(unsafe.Pointer(p))
	var bs = unsafe.Slice(slice, len)

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

//export http_read_head
func http_read_head(fd uint32, b *C.char, len uint32, retn *uint32) uint32 {

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
