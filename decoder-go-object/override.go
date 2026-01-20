package decoder_go_object

import (
	nbt "neomega-nbt"
)

type GetDecoderFunc func(data any) (DecoderOverride, error)

type DecoderOverride interface {
	GetNBType() nbt.HighLevelValueType
	DecodeTo(valuePos nbt.ValuePos, encoder nbt.EncoderHighLevel, getChildDecoder GetDecoderFunc) error
}
