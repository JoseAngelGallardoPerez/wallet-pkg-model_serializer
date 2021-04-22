package model_serializer

import (
	"fmt"
	"reflect"
)

// FieldSerializer is a handy callback type
// it must return field name as a first return argument and value as a second one
type FieldSerializer func(model interface{}) (fieldName string, value interface{})

// Serialize serializes any model struct by passed fields map.
// Rules to use:
// 1. String fields should equal names defined in struct
// 2. All passed string fields should have "json" tag in struct
//
// Parameters:
// model  - pointer to struct should be serialized
// fields - array of fields that should be included in result. Element can be a string or map[string][]interface{} for nested structs or array of structs
//	the list may also include FieldSerializer functions
func Serialize(model interface{}, fields []interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	val := reflect.ValueOf(model)
	if val.IsNil() {
		return result
	}

	elem := val.Elem()
	modelType := elem.Type()
	// range all passed fields
	for _, name := range fields {

		switch reflect.ValueOf(name).Kind() {
		case reflect.String:
			serializeStrField(elem, modelType, name.(string), result)
		case reflect.Func:
			if fieldSerializer, ok := name.(FieldSerializer); ok {
				fieldName, value := fieldSerializer(model)
				result[fieldName] = value
				continue
			}
			panic("Undefined func type for serializer.")
		default:
			// serialize nested structs or container of structs
			for fieldNameStr, mapFields := range name.(map[string][]interface{}) {
				fieldType, _ := modelType.FieldByName(fieldNameStr)
				serializedName := getSerializedName(fieldType)
				fieldValue := elem.FieldByName(fieldNameStr)
				switch fieldType.Type.Kind() {
				case reflect.Slice:
					serializeArrayField(fieldValue, serializedName, mapFields, result)
				case reflect.Array:
					serializeArrayField(fieldValue, serializedName, mapFields, result)
				case reflect.Ptr:
					result[serializedName] = Serialize(fieldValue.Interface(), mapFields)
				case reflect.Struct:
					result[serializedName] = Serialize(fieldValue.Addr().Interface(), mapFields)
				default:
					panic("Undefined type for serializer. Need to implement it")
				}
			}
		}
	}
	return result
}

func SerializeList(models interface{}, fields []interface{}) []map[string]interface{} {
	slice := reflect.ValueOf(models)
	res := make([]map[string]interface{}, slice.Len())

	for i := 0; i < slice.Len(); i++ {
		res[i] = Serialize(slice.Index(i).Interface(), fields)
	}

	return res
}

// FilterFields sets nil fot struct field if field is not in fields array.
// Does not work for nested maps
func FilterFields(model interface{}, fields []string) {
	modelValue := reflect.ValueOf(model).Elem()
	modelType := modelValue.Type()
	fieldsCount := modelValue.NumField()

	for i := 0; i < fieldsCount; i++ {
		field := modelValue.Field(i)
		if !field.IsNil() && !containsField(modelType.Field(i).Name, fields) {
			field.Set(reflect.Zero(field.Type()))
		}
	}
}

// FilterMapFields removes fields not in array and nils.
// Does not work for nested maps.
// Can be used for updating only specified fields in model
func FilterMapFields(mapData map[string]interface{}, fields []string) {
	for k, v := range mapData {
		if !containsField(k, fields) || isNilInterface(v) {
			delete(mapData, k)
		}
	}
}

func getSerializedName(structField reflect.StructField) string {
	if val, ok := structField.Tag.Lookup("json"); !ok {
		panic(fmt.Sprintf(`Field "%s" has no json tag`, structField.Name))
	} else {
		return val
	}
}

func serializeArrayField(fieldValue reflect.Value, fieldName string,
	fields []interface{}, targetMap map[string]interface{},
) {
	serializedContainer := make([]interface{}, fieldValue.Len())
	for i := 0; i < fieldValue.Len(); i++ {
		elemInterface := fieldValue.Index(i).Interface()
		serializedContainer[i] = Serialize(elemInterface, fields)
	}
	targetMap[fieldName] = serializedContainer
}

func serializeStrField(structValue reflect.Value, structType reflect.Type,
	fieldName string, targetMap map[string]interface{},
) {
	fieldValue := structValue.FieldByName(fieldName)
	fieldType, _ := structType.FieldByName(fieldName)
	serializedName := getSerializedName(fieldType)
	targetMap[serializedName] = fieldValue.Interface()
}

func isNilInterface(v interface{}) bool {
	if v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil()) {
		return true
	}
	return false
}

func containsField(field string, fields []string) bool {
	for _, v := range fields {
		if field == v {
			return true
		}
	}
	return false
}
