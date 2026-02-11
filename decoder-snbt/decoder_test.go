package decoder_snbt

import (
	"fmt"
	encoder_go_object "neomega-nbt/encoder-go-object"
	"testing"
)

func TestDecoderNumArray(t *testing.T) {
	assertValue := func(src string, valueStr string) {
		var reciver any
		encoder := encoder_go_object.NewGoObjEncoder(&reciver, encoder_go_object.EncodeOption{
			ArrayAsSlice: true,
		})
		_, err := DecodeSnbtTo([]uint8(src), encoder, DecodeOption{})
		if err != nil {
			t.Fatalf("when decoding %v, get error: %v", src, err)
		}
		if fmt.Sprintf("%T%v", reciver, reciver) != valueStr {
			t.Fatalf("when decoding %v, expect: %v, get: %v", src, valueStr, fmt.Sprintf("%T%v", reciver, reciver))
		}
	}
	assertValue("[I;  12,  -34i,567l, 89b,]", "[]int32[12 -34 567 89]")
	assertValue("[B;12,  -34i,567l, 89b,true,false]", "[]uint8[12 222 55 89 1 0]")
	assertValue("[L;12,  -34i,567l, 89b,true,false]", "[]int64[12 -34 567 89 1 0]")
}

func TestDecoderString(t *testing.T) {
	assertValue := func(src string, valueStr string) {
		var reciver any
		encoder := encoder_go_object.NewGoObjEncoder(&reciver, encoder_go_object.EncodeOption{
			ArrayAsSlice: true,
		})
		_, err := DecodeSnbtTo([]uint8(src), encoder, DecodeOption{})
		if err != nil {
			t.Fatalf("when decoding %v, get error: %v", src, err)
		}
		if fmt.Sprintf("%T:%v", reciver, reciver) != valueStr {
			t.Fatalf("when decoding %v, expect: %v, get: %v", src, valueStr, fmt.Sprintf("%T:%v", reciver, reciver))
		}
	}
	assertValue(`"abc()  _+:'\"\\测试1"`, `string:abc()  _+:'"\测试1`)
	assertValue(`'abc()  _+:"\'\\测试2'`, `string:abc()  _+:"'\测试2`)
}

func TestDecoderStringOrNumber(t *testing.T) {
	assertValue := func(src string, valueStr string) {
		var reciver any
		encoder := encoder_go_object.NewGoObjEncoder(&reciver, encoder_go_object.EncodeOption{
			ArrayAsSlice: true,
		})
		_, err := DecodeSnbtTo([]uint8(src), encoder, DecodeOption{})
		if err != nil {
			t.Fatalf("when decoding %v, get error: %v", src, err)
		}
		if fmt.Sprintf("%T:%v", reciver, reciver) != valueStr {
			t.Fatalf("when decoding %v, expect: %v, get: %v", src, valueStr, fmt.Sprintf("%T:%v", reciver, reciver))
		}
	}
	assertValue(`abc()  _+:'\"\\测试1"`, `string:abc`)
	assertValue(`abc()  _+:"\'\\测试2'`, `string:abc`)
	assertValue(`123`, `int32:123`)
	assertValue(`123abc`, `string:123abc`)
}

func TestDecoderCompound(t *testing.T) {
	assertValue := func(src string, valueStr string) {
		var reciver any
		encoder := encoder_go_object.NewGoObjEncoder(&reciver, encoder_go_object.EncodeOption{
			ArrayAsSlice: true,
		})
		_, err := DecodeSnbtTo([]uint8(src), encoder, DecodeOption{})
		if err != nil {
			t.Fatalf("when decoding %v, get error: %v", src, err)
		}
		if fmt.Sprintf("%T:%v", reciver, reciver) != valueStr {
			t.Fatalf("when decoding %v, expect: %v, get: %v", src, valueStr, fmt.Sprintf("%T:%v", reciver, reciver))
		}
	}
	assertValue(`{}`, `map[string]interface {}:map[]`)
	assertValue(`{
		key1: 123,
		'key2': 'somevalue1',
		"key3": {
			subkey1: 12e34,
			"subkey2": "somevalue2"
		}
	}`, `map[string]interface {}:map[key1:123 key2:somevalue1 key3:map[subkey1:1.2e+35 subkey2:somevalue2]]`)
}

func TestDecoderList(t *testing.T) {
	assertValue := func(src string, valueStr string) {
		var reciver any
		encoder := encoder_go_object.NewGoObjEncoder(&reciver, encoder_go_object.EncodeOption{
			ArrayAsSlice: true,
		})
		_, err := DecodeSnbtTo([]uint8(src), encoder, DecodeOption{})
		if err != nil {
			t.Fatalf("when decoding %v, get error: %v", src, err)
		}
		if fmt.Sprintf("%T:%v", reciver, reciver) != valueStr {
			t.Fatalf("when decoding %v, expect: %v, get: %v", src, valueStr, fmt.Sprintf("%T:%v", reciver, reciver))
		}
	}
	// assertValue(`[]`, `[]interface {}:[]`)
	assertValue(`[{key1: 123},{},abc,[{key1: 123},ef]]`, `[]interface {}:[map[key1:123] map[] abc [map[key1:123] ef]]`)
}

