package helpers

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/elastic/elastic-agent-libs/mapstr"
	"github.com/elastic/elastic-agent-shipper-client/pkg/proto/messages"
	"github.com/stretchr/testify/require"
)

var result *messages.Value

func BenchmarkCustomUnmarshal(b *testing.B) {
	testStructType := struct {
		A int
		B string
	}{
		A: 5,
		B: "test",
	}
	testList := mapstr.M{
		"field1": mapstr.M{
			"value":          3,
			"value-str":      "test",
			"value-list":     []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			"value-list-str": []string{"value1", "value2", "value3", "valu4", "value5", "value6"},
			"value-struct":   testStructType,
			"value-bytes":    []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			"value-time":     time.Now(),
			"value-map": map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
				"key4": "value4",
			},
		},
	}
	var r *messages.Value
	var err error
	// benchmark performance for the map, which will usually be the most complex type to marshal
	// there's a handful of different possilbe ways to handle the NewValue conversion, so it's helpful
	// to have a benchmark in case we decide to adjust this in the future
	for i := 0; i < b.N; i++ {
		r, err = NewValue(testList)
		if err != nil {
			b.Logf("error: %s", err)
			b.FailNow()
		}
		result = r
	}
}

func TestStructValue(t *testing.T) {
	testStructType := struct {
		A int
		B string
	}{
		A: 5,
		B: "test",
	}
	ts := time.Now()
	cases := []struct {
		name string
		in   interface{}
		exp  *messages.Value
	}{
		{
			name: "string conversion",
			in:   "test-string",
			exp:  &messages.Value{Kind: &messages.Value_StringValue{StringValue: "test-string"}},
		},
		{
			name: "int conversion",
			in:   32,
			exp:  &messages.Value{Kind: &messages.Value_NumberValue{NumberValue: float64(32)}},
		},
		{
			name: "nil value",
			in:   nil,
			exp:  &messages.Value{Kind: &messages.Value_NullValue{NullValue: messages.NullValue_NULL_VALUE}},
		},
		{
			name: "test map conversion",
			in: mapstr.M{
				"field1": mapstr.M{
					"value":     3,
					"value-str": "test",
				},
			},
			exp: NewStructValue(&messages.Struct{Data: map[string]*messages.Value{
				"field1": NewStructValue(&messages.Struct{Data: map[string]*messages.Value{
					"value":     NewNumberValue(3),
					"value-str": NewStringValue("test"),
				}}),
			}}),
		},
		{
			name: "test struct conversion",
			in:   testStructType,
			exp: NewStructValue(&messages.Struct{Data: map[string]*messages.Value{
				"A": NewNumberValue(5),
				"B": NewStringValue("test"),
			}}),
		},
		{
			name: "list conversion of string type",
			in:   []string{"value1", "value2"},
			exp: NewListValue(&messages.ListValue{Values: []*messages.Value{
				NewStringValue("value1"),
				NewStringValue("value2"),
			}}),
		},
		{
			name: "list conversion of interface type",
			in:   []interface{}{"value1", 3},
			exp: NewListValue(&messages.ListValue{Values: []*messages.Value{
				NewStringValue("value1"),
				NewNumberValue(3),
			}}),
		},
		{
			name: "proper handling of byte arrays",
			in:   []byte{0xFF, 0xFF},
			exp:  NewStringValue(base64.StdEncoding.EncodeToString([]byte{0xFF, 0xFF})),
		},
		{
			name: "proper handling of timestamps",
			in:   ts,
			exp:  NewTimestampValue(ts),
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := NewValue(c.in)
			require.NoError(t, err)
			require.Equal(t, c.exp, res)
		})
	}
}
