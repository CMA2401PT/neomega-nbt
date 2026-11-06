package v1_test

import (
	"bytes"
	"errors"
	"io"
	nbt "neomega-nbt"
	goencode "neomega-nbt/encoder-go-object"
	preset "neomega-nbt/presets/v1"
	"testing"
)

// collectingArray 故意用两字节缓冲，验证接收器可以分批消费而不接管父容器。
type collectingArray struct {
	data  []byte
	fail  error
	calls int
}

func (a *collectingArray) GetReciver(nbt.HighLevelValueType) (goencode.ValueReciver, error) {
	return goencode.ValueReciver{}, nil
}
func (a *collectingArray) ConsumeByteArray(length int, read func([]byte) error) (bool, error) {
	a.data = nil
	a.calls++
	if a.fail != nil {
		return true, a.fail
	}
	var buffer [2]byte
	for length > 0 {
		batch := buffer[:min(length, 2)]
		if err := read(batch); err != nil {
			return true, err
		}
		a.data = append(a.data, batch...)
		length -= len(batch)
	}
	return true, nil
}
func TestStreamingByteArray(t *testing.T) {
	for _, encoding := range []preset.Encoding{preset.EncodingBigEndian, preset.EncodingLittleEndian, preset.EncodingNetworkLittleEndian} {
		t.Run(string(rune('0'+encoding)), func(t *testing.T) {
			for _, values := range [][]byte{{0, 127, 128, 255, 5}, {}} {
				data, err := preset.EncodeToNBTBytes(struct {
					Array any
					After int32
				}{preset.Array(values), 42}, preset.OptEncodeToNBT{Encoding: encoding})
				if err != nil {
					t.Fatal(err)
				}
				target := struct {
					Array collectingArray
					After int32
				}{}
				if err = preset.DecodeFromNBT(bytes.NewReader(data), &target, preset.OptDecodeFromNBT{Encoding: encoding}); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(target.Array.data, values) || target.Array.calls != 1 || target.After != 42 {
					t.Fatalf("%+v", target)
				}
				// 普通数组接收器走原有路径，数据语义一致。
				plain := struct {
					Array []byte
					After int32
				}{}
				if err = preset.DecodeFromNBTBytes(data, &plain, preset.OptDecodeFromNBT{Encoding: encoding}); err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(plain.Array, values) || plain.After != 42 {
					t.Fatal(plain)
				}
			}
		})
	}
}
func TestStreamingArrayInsideListAndErrors(t *testing.T) {
	data, err := preset.EncodeToNBTBytes([]any{preset.Array([]byte{1, 2, 3}), preset.Array([]byte{4})}, preset.OptEncodeToNBT{Encoding: preset.EncodingBigEndian})
	if err != nil {
		t.Fatal(err)
	}
	var target []collectingArray
	if err = preset.DecodeFromNBTBytes(data, &target, preset.OptDecodeFromNBT{Encoding: preset.EncodingBigEndian}); err != nil {
		t.Fatal(err)
	}
	if len(target) != 2 || !bytes.Equal(target[0].data, []byte{1, 2, 3}) || !bytes.Equal(target[1].data, []byte{4}) {
		t.Fatal(target)
	}
	root, err := preset.EncodeToNBTBytes(preset.Array([]byte{1, 2, 3}), preset.OptEncodeToNBT{Encoding: preset.EncodingBigEndian})
	if err != nil {
		t.Fatal(err)
	}
	var truncated collectingArray
	if err = preset.DecodeFromNBTBytes(root[:len(root)-1], &truncated, preset.OptDecodeFromNBT{Encoding: preset.EncodingBigEndian}); err == nil {
		t.Fatal("truncated array accepted")
	}
	expected := errors.New("receiver stopped")
	failed := collectingArray{fail: expected}
	if err = preset.DecodeFromNBT(io.LimitReader(bytes.NewReader(root), int64(len(root))), &failed, preset.OptDecodeFromNBT{Encoding: preset.EncodingBigEndian}); !errors.Is(err, expected) {
		t.Fatal(err)
	}
}
