package LocalDb

import (
	"bytes"
	"compress/gzip"
	"io"
)

func compress(inp string) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(inp)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func uncompress(inp []byte) (string, error) {
	r := bytes.NewReader(inp)
	gz, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}

	oup, err := io.ReadAll(gz)
	if err != nil {
		return "", err
	}

	return string(oup), nil
}