// {IsChickenJockey: 0b, Brain: {memories: {}}, HurtByTimestamp: 0, Invulnerable: 0b, FallFlying: 0b, ForcedAge: 0, PortalCooldown: 0, AbsorptionAmount: 0.0f, FallDistance: 0.0f, InLove: 0, DeathTime: 0s, HandDropChances: [0.085f, 0.085f], PersistenceRequired: 0b, UUID: [I; -979712766, -1308996237, -1267152056, 27464117], Age: 0, Motion: [0.0d, -0.0784000015258789d, 0.0d], Health: 4.0f, LeftHanded: 0b, Air: 300s, OnGround: 1b, Rotation: [219.4912f, 0.0f], HandItems: [{}, {}], ArmorDropChances: [0.085f, 0.085f, 0.085f, 0.085f], EggLayTime: 10841, Pos: [22.169742620891626d, -60.0d, 17.89903512329325d], Fire: -1s, ArmorItems: [{}, {}, {}, {}], CanPickUpLoot: 0b, attributes: [{id: "minecraft:generic.movement_speed", base: 0.25d}, {id: "minecraft:generic.follow_range", modifiers: [{amount: 0.011031093445209002d, id: "minecraft:random_spawn_bonus", operation: "add_multiplied_base"}], base: 16.0d}], HurtTime: 0s}

func TestDecoder(t *testing.T) {
	assertValue := func(src string, valueStr string) {
		var reciver any
		encoder := encoder_go_object.NewGoObjEncoder(&reciver, encoder_go_object.EncodeOption{
			ArrayAsSlice: true,
		})
		_, err := DecodeSnbtTo([]uint8(src), encoder, DecodeOption{})
		if err != nil {
			t.Fatalf("when decoding %v, get error: %v", src, err)
		}
		if fmt.Sprintf("%T:%v", reciver, reciver) != valueStr {
			t.Fatalf("when decoding %v, expect: %v, get: %v", src, valueStr, fmt.Sprintf("%T:%v", reciver, reciver))
		}
	}
	// assertValue(`[]`, `[]interface {}:[]`)
	assertValue(
		`{IsChickenJockey: 0b, Brain: {memories: {}}, HurtByTimestamp: 0, Invulnerable: 0b, FallFlying: 0b, ForcedAge: 0, PortalCooldown: 0, AbsorptionAmount: 0.0f, FallDistance: 0.0f, InLove: 0, DeathTime: 0s, HandDropChances: [0.085f, 0.085f], PersistenceRequired: 0b, UUID: [I; -979712766, -1308996237, -1267152056, 27464117], Age: 0, Motion: [0.0d, -0.0784000015258789d, 0.0d], Health: 4.0f, LeftHanded: 0b, Air: 300s, OnGround: 1b, Rotation: [219.4912f, 0.0f], HandItems: [{}, {}], ArmorDropChances: [0.085f, 0.085f, 0.085f, 0.085f], EggLayTime: 10841, Pos: [22.169742620891626d, -60.0d, 17.89903512329325d], Fire: -1s, ArmorItems: [{}, {}, {}, {}], CanPickUpLoot: 0b, attributes: [{id: "minecraft:generic.movement_speed", base: 0.25d}, {id: "minecraft:generic.follow_range", modifiers: [{amount: 0.011031093445209002d, id: "minecraft:random_spawn_bonus", operation: "add_multiplied_base"}], base: 16.0d}], HurtTime: 0s}`,
		`map[string]interface {}:map[AbsorptionAmount:0 Age:0 Air:300 ArmorDropChances:[0.085 0.085 0.085 0.085] ArmorItems:[map[] map[] map[] map[]] Brain:map[memories:map[]] CanPickUpLoot:0 DeathTime:0 EggLayTime:10841 FallDistance:0 FallFlying:0 Fire:-1 ForcedAge:0 HandDropChances:[0.085 0.085] HandItems:[map[] map[]] Health:4 HurtByTimestamp:0 HurtTime:0 InLove:0 Invulnerable:0 IsChickenJockey:0 LeftHanded:0 Motion:[0 -0.07840000152587889 0] OnGround:1 PersistenceRequired:0 PortalCooldown:0 Pos:[22.169742620891626 -60 17.89903512329325] Rotation:[219.4912 0] UUID:[-979712766 -1308996237 -1267152056 27464117] attributes:[map[base:0.25 id:minecraft:generic.movement_speed] map[base:16 id:minecraft:generic.follow_range modifiers:[map[amount:0.011031093445209002 id:minecraft:random_spawn_bonus operation:add_multiplied_base]]]]]`,
	)
}
