package lobby

import (
	"bytes"
	"io"
)

// NewPrefixWriter creates a PrefixWriter.
func NewPrefixWriter(prefix string, to io.Writer) *PrefixWriter {
	return &PrefixWriter{
		prefix: []byte(prefix),
		to:     to,
	}
}

// NewFuncPrefixWriter creates a PrefixWriter that uses a function to generate a prefix.
func NewFuncPrefixWriter(prefixFn func() []byte, to io.Writer) *PrefixWriter {
	return &PrefixWriter{
		prefixFn: prefixFn,
		to:       to,
	}
}

// PrefixWriter is a writer that adds a prefix before every line.
type PrefixWriter struct {
	to       io.Writer
	buf      bytes.Buffer
	prefix   []byte
	prefixFn func() []byte
}

func (w *PrefixWriter) Write(p []byte) (int, error) {
	lenp := len(p)

	idx := bytes.IndexByte(p, '\n')
	if idx == -1 {
		return w.buf.Write(p)
	}

	for idx != -1 {
		if idx == 0 && w.buf.Len() == 0 {
			p = p[1:]
			idx = bytes.IndexByte(p, '\n')
			continue
		}

		idx++

		n, err := w.buf.Write(p[:idx])
		if err != nil {
			return n, err
		}

		if w.prefixFn != nil {
			n, err = w.to.Write(w.prefixFn())
		} else {
			n, err = w.to.Write(w.prefix)
		}
		if err != nil {
			return n, err
		}

		n, err = w.to.Write(w.buf.Bytes())
		if err != nil {
			return n, err
		}
		w.buf.Reset()

		p = p[idx:]
		idx = bytes.IndexByte(p, '\n')
	}

	if len(p) != 0 {
		n, err := w.buf.Write(p)
		if err != nil {
			return n, err
		}
	}

	return lenp, nil
}
