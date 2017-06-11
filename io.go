package lobby

import (
	"bytes"
	"io"
)

// NewPrefixWriter creates PrefixWriter.
func NewPrefixWriter(prefix string, to io.Writer) *PrefixWriter {
	return &PrefixWriter{
		prefix: []byte(prefix),
		to:     to,
	}
}

// PrefixWriter is a writer that adds a prefix before every line.
type PrefixWriter struct {
	to     io.Writer
	buf    bytes.Buffer
	prefix []byte
}

func (w *PrefixWriter) Write(p []byte) (int, error) {
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
