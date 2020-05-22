package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

const (
	columnCreated      = "CREATED"
	columnState        = "STATE"
	columnName         = "NAME"
	columnPID          = "PID"
	columnDirectory    = "DIRECTORY"
	columnInstanceID   = "INSTANCE ID"
	columnIP           = "IP"
)

// display use to output something on screen with table format.
type display struct {
	w *tabwriter.Writer
}

//creates a display instance
func newTableDisplay(minwidth, tabwidth, padding int, padchar byte, flags uint) *display {
	w := tabwriter.NewWriter(os.Stdout, minwidth, tabwidth, padding, padchar, 0)
	return &display{w}
}

// add a row of data.
func (d *display) AddRow(row []string) {
	fmt.Fprintln(d.w, strings.Join(row, "\t"))
}

// output on screen.
func (d *display) Flush() error {
	return d.w.Flush()
}

