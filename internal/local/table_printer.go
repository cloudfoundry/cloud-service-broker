package local

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

type tablePrinter struct {
	w *tabwriter.Writer
}

func newTablePrinter(headings ...string) *tablePrinter {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.StripEscape)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, strings.Join(headings, "\t"))
	_, _ = fmt.Fprintln(w, strings.Join(mapSlice(headings, toDivider), "\t"))
	return &tablePrinter{w: w}
}

func (t *tablePrinter) row(data ...string) {
	_, _ = fmt.Fprintf(t.w, "%s\n", strings.Join(data, "\t"))
}

func (t *tablePrinter) print() {
	_, _ = fmt.Fprintln(t.w)
	_ = t.w.Flush()
}

func toDivider(input string) string {
	return strings.Repeat("-", len(input))
}

func mapSlice[A any](input []A, cb func(A) A) (result []A) {
	for _, e := range input {
		result = append(result, cb(e))
	}
	return
}
