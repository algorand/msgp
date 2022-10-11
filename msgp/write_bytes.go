package msgp

import (
	"math"
	"time"
)

// ensure 'sz' extra bytes in 'b' btw len(b) and cap(b)
// If the growth length overflows, we are anyway running
// out of memory, so panic (on a subsequent out-of-bounds
// slice reference) seems like as good of a result as any.
func ensure(b []byte, sz int) ([]byte, int) {
	l := len(b)
	c := cap(b)
	if c-l < sz {
		o := make([]byte, (2*c)+sz) // exponential growth
		n := copy(o, b)
		return o[:n+sz], n
	}
	return b[:l+sz], l
}

// AppendMapHeader appends a map header with the
// given size to the slice
func AppendMapHeader(b []byte, sz uint32) []byte {
	switch {
	case sz <= 15:
		return append(b, wfixmap(uint8(sz)))

	case sz <= math.MaxUint16:
		o, n := ensure(b, 3)
		prefixu16(o[n:], mmap16, uint16(sz))
		return o

	default:
		o, n := ensure(b, 5)
		prefixu32(o[n:], mmap32, sz)
		return o
	}
}

// AppendArrayHeader appends an array header with
// the given size to the slice
func AppendArrayHeader(b []byte, sz uint32) []byte {
	switch {
	case sz <= 15:
		return append(b, wfixarray(uint8(sz)))

	case sz <= math.MaxUint16:
		o, n := ensure(b, 3)
		prefixu16(o[n:], marray16, uint16(sz))
		return o

	default:
		o, n := ensure(b, 5)
		prefixu32(o[n:], marray32, sz)
		return o
	}
}

// AppendNil appends a 'nil' byte to the slice
func AppendNil(b []byte) []byte { return append(b, mnil) }

// AppendFloat64 appends a float64 to the slice
func AppendFloat64(b []byte, f float64) []byte {
	o, n := ensure(b, Float64Size)
	prefixu64(o[n:], mfloat64, math.Float64bits(f))
	return o
}

// AppendFloat32 appends a float32 to the slice
func AppendFloat32(b []byte, f float32) []byte {
	o, n := ensure(b, Float32Size)
	prefixu32(o[n:], mfloat32, math.Float32bits(f))
	return o
}

// AppendDuration appends a time.Duration to the slice
func AppendDuration(b []byte, d time.Duration) []byte {
	return AppendInt64(b, int64(d))
}

// AppendInt64 appends an int64 to the slice
func AppendInt64(b []byte, i int64) []byte {
	if i >= 0 {
		return AppendUint64(b, uint64(i))
	}
	switch {
	case i >= -32:
		return append(b, wnfixint(int8(i)))
	case i >= math.MinInt8:
		o, n := ensure(b, 2)
		putMint8(o[n:], int8(i))
		return o
	case i >= math.MinInt16:
		o, n := ensure(b, 3)
		putMint16(o[n:], int16(i))
		return o
	case i >= math.MinInt32:
		o, n := ensure(b, 5)
		putMint32(o[n:], int32(i))
		return o
	default:
		o, n := ensure(b, 9)
		putMint64(o[n:], i)
		return o
	}
}

// AppendInt8 appends an int8 to the slice
func AppendInt8(b []byte, i int8) []byte { return AppendInt64(b, int64(i)) }

// AppendInt16 appends an int16 to the slice
func AppendInt16(b []byte, i int16) []byte { return AppendInt64(b, int64(i)) }

// AppendInt32 appends an int32 to the slice
func AppendInt32(b []byte, i int32) []byte { return AppendInt64(b, int64(i)) }

// AppendUint64 appends a uint64 to the slice
func AppendUint64(b []byte, u uint64) []byte {
	switch {
	case u <= (1<<7)-1:
		return append(b, wfixint(uint8(u)))

	case u <= math.MaxUint8:
		o, n := ensure(b, 2)
		putMuint8(o[n:], uint8(u))
		return o

	case u <= math.MaxUint16:
		o, n := ensure(b, 3)
		putMuint16(o[n:], uint16(u))
		return o

	case u <= math.MaxUint32:
		o, n := ensure(b, 5)
		putMuint32(o[n:], uint32(u))
		return o

	default:
		o, n := ensure(b, 9)
		putMuint64(o[n:], u)
		return o

	}
}

// AppendUint8 appends a uint8 to the slice
func AppendUint8(b []byte, u uint8) []byte { return AppendUint64(b, uint64(u)) }

// AppendByte is analogous to AppendUint8
func AppendByte(b []byte, u byte) []byte { return AppendUint8(b, uint8(u)) }

