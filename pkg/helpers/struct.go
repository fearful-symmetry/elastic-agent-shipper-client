// Protocol Buffers - Google's data interchange format
// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

// Package structpb contains generated types for google/protobuf/struct.proto.
//
// The messages (i.e., Value, Struct, and ListValue) defined in struct.proto are
// used to represent arbitrary JSON. The Value message represents a JSON value,
// the Struct message represents a JSON object, and the ListValue message
// represents a JSON array. See https://json.org for more information.
//
// The Value, Struct, and ListValue types have generated MarshalJSON and
// UnmarshalJSON methods such that they serialize JSON equivalent to what the
// messages themselves represent. Use of these types with the
// "google.golang.org/protobuf/encoding/protojson" package
// ensures that they will be serialized as their JSON equivalent.
//
//
// Conversion to and from a Go interface
//
// The standard Go "encoding/json" package has functionality to serialize
// arbitrary types to a large degree. The Value.AsInterface, Struct.AsMap, and
// ListValue.AsSlice methods can convert the protobuf message representation into
// a form represented by interface{}, map[string]interface{}, and []interface{}.
// This form can be used with other packages that operate on such data structures
// and also directly with the standard json package.
//
// In order to convert the interface{}, map[string]interface{}, and []interface{}
// forms back as Value, Struct, and ListValue messages, use the NewStruct,
// NewList, and NewValue constructor functions.
//
//
// Example usage
//
// Consider the following example JSON object:
//
//	{
//		"firstName": "John",
//		"lastName": "Smith",
//		"isAlive": true,
//		"age": 27,
//		"address": {
//			"streetAddress": "21 2nd Street",
//			"city": "New York",
//			"state": "NY",
//			"postalCode": "10021-3100"
//		},
//		"phoneNumbers": [
//			{
//				"type": "home",
//				"number": "212 555-1234"
//			},
//			{
//				"type": "office",
//				"number": "646 555-4567"
//			}
//		],
//		"children": [],
//		"spouse": null
//	}
//
// To construct a Value message representing the above JSON object:
//
//	m, err := structpb.NewValue(map[string]interface{}{
//		"firstName": "John",
//		"lastName":  "Smith",
//		"isAlive":   true,
//		"age":       27,
//		"address": map[string]interface{}{
//			"streetAddress": "21 2nd Street",
//			"city":          "New York",
//			"state":         "NY",
//			"postalCode":    "10021-3100",
//		},
//		"phoneNumbers": []interface{}{
//			map[string]interface{}{
//				"type":   "home",
//				"number": "212 555-1234",
//			},
//			map[string]interface{}{
//				"type":   "office",
//				"number": "646 555-4567",
//			},
//		},
//		"children": []interface{}{},
//		"spouse":   nil,
//	})
//	if err != nil {
//		... // handle error
//	}
//	... // make use of m as a *structpb.Value
//
package helpers

