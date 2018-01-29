// Code generated by array/bufferbuilder_numeric.gen.go.tmpl.
// DO NOT EDIT.

package array

import (
	"github.com/influxdata/arrow"
	"github.com/influxdata/arrow/internal/bitutil"
	"github.com/influxdata/arrow/memory"
)

type Int32BufferBuilder struct {
	bufferBuilder
}

func NewInt32BufferBuilder(mem memory.Allocator) *Int32BufferBuilder {
	return &Int32BufferBuilder{bufferBuilder: bufferBuilder{mem: mem}}
}

// AppendValues appends the contents of v to the buffer, growing the buffer as needed.
func (b *Int32BufferBuilder) AppendValues(v []int32) { b.Append(arrow.Int32Traits{}.CastToBytes(v)) }

// Values returns a slice of length b.Len().
// The slice is only valid for use until the next buffer modification. That is, until the next call
// to Advance, Reset, Finish or any Append function. The slice aliases the buffer content at least until the next
// buffer modification.
func (b *Int32BufferBuilder) Values() []int32 { return arrow.Int32Traits{}.CastFromBytes(b.Bytes()) }

// Value returns the int32 element at the index i. Value will panic if i is negative or ≥ Len.
func (b *Int32BufferBuilder) Value(i int) int32 { return b.Values()[i] }

// Len returns the number of int32 elements in the buffer.
func (b *Int32BufferBuilder) Len() int { return b.length / arrow.Int32SizeBytes }

// AppendValue appends v to the buffer, growing the buffer as needed.
func (b *Int32BufferBuilder) AppendValue(v int32) {
	if b.capacity < b.length+arrow.Int32SizeBytes {
		newCapacity := bitutil.NextPowerOf2(b.length + arrow.Int32SizeBytes)
		b.resize(newCapacity)
	}
	arrow.Int32Traits{}.PutValue(b.bytes[b.length:], v)
	b.length += arrow.Int32SizeBytes
}
