package gzip

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func Compress(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	gw := gzip.NewWriter(buf)
	_, err := gw.Write(data)
	if err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(ioutil.NopCloser(bytes.NewBuffer(data)))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return ioutil.ReadAll(gr)
}
