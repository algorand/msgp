package msgp

import (
	"fmt"
)

// Raw represents raw formatted bytes.
// We "blindly" store it during encode and retrieve the raw bytes during decode.
type Raw []byte

// MsgIsZero returns whether this is a zero value
func (z *Raw) MsgIsZero() bool {
	return len(*z) == 0
}

// MarshalMsg marshal raw bytes
func (z *Raw) MarshalMsg(b []byte) (o []byte, err error) {
	o = append(b, (*z)...)
	return o, nil
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *Raw) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var isnil bool
	var sz int

	o = bts
	switch NextType(o) {
	case StrType:
		_, o, err = ReadStringBytes(o)
		return
	case BinType:
		_, o, err = ReadBytesZC(o)
	case MapType:
		sz, isnil, o, err = ReadMapHeaderBytes(o)
		if _, ok := err.(TypeError); ok {
			sz, isnil, o, err = ReadArrayHeaderBytes(o)
			if err != nil {
				return
			}
			for i := 0; i < sz; i++ {
				// unmarshal all the child fields.
				o, err = z.UnmarshalMsg(o)
				if err != nil {
					return
				}
			}
		} else {
			if err != nil {
				return
			}
			if isnil {
				*z = []byte{}
			}
			for key := 0; key < sz; key++ {
				// unmarshal all the child keys
				switch NextType(o) {
				case StrType:
					_, o, err = ReadMapKeyZC(o)
				case UintType:
					_, o, err = ReadUint64Bytes(o)
				case IntType:
					_, o, err = ReadInt64Bytes(o)
				case Float64Type:
					_, o, err = ReadFloat64Bytes(o)
				case Float32Type:
					_, o, err = ReadFloat32Bytes(o)
				case BoolType:
					_, o, err = ReadBoolBytes(o)
				default:
					err = fmt.Errorf("unexpected map key type %v", NextType(o))
				}
				if err != nil {
					return
				}
				o, err = z.UnmarshalMsg(o)
				if err != nil {
					return
				}
			}
		}
	case ArrayType:
		sz, isnil, o, err = ReadArrayHeaderBytes(o)
		if err != nil {
			return
		}
		for i := 0; i < sz; i++ {
			// unmarshal all the child fields.
			o, err = z.UnmarshalMsg(o)
			if err != nil {
				return
			}
		}
	case Float64Type:
		_, o, err = ReadFloat64Bytes(o)
	case Float32Type:
		_, o, err = ReadFloat32Bytes(o)
	case BoolType:
		_, o, err = ReadBoolBytes(o)
	case UintType:
		_, o, err = ReadUint64Bytes(o)
	case IntType:
		_, o, err = ReadInt64Bytes(o)
	case ExtensionType:
		err = fmt.Errorf("msgp raw UnmarshalMsg does not support extensions")
	default:
		err = fmt.Errorf("unexpected type %v", NextType(o))
	}

	if err == nil {
		*z = bts[:len(bts)-len(o)]

	}

	return
}

// Msgsize returns the number of bytes in raw
func (z *Raw) Msgsize() (s int) {
	return len(*z)
}
