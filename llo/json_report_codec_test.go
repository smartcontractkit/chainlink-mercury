package llo

import (
	"bytes"
	"fmt"
	"math"
	"math/rand"
	reflect "reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2/types"

	"github.com/smartcontractkit/chainlink-common/pkg/types/llo"
	llotypes "github.com/smartcontractkit/chainlink-common/pkg/types/llo"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Fuzz_JSONCodec_Decode_Unpack(f *testing.F) {
	validJson := []byte(`{"foo":"bar"}`)
	emptyInput := []byte(``)
	nilInput := []byte(nil)
	nullJson := []byte(`null`)
	incompleteJson := []byte(`{`)
	notJson := []byte(`"random string"`)
	unprintable := []byte{1, 2, 3}
	validJsonReport := []byte(`{"ConfigDigest":"0102030000000000000000000000000000000000000000000000000000000000","SeqNr":43,"ChannelID":46,"ValidAfterSeconds":44,"ObservationTimestampSeconds":45,"Values":[{"Type":0,"Value":"1"},{"Type":0,"Value":"2"},{"Type":1,"Value":"Q{Bid: 3.13, Benchmark: 4.4, Ask: 5.12}"}],"Specimen":true}`)
	invalidConfigDigest := []byte(`{"SeqNr":42,"ConfigDigest":"foo"}`)
	invalidConfigDigestNotEnoughBytes := []byte(`{"SeqNr":42,"ConfigDigest":"0xdead"}`)
	badStreamValues := []byte(`{"SeqNr":42,"ConfigDigest":"0102030000000000000000000000000000000000000000000000000000000000", "Values":[{"Type":0,"Value":null},{"Type":-1,"Value":"2"}]}`)

	f.Add(validJson)
	f.Add(emptyInput)
	f.Add(nilInput)
	f.Add(nullJson)
	f.Add(incompleteJson)
	f.Add(notJson)
	f.Add(unprintable)
	f.Add(validJsonReport)
	f.Add(invalidConfigDigest)
	f.Add(invalidConfigDigestNotEnoughBytes)
	f.Add(badStreamValues)

	validPackedJsonTemplate := `{"configDigest":"0102030000000000000000000000000000000000000000000000000000000000","seqNr":43,"report":%s,"sigs":[{"Signature":"AgME","Signer":2}]}`
	packedJsonReports := [][]byte{
		[]byte(fmt.Sprintf(validPackedJsonTemplate, validJson)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, emptyInput)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, nilInput)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, nullJson)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, incompleteJson)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, notJson)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, unprintable)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, validJsonReport)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, invalidConfigDigest)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, invalidConfigDigestNotEnoughBytes)),
		[]byte(fmt.Sprintf(validPackedJsonTemplate, badStreamValues)),
	}
	for _, packedJsonReport := range packedJsonReports {
		f.Add(packedJsonReport)
	}

	packedJsonSigTemplate := `{"configDigest":"0102030000000000000000000000000000000000000000000000000000000000","seqNr":43,"report":{},"sigs":[{"Signature":%s,"Signer":2}]}`
	badSigs := [][]byte{
		[]byte(fmt.Sprintf(packedJsonSigTemplate, `null`)),
		[]byte(fmt.Sprintf(packedJsonSigTemplate, `""`)),
		[]byte(fmt.Sprintf(packedJsonSigTemplate, `1`)),
		[]byte(fmt.Sprintf(packedJsonSigTemplate, `[]`)),
		[]byte(fmt.Sprintf(packedJsonSigTemplate, `"abc$def#ghi!"`)),
	}
	for _, badSig := range badSigs {
		f.Add(badSig)
	}

	// Set up the random number generator
	rand.Seed(time.Now().UnixNano())

	var codec JSONReportCodec
	f.Fuzz(func(t *testing.T, data []byte) {
		// test that it doesn't panic, don't care about errors
		codec.Decode(data)
		codec.Unpack(data)
		codec.UnpackDecode(data)
	})
}

func Test_JSONCodec_Properties(t *testing.T) {
	properties := gopter.NewProperties(nil)

	ctx := tests.Context(t)
	cd := llotypes.ChannelDefinition{}
	codec := JSONReportCodec{}

	properties.Property("Encode/Decode", prop.ForAll(
		func(r Report) bool {
			b, err := codec.Encode(ctx, r, cd)
			require.NoError(t, err)
			r2, err := codec.Decode(b)
			require.NoError(t, err)
			return equalReports(r, r2)

		},
		gen.StrictStruct(reflect.TypeOf(&Report{}), map[string]gopter.Gen{
			"ConfigDigest":                genConfigDigest(),
			"SeqNr":                       genSeqNr(),
			"ChannelID":                   gen.UInt32(),
			"ValidAfterSeconds":           gen.UInt32(),
			"ObservationTimestampSeconds": gen.UInt32(),
			"Values":                      genStreamValues(),
			"Specimen":                    gen.Bool(),
		}),
	))

	properties.Property("Pack/Unpack", prop.ForAll(
		func(digest types.ConfigDigest, seqNr uint64, report ocr2types.Report, sigs []types.AttributedOnchainSignature) bool {
			b, err := codec.Pack(digest, seqNr, report, sigs)
			require.NoError(t, err)
			digest2, seqNr2, report2, sigs2, err := codec.Unpack(b)
			require.NoError(t, err)

			if digest != digest2 {
				return false
			}
			if seqNr != seqNr2 {
				return false
			}
			if !bytes.Equal(report, report2) {
				return false
			}
			if len(sigs) != len(sigs2) {
				return false
			}
			for i := range sigs {
				if sigs[i].Signer != sigs2[i].Signer || !bytes.Equal(sigs[i].Signature, sigs2[i].Signature) {
					return false
				}
			}
			return true
		},
		genConfigDigest(),
		genSeqNr(),
		genSerializedReport(),
		genSigs(),
	))

	properties.TestingRun(t)
}

