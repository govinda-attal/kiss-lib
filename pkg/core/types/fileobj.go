package types

import (
	"io"
	"strings"
)

// FileObj is custom type that represents a file data in memory. This type was defined to facilitate reading of file data passed within multipart/form-data.
// Once the file read from the http request within an instance of this type, it can be used within the application logic for relevant purpose.
// Example can be to load file data into S3 or Google Cloud Storage.
type FileObj struct {
	name   string
	size   int64
	reader io.Reader
}

// NewFileObj returns an instance of type FileObj which is simply aggregation of the input parameters.
// Reader can be bytes.Reader to represent unstructured data in bytes.
func NewFileObj(name string, size int64, reader io.Reader) *FileObj {
	return &FileObj{name, size, reader}
}

// Name returns name of the file.
func (f FileObj) Name() string {
	return f.name
}

// Size returns size of file in number of bytes.
func (f FileObj) Size() int64 {
	return f.size
}

// Reader is used to read the file data
func (f FileObj) Reader() io.Reader {
	return f.reader
}

// Ext returns file extension from the file name if available.
// There is no intention to verify or read the file data to return correct extension of the file.
func (f FileObj) Ext() string {
	if i := strings.LastIndex(f.name, "."); i > -1 {
		return f.name[i+1 : len(f.name)]
	}
	return ""
}
