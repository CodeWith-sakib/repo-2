package cli

import (
	"fmt"
	"io"
	"strings"
)

type SimpleTableWriter struct {
	out io.Writer
}

func NewSimpleTableWriter(out io.Writer) *SimpleTableWriter {
	return &SimpleTableWriter{out: out}
}

func (w *SimpleTableWriter) RenderHeader(cols ...string) {
	fmt.Fprintln(w.out, strings.Join(cols, "	"))
}

func (w *SimpleTableWriter) RenderRow(vals ...string) {
	fmt.Fprintln(w.out, strings.Join(vals, "	"))
}
