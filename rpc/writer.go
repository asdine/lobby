package rpc

import (
	"bytes"
	"io"
)

func newPrefixWriter(prefix string, to io.Writer) *prefixWriter {
	return &prefixWriter{
		prefix: []byte(prefix),
		to:     to,
	}
}

type prefixWriter struct {
	to     io.Writer
	buf    bytes.Buffer
	prefix []byte
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	idx := bytes.IndexByte(p, '\n')
	if idx == -1 {
		return w.buf.Write(p)
	}

	n, err := w.buf.Write(p[:idx+1])
	if err != nil {
		return n, err
	}

	n, err = w.to.Write(w.prefix)
	if err != nil {
		return n, err
	}

	n, err = w.to.Write(w.buf.Bytes())
	if err != nil {
		return n, err
	}
	w.buf.Reset()

	n, err = w.buf.Write(p[idx+1:])
	if err != nil {
		return n, err
	}

	return len(p), nil
}
