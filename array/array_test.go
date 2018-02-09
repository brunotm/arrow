package array_test

import (
	"testing"

	"github.com/influxdata/arrow"
	"github.com/influxdata/arrow/array"
	"github.com/influxdata/arrow/internal/testing/tools"
	"github.com/influxdata/arrow/memory"
	"github.com/stretchr/testify/assert"
)

type testDataType struct {
	id arrow.Type
}

func (d *testDataType) ID() arrow.Type { return d.id }
func (d *testDataType) Name() string   { panic("implement me") }

func TestMakeFromData(t *testing.T) {
	tests := []struct {
		name     string
		d        arrow.DataType
		expPanic bool
		expError string
	}{
		// unsupported types
		{name: "null", d: &testDataType{arrow.NULL}, expPanic: true, expError: "unsupported data type: NULL"},
		{name: "map", d: &testDataType{arrow.MAP}, expPanic: true, expError: "unsupported data type: MAP"},

		// supported types
		{name: "bool", d: &testDataType{arrow.BOOL}},

		// invalid types
		{name: "invalid(-1)", d: &testDataType{arrow.Type(-1)}, expPanic: true, expError: "invalid data type: Type(-1)"},
		{name: "invalid(28)", d: &testDataType{arrow.Type(28)}, expPanic: true, expError: "invalid data type: Type(28)"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var b [4]*memory.Buffer
			data := array.NewData(test.d, 0, b[:], 0)

			if test.expPanic {
				assert.PanicsWithValue(t, test.expError, func() {
					array.MakeFromData(data)
				})
			} else {
				assert.NotNil(t, array.MakeFromData(data))
			}
		})
	}
}

func bbits(v ...int32) []byte {
	return tools.IntsToBitsLSB(v...)
}

func TestArray_NullN(t *testing.T) {
	tests := []struct {
		name string
		l    int
		bm   []byte
		n    int
		exp  int
	}{
		{name: "unknown,l16", l: 16, bm: bbits(0x11001010, 0x00110011), n: array.UnknownNullCount, exp: 8},
		{name: "unknown,l12,ignores last nibble", l: 12, bm: bbits(0x11001010, 0x00111111), n: array.UnknownNullCount, exp: 6},
		{name: "unknown,l12,12 nulls", l: 12, bm: bbits(0x00000000, 0x00000000), n: array.UnknownNullCount, exp: 12},
		{name: "unknown,l12,00 nulls", l: 12, bm: bbits(0x11111111, 0x11111111), n: array.UnknownNullCount, exp: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := memory.NewBufferBytes(test.bm)
			data := array.NewData(arrow.FixedWidthTypes.Boolean, test.l, []*memory.Buffer{buf, nil}, test.n)
			buf.Release()
			ar := array.MakeFromData(data)
			data.Release()
			got := ar.NullN()
			ar.Release()
			assert.Equal(t, test.exp, got)
		})
	}
}
