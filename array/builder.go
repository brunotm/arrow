package array

import (
	"github.com/influxdata/arrow/internal/bitutil"
	"github.com/influxdata/arrow/memory"
)

const (
	minBuilderCapacity = 1 << 5
)

// builder provides common functionality for managing the validity bitmap (nulls) when building arrays.
type builder struct {
	pool       memory.Allocator
	nullBitmap *memory.PoolBuffer
	nullN      int
	length     int
	capacity   int
}

// Len returns the number of elements in the array builder.
func (b *builder) Len() int { return b.length }

// Cap returns the total number of elements that can be stored without allocating additional memory.
func (b *builder) Cap() int { return b.capacity }

// NullN returns the number of null values in the array builder.
func (b *builder) NullN() int { return b.nullN }

func (b *builder) init(capacity int) {
	toAlloc := bitutil.CeilByte(capacity) / 8
	b.nullBitmap = memory.NewPoolBuffer(b.pool)
	b.nullBitmap.Resize(toAlloc)
	b.capacity = capacity
	memory.Set(b.nullBitmap.Buf(), 0)
}

func (b *builder) reset() {
	b.nullBitmap = nil
	b.nullN = 0
	b.length = 0
	b.capacity = 0
}

func (b *builder) resize(newBits int, init func(int)) {
	if b.nullBitmap == nil {
		init(newBits)
		return
	}

	newBytesN := bitutil.CeilByte(newBits) / 8
	oldBytesN := b.nullBitmap.Len()
	b.nullBitmap.Resize(newBytesN)
	b.capacity = newBits
	if oldBytesN < newBytesN {
		// TODO(sgc): necessary?
		memory.Set(b.nullBitmap.Buf()[oldBytesN:], 0)
	}
}

func (b *builder) reserve(elements int, resize func(int)) {
	if b.length+elements > b.capacity {
		newCap := bitutil.NextPowerOf2(b.length + elements)
		resize(newCap)
	}
}

// unsafeAppendBoolsToBitmap appends the contents of valid to the validity bitmap.
// As an optimization, if the valid slice is empty, the next length bits will be set to valid (not null).
func (b *builder) unsafeAppendBoolsToBitmap(valid []bool, length int) {
	if len(valid) == 0 {
		b.unsafeSetValid(length)
		return
	}

	byteOffset := b.length / 8
	bitOffset := byte(b.length % 8)
	nullBitmap := b.nullBitmap.Bytes()
	bitSet := nullBitmap[byteOffset]

	for _, v := range valid {
		if bitOffset == 8 {
			bitOffset = 0
			nullBitmap[byteOffset] = bitSet
			byteOffset++
			bitSet = nullBitmap[byteOffset]
		}

		if v {
			bitSet |= bitutil.BitMask[bitOffset]
		} else {
			bitSet &= bitutil.FlippedBitMask[bitOffset]
			b.nullN++
		}
		bitOffset++
	}

	if bitOffset != 0 {
		nullBitmap[byteOffset] = bitSet
	}
	b.length += len(valid)
}

// unsafeSetValid sets the next length bits to valid in the validity bitmap.
func (b *builder) unsafeSetValid(length int) {
	padToByte := min(8-(b.length%8), length)
	if padToByte == 8 {
		padToByte = 0
	}
	bits := b.nullBitmap.Bytes()
	for i := b.length; i < b.length+padToByte; i++ {
		bitutil.SetBit(bits, i)
	}

	start := (b.length + padToByte) / 8
	fastLength := (length - padToByte) / 8
	memory.Set(bits[start:start+fastLength], 0xff)

	newLength := b.length + length
	// trailing bytes
	for i := b.length + padToByte + (fastLength * 8); i < newLength; i++ {
		bitutil.SetBit(bits, i)
	}

	b.length = newLength
}

func (b *builder) UnsafeAppendBoolToBitmap(isValid bool) {
	if isValid {
		bitutil.SetBit(b.nullBitmap.Bytes(), b.length)
	} else {
		b.nullN++
	}
	b.length++
}
