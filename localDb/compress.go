package LocalDb

import (
	"bytes"
	"compress/gzip"
	"io"
)

func compress(inp string) (error, []byte) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write([]byte(inp)); err != nil {
		return err, nil
	}
	if err := gz.Close(); err != nil {
		return err, nil
	}
	return nil, b.Bytes()
}

func uncompress(inp []byte) (error, string) {
	r := bytes.NewReader(inp)
	gz, err := gzip.NewReader(r);
	if err != nil {
		return err, ""
	}

	oup, err := io.ReadAll(gz)
	if err != nil {
		return err, ""
	}

	return nil, string(oup)
}
