package main

import (
	"fmt"
	"strings"
)

type Debugger struct {
	enabled bool
	text    strings.Builder
}

func (d *Debugger) SetEnabled(enabled bool) {
	d.enabled = enabled
}

func (d *Debugger) Enabled() bool {
	return d.enabled
}

func (d *Debugger) Debug(args ...any) {
	_, _ = d.text.WriteString(fmt.Sprint(args...) + "\n")
}

func (d *Debugger) Clean() {
	d.text.Reset()
}

func (d *Debugger) String() string {
	return d.text.String()
}