func equalReports(r, r2 Report) bool {
	if r.ConfigDigest != r2.ConfigDigest {
		return false
	}
	if r.SeqNr != r2.SeqNr {
		return false
	}
	if r.ChannelID != r2.ChannelID {
		return false
	}
	if r.ValidAfterSeconds != r2.ValidAfterSeconds {
		return false
	}
	if r.ObservationTimestampSeconds != r2.ObservationTimestampSeconds {
		return false
	}
	if len(r.Values) != len(r2.Values) {
		return false
	}
	for i := range r.Values {
		if !equalStreamValues(r.Values[i], r2.Values[i]) {
			return false
		}
	}
	if r.Specimen != r2.Specimen {
		return false
	}
	return true
}

func equalStreamValues(sv, sv2 StreamValue) bool {
	if sv.Type() != sv2.Type() {
		return false
	}
	m1, err := sv.MarshalBinary()
	if err != nil {
		panic(err)
	}
	m2, err := sv2.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return bytes.Equal(m1, m2)
}

func genConfigDigest() gopter.Gen {
	return func(p *gopter.GenParameters) *gopter.GenResult {
		var cd types.ConfigDigest
		p.Rng.Read(cd[:])
		return gopter.NewGenResult(cd, gopter.NoShrinker)
	}
}

func genSeqNr() gopter.Gen {
	return gen.UInt64Range(1, math.MaxUint64)
}

func genSerializedReport() gopter.Gen {
	return gen.Const(ocr2types.Report(`{"foo":"bar"}`))
}

func genSigs() gopter.Gen {
	return gen.SliceOf(genSig())
}

func genSig() gopter.Gen {
	return gen.StrictStruct(reflect.TypeOf(&types.AttributedOnchainSignature{}), map[string]gopter.Gen{
		"Signature": genSigBytes(),
		"Signer":    genSigner(),
	})
}

func genSigner() gopter.Gen {
	return gen.UInt8().Map(func(v uint8) commontypes.OracleID {
		return commontypes.OracleID(v)
	})
}

func genSigBytes() gopter.Gen {
	return gen.SliceOf(gen.UInt8())
}

func genDecimalValue() gopter.Gen {
	return func(p *gopter.GenParameters) *gopter.GenResult {
		var sv StreamValue
		sv = ToDecimal(decimal.NewFromFloat(p.Rng.Float64()))
		return gopter.NewGenResult(sv, gopter.NoShrinker)
	}
}

func genQuote() gopter.Gen {
	return func(p *gopter.GenParameters) *gopter.GenResult {
		var sv StreamValue
		sv = &Quote{
			Bid:       decimal.NewFromFloat(p.Rng.Float64()),
			Benchmark: decimal.NewFromFloat(p.Rng.Float64()),
			Ask:       decimal.NewFromFloat(p.Rng.Float64()),
		}
		return gopter.NewGenResult(sv, gopter.NoShrinker)
	}
}

func genStreamValue() gopter.Gen {
	return func(p *gopter.GenParameters) *gopter.GenResult {
		switch p.Rng.Intn(3) {
		case 0:
			return genDecimalValue()(p)
		case 1:
			return genQuote()(p)
		case 2:
			return gopter.NewGenResult((StreamValue)(nil), gopter.NoShrinker)
		}
		return nil
	}
}

var streamValueSliceType = reflect.TypeOf((*StreamValue)(nil)).Elem()

func genStreamValues() gopter.Gen {
	return gen.SliceOf(genStreamValue(), streamValueSliceType)
}

