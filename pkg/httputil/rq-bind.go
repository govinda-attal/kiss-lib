package httputil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/govinda-attal/kiss-lib/pkg/core/status"
	"github.com/govinda-attal/kiss-lib/pkg/core/types"
)

// RqBind method populates a given variable with data passed in HTTP Request Body. This method as of now supports following binders:
//
// 1) JSONBind: To be used when Content-Type of HTTP Request is 'application/json'
//
// 2) TxtBind: To be used when Content-Type of HTTP Request is 'text/plain'
//
// 3) MPFormBind: To be used when Content-Type of HTTP Request is 'multipart/form-data'
//
func RqBind(r *http.Request, b Binder) error {

	if !b.ContentCompatible(r.Header.Get("Content-Type")) {
		return status.ErrContentTypeNotSupported()
	}
	if err := b.Bind(r); err != nil {
		return err
	}
	return nil
}

// Binder interface allows binding/populating a variable from the HTTP request body.
type Binder interface {
	// Bind populates variable from the HTTP Request body.
	Bind(r *http.Request) error
	// ContentType returns http header Content-Type of the concrete binder implementation.
	ContentType() string
	// ContentCompatible returns true or false based on the input http header Content-Type.
	ContentCompatible(contentType string) bool
	// Data returns bound data
	Data() interface{}
}

// JSONBind populates given variable 'd' from application/json HTTP request body.
func JSONBind(d interface{}) Binder {
	return jsonBind{data: d}
}

// TxtBind populates given variable 'd' from text/plain HTTP request body.
func TxtBind(d interface{}) Binder {
	return txtBind{data: d}
}

// MPFormBind populates given map of variable(s) and files from multipart/form-data HTTP request body.
// Size of the form-data read is restricted to size 's'.
// For now only text and JSON variables are supported for structured multipart/form-data.
// Unstructured data like files will be read into map of type types.FileObj.
func MPFormBind(d map[string]interface{}, f map[string]*types.FileObj, s int64) *mpFormBind {
	return &mpFormBind{dataM: d, fileM: f, size: s}
}

type jsonBind struct {
	data interface{}
}

func (jb jsonBind) ContentType() string {
	return "application/json"
}

func (jb jsonBind) ContentCompatible(contentType string) bool {
	if !strings.Contains(contentType, jb.ContentType()) {
		return false
	}
	return true
}

func (jb jsonBind) Bind(r *http.Request) error {
	if err := json.NewDecoder(r.Body).Decode(jb.data); err != nil {
		return status.ErrBadRequest().WithError(err)
	}
	return nil
}

func (jb jsonBind) Data() interface{} {
	return jb.data
}

type txtBind struct {
	data interface{}
}

func (tb txtBind) ContentType() string {
	return "text/plain"
}

func (tb txtBind) ContentCompatible(contentType string) bool {
	if !strings.Contains(contentType, tb.ContentType()) {
		return false
	}
	return true
}

func (tb txtBind) Bind(r *http.Request) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return status.ErrBadRequest().WithError(err)
	}
	tb.data = string(body)
	return nil
}

func (tb txtBind) Data() interface{} {
	return tb.data
}

type mpFormBind struct {
	dataM map[string]interface{}
	fileM map[string]*types.FileObj
	size  int64
}

func (mf *mpFormBind) ContentType() string {
	return "multipart/form-data"
}

func (mf *mpFormBind) ContentCompatible(contentType string) bool {
	if !strings.Contains(contentType, mf.ContentType()) {
		return false
	}
	return true
}

func (mf *mpFormBind) Bind(r *http.Request) error {
	if err := r.ParseMultipartForm(mf.size); err != nil {
		return status.ErrInternal().WithError(err)
	}
	if r.MultipartForm.File != nil {
		for k, _ := range mf.fileM {
			fh, ok := r.MultipartForm.File[k]
			if !ok {
				continue
				//return status.ErrBadRequest().WithError(fmt.Errorf("%s file missing in multi-part message in the http request", k))
			}
			f, err := fh[0].Open()
			defer f.Close()
			if err != nil {
				return status.ErrInternal().WithError(err)
			}
			b, err := ioutil.ReadAll(f)
			if err != nil {
				return status.ErrBadRequest().WithError(err)
			}
			mf.fileM[k] = types.NewFileObj(fh[0].Filename, fh[0].Size, bytes.NewReader(b))
		}
	}
	if r.MultipartForm.Value != nil {
		for k, v := range mf.dataM {
			data, ok := r.MultipartForm.Value[k]
			if !ok {
				continue
				//return status.ErrBadRequest().WithError(fmt.Errorf("%s part missing in multi-part message in the http request", k))
			}
			var err error
			switch v.(type) {
			case string:
				mf.dataM[k] = data[0]
			default:
				err = json.Unmarshal([]byte(data[0]), v)
			}
			if err != nil {
				return status.ErrBadRequest().WithError(fmt.Errorf("error un-marshaling %s part in multi-part message in the http request", k))
			}
		}
	}
	return nil
}

func (mf *mpFormBind) Data() interface{} {
	return []interface{}{mf.dataM, mf.fileM}
}

func FileBind(fname string, file *types.FileObj, contTypes ...string) *fileBind {
	fb := &fileBind{fname: fname, contTypes: contTypes}
	if len(fb.contTypes) == 0 {
		fb.contTypes = append(fb.contTypes, "application/binary")
	}
	return fb
}

type fileBind struct {
	fname     string
	file      *types.FileObj
	contTypes []string
}

func (fb *fileBind) ContentType() string {
	return "multipart/form-data"
}

func (fb *fileBind) ContentCompatible(contentType string) bool {
	for _, acceptableContentType := range fb.contTypes {
		if strings.Contains(contentType, acceptableContentType) {
			return true
		}
	}
	return false
}

func (fb *fileBind) Data() interface{} {
	return fb.file
}

func (fb *fileBind) Bind(r *http.Request) error {
	contType := r.Header.Get("Content-type")
	fname := fb.fname
	if !strings.Contains(contType, "application/binary") {
		ext := strings.Split(contType, "/")[1]
		fname = strings.Split(fname, ".")[0]
		fname = strings.Join([]string{fname, ext}, ".")
	}
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return status.ErrBadRequest().WithError(err)
	}
	fb.file = types.NewFileObj(fname, int64(len(b)), bytes.NewReader(b))
	return nil
}
