package model_serializer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSerializeListFullStruct(t *testing.T) {
	type testStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	obj := testStruct{"foo", 42}

	fields := []interface{}{"Field1", "Field2"}
	res := SerializeList([]*testStruct{&obj}, fields)
	assert.Equal(t, []map[string]interface{}{{"field1": "foo", "field2": 42}}, res)
}

func TestSerializeFullStruct(t *testing.T) {
	type testStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	obj := testStruct{"foo", 42}

	fields := []interface{}{"Field1", "Field2"}
	res := Serialize(&obj, fields)
	assert.Equal(t, map[string]interface{}{"field1": "foo", "field2": 42}, res)
}

func TestSerializeSelectedFields(t *testing.T) {
	type testStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	obj := testStruct{"foo", 42}

	fields := []interface{}{"Field1"}
	res := Serialize(&obj, fields)
	assert.Equal(t, map[string]interface{}{"field1": "foo"}, res)
}

func TestSerializeNestedStruct(t *testing.T) {
	type nested struct {
		FieldNested bool `json:"fieldNested"`
	}
	type testStruct struct {
		Field1 string  `json:"field1"`
		Field2 *nested `json:"field2"`
	}

	obj := testStruct{"foo", &nested{true}}

	fields := []interface{}{
		"Field1",
		map[string][]interface{}{"Field2": {"FieldNested"}},
	}
	res := Serialize(&obj, fields)
	assert.Equal(t, map[string]interface{}{"field1": "foo", "field2": map[string]interface{}{"fieldNested": true}}, res)
}

func TestSerializeTag(t *testing.T) {
	type testStruct struct {
		Field1 string `json:"mySuper Field"`
	}

	obj := testStruct{"foo"}
	fields := []interface{}{"Field1"}
	res := Serialize(&obj, fields)
	assert.Equal(t, map[string]interface{}{"mySuper Field": "foo"}, res)
}

func TestSerializeFieldWithoutTag(t *testing.T) {
	type testStruct struct {
		Field1 string
	}

	obj := testStruct{"foo"}
	fields := []interface{}{"Field1"}
	assert.PanicsWithValue(t, `Field "Field1" has no json tag`, func() {
		Serialize(&obj, fields)
	})
}

func TestSerializeFieldWithArray(t *testing.T) {
	type testStruct struct {
		Field1 []string `json:"field"`
	}

	obj := testStruct{[]string{"foo", "bar"}}
	fields := []interface{}{"Field1"}
	res := Serialize(&obj, fields)
	assert.Equal(t, map[string]interface{}{"field": []string{"foo", "bar"}}, res)
}

func TestSerializeFieldWithSliceOfStructs(t *testing.T) {
	type nested struct {
		FieldNested bool `json:"fieldNested"`
	}
	type testStruct struct {
		Field1 []*nested `json:"field"`
	}

	obj := testStruct{[]*nested{{true}, {false}}}
	fields := []interface{}{
		map[string][]interface{}{"Field1": {"FieldNested"}},
	}
	res := Serialize(&obj, fields)
	expected := map[string]interface{}{
		"field": []interface{}{
			map[string]interface{}{"fieldNested": true},
			map[string]interface{}{"fieldNested": false},
		},
	}
	assert.Equal(t, expected, res)
}

func TestSerializeFieldWithArrayOfStructs(t *testing.T) {
	type nested struct {
		FieldNested bool `json:"fieldNested"`
	}
	type testStruct struct {
		Field1 [2]*nested `json:"field"`
	}

	obj := testStruct{[2]*nested{{true}, {false}}}
	fields := []interface{}{
		map[string][]interface{}{"Field1": {"FieldNested"}},
	}
	res := Serialize(&obj, fields)
	expected := map[string]interface{}{
		"field": []interface{}{
			map[string]interface{}{"fieldNested": true},
			map[string]interface{}{"fieldNested": false},
		},
	}
	assert.Equal(t, expected, res)
}

func TestSerializeFieldWithUnexpectedPassedField(t *testing.T) {
	type testStruct struct {
		Field1 map[string]string `json:"field"`
	}

	obj := testStruct{}
	fields := []interface{}{map[string][]interface{}{"Field1": {"Field2"}}}
	assert.PanicsWithValue(t, "Undefined type for serializer. Need to implement it", func() {
		Serialize(&obj, fields)
	})
}

func TestFilterFields(t *testing.T) {
	type testStruct struct {
		Field1 *string `json:"field1"`
		Field2 *int    `json:"field2"`
	}

	field1V := "foo"
	field2V := 42
	obj := testStruct{&field1V, &field2V}
	fields := []string{"Field2"}
	FilterFields(&obj, fields)
	assert.Nil(t, obj.Field1)
	assert.Equal(t, 42, *obj.Field2)
}

func TestFilterMapFields(t *testing.T) {
	obj := map[string]interface{}{"field1": 1, "field2": "val", "field3": nil}
	fields := []string{"field1", "field3"}
	FilterMapFields(obj, fields)
	expected := map[string]interface{}{"field1": 1}
	assert.Equal(t, expected, obj)
}