func Test_JSONCodec(t *testing.T) {
	t.Run("Encode=>Decode", func(t *testing.T) {
		ctx := tests.Context(t)
		r := Report{
			ConfigDigest:                types.ConfigDigest([32]byte{1, 2, 3}),
			SeqNr:                       43,
			ChannelID:                   llotypes.ChannelID(46),
			ValidAfterSeconds:           44,
			ObservationTimestampSeconds: 45,
			Values:                      []StreamValue{ToDecimal(decimal.NewFromInt(1)), ToDecimal(decimal.NewFromInt(2)), &Quote{Bid: decimal.NewFromFloat(3.13), Benchmark: decimal.NewFromFloat(4.4), Ask: decimal.NewFromFloat(5.12)}},
			Specimen:                    true,
		}

		cdc := JSONReportCodec{}

		encoded, err := cdc.Encode(ctx, r, llo.ChannelDefinition{})
		require.NoError(t, err)

		assert.Equal(t, `{"ConfigDigest":"0102030000000000000000000000000000000000000000000000000000000000","SeqNr":43,"ChannelID":46,"ValidAfterSeconds":44,"ObservationTimestampSeconds":45,"Values":[{"Type":0,"Value":"1"},{"Type":0,"Value":"2"},{"Type":1,"Value":"Q{Bid: 3.13, Benchmark: 4.4, Ask: 5.12}"}],"Specimen":true}`, string(encoded))

		decoded, err := cdc.Decode(encoded)
		require.NoError(t, err)

		assert.Equal(t, r, decoded)
	})
	t.Run("Pack=>Unpack", func(t *testing.T) {
		t.Run("report is not valid JSON", func(t *testing.T) {
			digest := types.ConfigDigest([32]byte{1, 2, 3})
			seqNr := uint64(43)
			report := ocr2types.Report(`foobar`)
			sigs := []types.AttributedOnchainSignature{{Signature: []byte{2, 3, 4}, Signer: 2}}

			cdc := JSONReportCodec{}

			_, err := cdc.Pack(digest, seqNr, report, sigs)
			require.EqualError(t, err, "json: error calling MarshalJSON for type json.RawMessage: invalid character 'o' in literal false (expecting 'a')")
		})
		t.Run("report is valid JSON", func(t *testing.T) {
			digest := types.ConfigDigest([32]byte{1, 2, 3})
			seqNr := uint64(43)
			report := ocr2types.Report(`{"foo":"bar"}`)
			sigs := []types.AttributedOnchainSignature{{Signature: []byte{2, 3, 4}, Signer: 2}}

			cdc := JSONReportCodec{}

			packed, err := cdc.Pack(digest, seqNr, report, sigs)
			require.NoError(t, err)
			assert.Equal(t, `{"configDigest":"0102030000000000000000000000000000000000000000000000000000000000","seqNr":43,"report":{"foo":"bar"},"sigs":[{"Signature":"AgME","Signer":2}]}`, string(packed))

			digest2, seqNr2, report2, sigs2, err := cdc.Unpack(packed)
			require.NoError(t, err)
			assert.Equal(t, digest, digest2)
			assert.Equal(t, seqNr, seqNr2)
			assert.Equal(t, report, report2)
			assert.Equal(t, sigs, sigs2)
		})
	})
	t.Run("UnpackDecode unpacks and decodes report", func(t *testing.T) {
		b := []byte(`{"configDigest":"0102030000000000000000000000000000000000000000000000000000000000","seqNr":43,"report":{"ConfigDigest":"0102030000000000000000000000000000000000000000000000000000000000","SeqNr":43,"ChannelID":46,"ValidAfterSeconds":44,"ObservationTimestampSeconds":45,"Values":[{"Type":0,"Value":"1"},{"Type":0,"Value":"2"},{"Type":1,"Value":"Q{Bid: 3.13, Benchmark: 4.4, Ask: 5.12}"}],"Specimen":true},"sigs":[{"Signature":"AgME","Signer":2}]}`)

		cdc := JSONReportCodec{}
		digest, seqNr, report, sigs, err := cdc.UnpackDecode(b)
		require.NoError(t, err)

		assert.Equal(t, types.ConfigDigest([32]byte{1, 2, 3}), digest)
		assert.Equal(t, uint64(43), seqNr)
		assert.Equal(t, Report{
			ConfigDigest:                types.ConfigDigest([32]byte{1, 2, 3}),
			SeqNr:                       43,
			ChannelID:                   llotypes.ChannelID(46),
			ValidAfterSeconds:           44,
			ObservationTimestampSeconds: 45,
			Values:                      []StreamValue{ToDecimal(decimal.NewFromInt(1)), ToDecimal(decimal.NewFromInt(2)), &Quote{Bid: decimal.NewFromFloat(3.13), Benchmark: decimal.NewFromFloat(4.4), Ask: decimal.NewFromFloat(5.12)}},
			Specimen:                    true,
		}, report)
		assert.Equal(t, []types.AttributedOnchainSignature{{Signature: []byte{2, 3, 4}, Signer: 2}}, sigs)
	})
	t.Run("invalid input fails decode", func(t *testing.T) {
		cdc := JSONReportCodec{}
		_, err := cdc.Decode([]byte(`{}`))
		assert.EqualError(t, err, "missing SeqNr")
		_, err = cdc.Decode([]byte(`{"seqNr":1}`))
		assert.EqualError(t, err, "invalid ConfigDigest; cannot convert bytes to ConfigDigest. bytes have wrong length 0")
	})
}
