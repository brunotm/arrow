package array

import (
	"sync/atomic"

	"github.com/influxdata/arrow"
	"github.com/influxdata/arrow/internal/bitutil"
	"github.com/influxdata/arrow/internal/debug"
	"github.com/influxdata/arrow/memory"
)

type BooleanBuilder struct {
	builder

	data    *memory.Buffer
	rawData []byte
}

func NewBooleanBuilder(mem memory.Allocator) *BooleanBuilder {
	return &BooleanBuilder{builder: builder{refCount: 1, mem: mem}}
}

// Release decreases the reference count by 1.
// When the reference count goes to zero, the memory is freed.
// Release may be called simultaneously from multiple goroutines.
func (b *BooleanBuilder) Release() {
	debug.Assert(atomic.LoadInt64(&b.refCount) > 0, "too many releases")

	if atomic.AddInt64(&b.refCount, -1) == 0 {
		if b.nullBitmap != nil {
			b.nullBitmap.Release()
			b.nullBitmap = nil
		}
		if b.data != nil {
			b.data.Release()
			b.data = nil
		}
	}
}

func (b *BooleanBuilder) Append(v bool) {
	b.Reserve(1)
	b.UnsafeAppend(v)
}

func (b *BooleanBuilder) AppendByte(v byte) {
	b.Reserve(1)
	b.UnsafeAppend(v != 0)
}

func (b *BooleanBuilder) AppendNull() {
	b.Reserve(1)
	b.UnsafeAppendBoolToBitmap(false)
}

func (b *BooleanBuilder) UnsafeAppend(v bool) {
	bitutil.SetBit(b.nullBitmap.Bytes(), b.length)
	if v {
		bitutil.SetBit(b.rawData, b.length)
	} else {
		bitutil.ClearBit(b.rawData, b.length)
	}
	b.length++
}

func (b *BooleanBuilder) AppendValues(v []bool, valid []bool) {
	if len(v) != len(valid) && len(valid) != 0 {
		panic("len(v) != len(valid) && len(valid) != 0")
	}

	b.Reserve(len(v))
	for i, vv := range v {
		bitutil.SetBitTo(b.rawData, b.length+i, vv)
	}
	b.builder.unsafeAppendBoolsToBitmap(valid, len(v))
}

func (b *BooleanBuilder) init(capacity int) {
	b.builder.init(capacity)

	b.data = memory.NewResizableBuffer(b.mem)
	bytesN := arrow.BooleanTraits.BytesRequired(capacity)
	b.data.Resize(bytesN)
	b.rawData = b.data.Bytes()
}

// Reserve ensures there is enough space for appending n elements
// by checking the capacity and calling Resize if necessary.
func (b *BooleanBuilder) Reserve(n int) {
	b.builder.reserve(n, b.Resize)
}

// Resize adjusts the space allocated by b to n elements. If n is greater than b.Cap(),
// additional memory will be allocated. If n is smaller, the allocated memory may reduced.
func (b *BooleanBuilder) Resize(n int) {
	if n < minBuilderCapacity {
		n = minBuilderCapacity
	}

	if b.capacity == 0 {
		b.init(n)
	} else {
		b.builder.resize(n, b.init)
		b.data.Resize(arrow.BooleanTraits.BytesRequired(n))
		b.rawData = b.data.Bytes()
	}
}

func (b *BooleanBuilder) Finish() (a *Boolean) {
	data := b.finishInternal()
	a = NewBooleanData(data)
	data.Release()
	return
}

func (b *BooleanBuilder) finishInternal() *Data {
	bytesRequired := arrow.BooleanTraits.BytesRequired(b.length)
	if bytesRequired > 0 && bytesRequired < b.data.Len() {
		// trim buffers
		b.data.Resize(bytesRequired)
	}
	res := NewData(arrow.FixedWidthTypes.Boolean, b.length, []*memory.Buffer{b.nullBitmap, b.data}, b.nullN)
	b.reset()

	b.data.Release()
	b.data = nil
	b.rawData = nil

	return res
}
