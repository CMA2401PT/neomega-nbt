package decoder_bedrock_block_props

import (
	"errors"
	nbt "neomega-nbt"
)

var errNoLeftBracket = errors.New("missing left [")
var errNoRightBracket = errors.New("missing right ]")
var errNoBeingingKey = errors.New("missing \" at key begining")
var errNoEndKey = errors.New("missing \" at key end")
var errNoIdentifier = errors.New("missing = between key and value")
var errMissingValue = errors.New("missing value")
var errUnknowValueType = errors.New("value can only be true/false/positive int/string")

// ["byteKey"=true/false,"intKey"=0/1/2/3...,"stringKey"="string"]
func Decode[Seq ~[]uint8 | ~string](props Seq, enc nbt.EncoderHighLevel) (err error) {
	p := 0
	consumeWhiteSpace := func() (eof bool) {
		for p < len(props) {
			if props[p] != ' ' {
				break
			}
			p += 1
		}
		return p == len(props)
	}
	matchSpecific := func(b byte) bool {
		if p == len(props) {
			return false
		}
		ok := props[p] == b
		if ok {
			p += 1
			return true
		} else {
			return false
		}

	}
	if consumeWhiteSpace() {
		return nil
	}
	if !matchSpecific('[') {
		return errNoLeftBracket
	}
	err = enc.WriteRootTag("", nbt.ComplexCompound)
	if err != nil {
		return err
	}
	err = enc.BeginComplexCompound()
	if err != nil {
		return err
	}
	isFirst := true
	var keyStartP, keyEndP int
	writeTag := func(t nbt.HighLevelValueType) error {
		if !isFirst {
			err = enc.NextCompoundElement()
			if err != nil {
				return err
			}
		} else {
			isFirst = false
		}
		err = enc.WriteCompoundTag(string(props[keyStartP:keyEndP]), t)
		if err != nil {
			return err
		}
		return nil
	}
	for {
		// check eof
		if consumeWhiteSpace() {
			return errNoRightBracket
		}
		// ending
		if matchSpecific(']') {
			return enc.EndComplexCompound()

		}
		// read key
		if !matchSpecific('"') {
			return errNoBeingingKey
		}
		keyStartP = p
		for !matchSpecific('"') {
			p += 1
			if p == len(props) {
				return errNoEndKey
			}
		}
		keyEndP = p - 1
		if consumeWhiteSpace() || !matchSpecific('=') {
			return errNoIdentifier
		}
		if consumeWhiteSpace() {
			return errMissingValue
		}
		switch b := props[p]; b {
		case '"':
			p += 1
			valueStartP := p
			for !matchSpecific('"') {
				p += 1
				if p == len(props) {
					return errNoEndKey
				}
			}
			ValueEndP := p - 1
			err = writeTag(nbt.TrivialString)
			if err != nil {
				return err
			}
			err = enc.WriteString(string(props[valueStartP:ValueEndP]), nbt.ValuePosCompoundOrRoot)
			if err != nil {
				return err
			}
		case 't':
			p += 1
			if !matchSpecific('r') {
				return errUnknowValueType
			}
			if !matchSpecific('u') {
				return errUnknowValueType
			}
			if !matchSpecific('e') {
				return errUnknowValueType
			}
			err = writeTag(nbt.TrivialByte)
			if err != nil {
				return err
			}
			err = enc.WriteInt8(1, nbt.ValuePosCompoundOrRoot)
			if err != nil {
				return err
			}
		case 'f':
			p += 1
			if !matchSpecific('a') {
				return errUnknowValueType
			}
			if !matchSpecific('l') {
				return errUnknowValueType
			}
			if !matchSpecific('s') {
				return errUnknowValueType
			}
			if !matchSpecific('e') {
				return errUnknowValueType
			}
			err = writeTag(nbt.TrivialByte)
			if err != nil {
				return err
			}
			err = enc.WriteInt8(0, nbt.ValuePosCompoundOrRoot)
			if err != nil {
				return err
			}
		default:
			if b >= '0' && b <= '9' {
				v := int32(0)
				for b >= '0' && b <= '9' {
					p++
					v = v*10 + int32(b-'0')
					b = props[p]
				}
				err = writeTag(nbt.TrivialInt)
				if err != nil {
					return err
				}
				err = enc.WriteInt32(v, nbt.ValuePosCompoundOrRoot)
				if err != nil {
					return err
				}
			} else {
				return errUnknowValueType
			}
		}
		if consumeWhiteSpace() {
			return errNoRightBracket
		}
		matchSpecific(',')
	}

}