// AppendUint16 appends a uint16 to the slice
func AppendUint16(b []byte, u uint16) []byte { return AppendUint64(b, uint64(u)) }

// AppendUint32 appends a uint32 to the slice
func AppendUint32(b []byte, u uint32) []byte { return AppendUint64(b, uint64(u)) }

// AppendBytes appends bytes to the slice as MessagePack 'bin' data
func AppendBytes(b []byte, bts []byte) []byte {
	sz := len(bts)
	var o []byte
	var n int
	switch {
	case bts == nil:
		o, n = ensure(b, 1)
		o[n] = mnil
		n += 1
	case sz <= math.MaxUint8:
		o, n = ensure(b, 2+sz)
		prefixu8(o[n:], mbin8, uint8(sz))
		n += 2
	case sz <= math.MaxUint16:
		o, n = ensure(b, 3+sz)
		prefixu16(o[n:], mbin16, uint16(sz))
		n += 3
	default:
		o, n = ensure(b, 5+sz)
		prefixu32(o[n:], mbin32, uint32(sz))
		n += 5
	}
	return o[:n+copy(o[n:], bts)]
}

// AppendBool appends a bool to the slice
func AppendBool(b []byte, t bool) []byte {
	if t {
		return append(b, mtrue)
	}
	return append(b, mfalse)
}

// AppendString appends a string as a MessagePack 'str' to the slice
func AppendString(b []byte, s string) []byte {
	sz := len(s)
	var n int
	var o []byte
	switch {
	case sz <= 31:
		o, n = ensure(b, 1+sz)
		o[n] = wfixstr(uint8(sz))
		n++
	case sz <= math.MaxUint8:
		o, n = ensure(b, 2+sz)
		prefixu8(o[n:], mstr8, uint8(sz))
		n += 2
	case sz <= math.MaxUint16:
		o, n = ensure(b, 3+sz)
		prefixu16(o[n:], mstr16, uint16(sz))
		n += 3
	default:
		o, n = ensure(b, 5+sz)
		prefixu32(o[n:], mstr32, uint32(sz))
		n += 5
	}
	return o[:n+copy(o[n:], s)]
}

// AppendStringFromBytes appends a []byte
// as a MessagePack 'str' to the slice 'b.'
func AppendStringFromBytes(b []byte, str []byte) []byte {
	sz := len(str)
	var n int
	var o []byte
	switch {
	case sz <= 31:
		o, n = ensure(b, 1+sz)
		o[n] = wfixstr(uint8(sz))
		n++
	case sz <= math.MaxUint8:
		o, n = ensure(b, 2+sz)
		prefixu8(o[n:], mstr8, uint8(sz))
		n += 2
	case sz <= math.MaxUint16:
		o, n = ensure(b, 3+sz)
		prefixu16(o[n:], mstr16, uint16(sz))
		n += 3
	default:
		o, n = ensure(b, 5+sz)
		prefixu32(o[n:], mstr32, uint32(sz))
		n += 5
	}
	return o[:n+copy(o[n:], str)]
}

// AppendComplex64 appends a complex64 to the slice as a MessagePack extension
func AppendComplex64(b []byte, c complex64) []byte {
	o, n := ensure(b, Complex64Size)
	o[n] = mfixext8
	o[n+1] = Complex64Extension
	big.PutUint32(o[n+2:], math.Float32bits(real(c)))
	big.PutUint32(o[n+6:], math.Float32bits(imag(c)))
	return o
}

// AppendComplex128 appends a complex128 to the slice as a MessagePack extension
func AppendComplex128(b []byte, c complex128) []byte {
	o, n := ensure(b, Complex128Size)
	o[n] = mfixext16
	o[n+1] = Complex128Extension
	big.PutUint64(o[n+2:], math.Float64bits(real(c)))
	big.PutUint64(o[n+10:], math.Float64bits(imag(c)))
	return o
}

// AppendTime appends a time.Time to the slice as a MessagePack extension
func AppendTime(b []byte, t time.Time) []byte {
	o, n := ensure(b, TimeSize)
	t = t.UTC()
	o[n] = mext8
	o[n+1] = 12
	o[n+2] = TimeExtension
	putUnix(o[n+3:], t.Unix(), int32(t.Nanosecond()))
	return o
}

// AppendMapStrStr appends a map[string]string to the slice
// as a MessagePack map with 'str'-type keys and values
func AppendMapStrStr(b []byte, m map[string]string) []byte {
	sz := uint32(len(m))
	b = AppendMapHeader(b, sz)
	for key, val := range m {
		b = AppendString(b, key)
		b = AppendString(b, val)
	}
	return b
}
