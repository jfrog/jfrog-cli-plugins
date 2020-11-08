package testUtils

import "github.com/jedib0t/go-pretty/v6/table"

var PrevLineWidth int
var LinesSameWidth bool

// This is a table wrapper used to test table lines remain the same width.
type TableWrapper struct {
	*table.Table
}

func (tw *TableWrapper) AppendHeader(row table.Row, configs ...table.RowConfig) {
	updateWidth(row)
	tw.Table.AppendHeader(row, configs...)
}

func (tw *TableWrapper) AppendRow(row table.Row, configs ...table.RowConfig) {
	updateWidth(row)
	tw.Table.AppendRow(row, configs...)
}

func updateWidth(row table.Row) {
	if PrevLineWidth != 0 && PrevLineWidth != len(row) {
		LinesSameWidth = false
	}
	PrevLineWidth = len(row)
}

func ClearWidth() {
	PrevLineWidth = 0
	LinesSameWidth = false
}
