package jsonry

import (
	"encoding/json"
	"errors"
	"reflect"
)

func Marshal(source interface{}) ([]byte, error) {
	sourceValue, err := relectOnAndCheckStruct(source)
	if err != nil {
		return []byte{}, err
	}

	return json.Marshal(marshal(make(map[string]interface{}), sourceValue))
}

func relectOnAndCheckStruct(store interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(store)
	if !v.IsValid() {
		return reflect.Value{}, errors.New("the source object must be a valid struct")
	}

	if kind := actualKind(v.Type()); kind != reflect.Struct {
		return reflect.Value{}, errors.New("the source object must be a struct")
	}

	return v, nil
}

func marshal(tree map[string]interface{}, sourceValue reflect.Value) map[string]interface{} {
	if sourceValue.Kind() == reflect.Ptr {
		sourceValue = sourceValue.Elem()
	}

	sourceType := sourceValue.Type()

	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		path := computePath(field)
		navigateAndSet(path, tree, sourceValue.Field(i))
	}

	return tree
}

func navigateAndSet(path jsonryPath, tree map[string]interface{}, value reflect.Value) {
	branch := path.elements[0]

	if path.omitempty && value.IsZero() {
		return
	}

	if actualKind(value.Type()) == reflect.Struct {
		node := make(map[string]interface{})
		tree[branch.name] = node
		marshal(node, value)
		return
	}

	if path.len() == 1 {
		tree[branch.name] = value.Interface()
		return
	}

	if branch.list {
		tree[branch.name] = iterateAndSet(path.chop(), tree[branch.name], value)
		return
	}

	node, ok := tree[branch.name].(map[string]interface{})
	if !ok {
		node = make(map[string]interface{})
		tree[branch.name] = node
	}

	navigateAndSet(path.chop(), node, value)
}

func iterateAndSet(path jsonryPath, current interface{}, value reflect.Value) []interface{} {
	var result []interface{}
	orig, _ := current.([]interface{})

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Slice {
		value = reflect.ValueOf([]interface{}{value.Interface()})
	}

	for i := 0; i < value.Len(); i++ {
		var node map[string]interface{}

		if i < len(orig) {
			node, _ = orig[i].(map[string]interface{})
		}

		if node == nil {
			node = make(map[string]interface{})
		}

		navigateAndSet(path, node, value.Index(i))
		result = append(result, node)
	}

	if len(orig) > value.Len() {
		result = append(result, orig[value.Len():]...)
	}

	return result
}