import (
	base64 "encoding/base64"
	"math"
	"time"
	utf8 "unicode/utf8"

	"github.com/elastic/elastic-agent-shipper-client/pkg/proto/messages"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// NewStruct constructs a Struct from a general-purpose Go map.
// The map keys must be valid UTF-8.
// The map values are converted using NewValue.
func NewStruct(v map[string]interface{}) (*messages.Struct, error) {
	x := &messages.Struct{Data: make(map[string]*messages.Value, len(v))}
	for k, v := range v {
		if !utf8.ValidString(k) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", k)
		}
		var err error
		x.Data[k], err = NewValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}

// AsMap converts x to a general-purpose Go map.
// The map values are converted by calling Value.AsInterface.
func AsMap(x *messages.Struct) map[string]interface{} {
	vs := make(map[string]interface{})
	for k, v := range x.GetData() {
		vs[k] = AsInterface(v)
	}
	return vs
}

// AsInterface converts x to a general-purpose Go interface.
//
// Calling Value.MarshalJSON and "encoding/json".Marshal on this output produce
// semantically equivalent JSON (assuming no errors occur).
//
// Floating-point values (i.e., "NaN", "Infinity", and "-Infinity") are
// converted as strings to remain compatible with MarshalJSON.
func AsInterface(x *messages.Value) interface{} {
	switch v := x.GetKind().(type) {
	case *messages.Value_NumberValue:
		if v != nil {
			switch {
			case math.IsNaN(v.NumberValue):
				return "NaN"
			case math.IsInf(v.NumberValue, +1):
				return "Infinity"
			case math.IsInf(v.NumberValue, -1):
				return "-Infinity"
			default:
				return v.NumberValue
			}
		}
	case *messages.Value_StringValue:
		if v != nil {
			return v.StringValue
		}
	case *messages.Value_TimestampValue:
		if v != nil {
			return v.TimestampValue.AsTime()
		}
	case *messages.Value_BoolValue:
		if v != nil {
			return v.BoolValue
		}
	case *messages.Value_StructValue:
		if v != nil {
			return AsMap(v.StructValue)
		}
	case *messages.Value_ListValue:
		if v != nil {
			return AsSlice(v.ListValue)
		}
	}
	return nil
}

// AsSlice converts x to a general-purpose Go slice.
// The slice elements are converted by calling Value.AsInterface.
func AsSlice(x *messages.ListValue) []interface{} {
	vs := make([]interface{}, len(x.GetValues()))
	for i, v := range x.GetValues() {
		vs[i] = AsInterface(v)
	}
	return vs
}

// NewValue constructs a Value from a general-purpose Go interface.
//
//	╔════════════════════════╤════════════════════════════════════════════╗
//	║ Go type                │ Conversion                                 ║
//	╠════════════════════════╪════════════════════════════════════════════╣
//	║ nil                    │ stored as NullValue                        ║
//	║ bool                   │ stored as BoolValue                        ║
//	║ int, int32, int64      │ stored as NumberValue                      ║
//	║ uint, uint32, uint64   │ stored as NumberValue                      ║
//	║ float32, float64       │ stored as NumberValue                      ║
//	║ string                 │ stored as StringValue; must be valid UTF-8 ║
//	║ time.Time              │ stored as TimestampValue;                  ║
//	║ []byte                 │ stored as StringValue; base64-encoded      ║
//	║ map[string]interface{} │ stored as StructValue                      ║
//	║ []interface{}          │ stored as ListValue                        ║
//	╚════════════════════════╧════════════════════════════════════════════╝
//
// When converting an int64 or uint64 to a NumberValue, numeric precision loss
// is possible since they are stored as a float64.
func NewValue(v interface{}) (*messages.Value, error) {
	switch v := v.(type) {
	case nil:
		return NewNullValue(), nil
	case bool:
		return NewBoolValue(v), nil
	case int:
		return NewNumberValue(float64(v)), nil
	case int32:
		return NewNumberValue(float64(v)), nil
	case int64:
		return NewNumberValue(float64(v)), nil
	case uint:
		return NewNumberValue(float64(v)), nil
	case uint32:
		return NewNumberValue(float64(v)), nil
	case uint64:
		return NewNumberValue(float64(v)), nil
	case float32:
		return NewNumberValue(float64(v)), nil
	case float64:
		return NewNumberValue(float64(v)), nil
	case string:
		if !utf8.ValidString(v) {
			return nil, protoimpl.X.NewError("invalid UTF-8 in string: %q", v)
		}
		return NewStringValue(v), nil
	case time.Time:
		return NewTimestampValue(v), nil
	case []byte:
		s := base64.StdEncoding.EncodeToString(v)
		return NewStringValue(s), nil
	case map[string]interface{}:
		v2, err := NewStruct(v)
		if err != nil {
			return nil, err
		}
		return NewStructValue(v2), nil
	case []interface{}:
		v2, err := NewList(v)
		if err != nil {
			return nil, err
		}
		return NewListValue(v2), nil
	default:
		return nil, protoimpl.X.NewError("invalid type: %T", v)
	}
}

// NewNullValue constructs a new null Value.
func NewNullValue() *messages.Value {
	return &messages.Value{Kind: &messages.Value_NullValue{NullValue: messages.NullValue_NULL_VALUE}}
}

// NewBoolValue constructs a new boolean Value.
func NewBoolValue(v bool) *messages.Value {
	return &messages.Value{Kind: &messages.Value_BoolValue{BoolValue: v}}
}

// NewNumberValue constructs a new number Value.
func NewNumberValue(v float64) *messages.Value {
	return &messages.Value{Kind: &messages.Value_NumberValue{NumberValue: v}}
}

// NewStringValue constructs a new string Value.
func NewStringValue(v string) *messages.Value {
	return &messages.Value{Kind: &messages.Value_StringValue{StringValue: v}}
}

// NewTimestampValue constructs a new Timestamp Value.
func NewTimestampValue(v time.Time) *messages.Value {
	return &messages.Value{Kind: &messages.Value_TimestampValue{TimestampValue: timestamppb.New(v)}}
}

// NewStructValue constructs a new struct Value.
func NewStructValue(v *messages.Struct) *messages.Value {
	return &messages.Value{Kind: &messages.Value_StructValue{StructValue: v}}
}

// NewListValue constructs a new list Value.
func NewListValue(v *messages.ListValue) *messages.Value {
	return &messages.Value{Kind: &messages.Value_ListValue{ListValue: v}}
}

// NewList constructs a ListValue from a general-purpose Go slice.
// The slice elements are converted using NewValue.
func NewList(v []interface{}) (*messages.ListValue, error) {
	x := &messages.ListValue{Values: make([]*messages.Value, len(v))}
	for i, v := range v {
		var err error
		x.Values[i], err = NewValue(v)
		if err != nil {
			return nil, err
		}
	}
	return x, nil
}
