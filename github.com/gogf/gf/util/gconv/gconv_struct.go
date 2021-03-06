// Copyright 2017-2019 gf Author(https://github.com/gogf/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

package gconv

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gogf/gf/internal/structs"
	"github.com/gogf/gf/internal/utilstr"
)

// Struct maps the params key-value pairs to the corresponding struct object's properties.
// The third parameter <mapping> is unnecessary, indicating the mapping rules between the custom key name
// and the attribute name(case sensitive).
//
// Note:
// 1. The <params> can be any type of map/struct, usually a map.
// 2. The second parameter <pointer> should be a pointer to the struct object.
// 3. Only the public attributes of struct object can be mapped.
// 4. If <params> is a map, the key of the map <params> can be lowercase.
//    It will automatically convert the first letter of the key to uppercase
//    in mapping procedure to do the matching.
//    It ignores the map key, if it does not match.
func Struct(params interface{}, pointer interface{}, mapping ...map[string]string) error {
	if params == nil {
		return errors.New("params cannot be nil")
	}
	if pointer == nil {
		return errors.New("object pointer cannot be nil")
	}
	paramsMap := Map(params)
	if paramsMap == nil {
		return fmt.Errorf("invalid params: %v", params)
	}
	// Using reflect to do the converting,
	// it also supports type of reflect.Value for <pointer>(always in internal usage).
	elem, ok := pointer.(reflect.Value)
	if !ok {
		rv := reflect.ValueOf(pointer)
		if kind := rv.Kind(); kind != reflect.Ptr {
			return fmt.Errorf("object pointer should be type of: %v", kind)
		}
		// Using IsNil on reflect.Ptr variable is OK.
		if !rv.IsValid() || rv.IsNil() {
			return errors.New("object pointer cannot be nil")
		}
		elem = rv.Elem()
	}
	// Auto create struct object.
	// For example, if <pointer> is **User, then <elem> is *User, which is a pointer to User.
	if elem.Type().Kind() == reflect.Ptr && (!elem.IsValid() || elem.IsNil()) {
		e := reflect.New(elem.Type().Elem()).Elem()
		elem.Set(e.Addr())
		elem = e
	}
	// It only performs one converting to the same attribute.
	// doneMap is used to check repeated converting.
	doneMap := make(map[string]bool)
	// It first checks the passed mapping rules.
	if len(mapping) > 0 && len(mapping[0]) > 0 {
		for mapK, mapV := range mapping[0] {
			if v, ok := paramsMap[mapK]; ok {
				doneMap[mapV] = true
				if err := bindVarToStructAttr(elem, mapV, v); err != nil {
					return err
				}
			}
		}
	}
	// It secondly checks the tags of attributes.
	tagMap := structs.TagMapName(pointer, structTagPriority, true)
	for tagK, tagV := range tagMap {
		if _, ok := doneMap[tagV]; ok {
			continue
		}
		if v, ok := paramsMap[tagK]; ok {
			doneMap[tagV] = true
			if err := bindVarToStructAttr(elem, tagV, v); err != nil {
				return err
			}
		}
	}
	// It finally do the converting with default rules.
	attrMap := make(map[string]struct{})
	elemType := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		// Only do converting to public attributes.
		if !utilstr.IsLetterUpper(elemType.Field(i).Name[0]) {
			continue
		}
		attrMap[elemType.Field(i).Name] = struct{}{}
	}
	for mapK, mapV := range paramsMap {
		name := ""
		for _, checkName := range []string{
			utilstr.UcFirst(mapK),
			utilstr.ReplaceByMap(mapK, map[string]string{
				"_": "",
				"-": "",
				" ": "",
			})} {
			if _, ok := doneMap[checkName]; ok {
				continue
			}
			if _, ok := tagMap[checkName]; ok {
				continue
			}
			// Loop to find the matched attribute name.
			for value, _ := range attrMap {
				if strings.EqualFold(checkName, value) {
					name = value
					break
				}
				if strings.EqualFold(checkName, strings.Replace(value, "_", "", -1)) {
					name = value
					break
				}
			}
			doneMap[checkName] = true
			if name != "" {
				break
			}
		}
		// No matching, give up this attribute converting.
		if name == "" {
			continue
		}
		if err := bindVarToStructAttr(elem, name, mapV); err != nil {
			return err
		}
	}
	return nil
}

// StructDeep do Struct function recursively.
// See Struct.
func StructDeep(params interface{}, pointer interface{}, mapping ...map[string]string) error {
	if err := Struct(params, pointer, mapping...); err != nil {
		return err
	} else {
		rv, ok := pointer.(reflect.Value)
		if !ok {
			rv = reflect.ValueOf(pointer)
		}
		kind := rv.Kind()
		for kind == reflect.Ptr {
			rv = rv.Elem()
			kind = rv.Kind()
		}
		switch kind {
		case reflect.Struct:
			rt := rv.Type()
			for i := 0; i < rv.NumField(); i++ {
				// Only do converting to public attributes.
				if !utilstr.IsLetterUpper(rt.Field(i).Name[0]) {
					continue
				}
				trv := rv.Field(i)
				switch trv.Kind() {
				case reflect.Struct:
					if err := StructDeep(params, trv, mapping...); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// ???????????????????????????????????????????????????
func bindVarToStructAttr(elem reflect.Value, name string, value interface{}) (err error) {
	structFieldValue := elem.FieldByName(name)
	// ????????????????????????????????????map????????????struct????????????????????????????????????????????????return
	if !structFieldValue.IsValid() {
		return nil
	}
	// CanSet??????????????????????????????(???????????????)
	if !structFieldValue.CanSet() {
		return nil
	}
	// ?????????value?????????struct??????????????????????????????????????????gconv???
	defer func() {
		// ?????????????????????????????????????????????????????????(?????????????????????????????????)???????????????????????????
		if recover() != nil {
			err = bindVarToReflectValue(structFieldValue, value)
		}
	}()
	structFieldValue.Set(reflect.ValueOf(Convert(value, structFieldValue.Type().String())))
	return nil
}

// ?????????????????????????????????????????????????????????
func bindVarToStructByIndex(elem reflect.Value, index int, value interface{}) (err error) {
	structFieldValue := elem.FieldByIndex([]int{index})
	// ?????????????????????????????????
	if !structFieldValue.IsValid() {
		return nil
	}
	// CanSet??????????????????????????????(???????????????)
	if !structFieldValue.CanSet() {
		return nil
	}
	// ?????????value?????????struct??????????????????????????????????????????gconv???
	defer func() {
		// ?????????????????????????????????????????????????????????(?????????????????????????????????)???????????????????????????
		if recover() != nil {
			err = bindVarToReflectValue(structFieldValue, value)
		}
	}()
	structFieldValue.Set(reflect.ValueOf(Convert(value, structFieldValue.Type().String())))
	return nil
}

// ????????????????????????????????????????????????recover?????????????????????????????????(??????????????????)
func bindVarToReflectValue(structFieldValue reflect.Value, value interface{}) error {
	switch structFieldValue.Kind() {
	// ??????????????????
	case reflect.Struct:
		if err := Struct(value, structFieldValue); err != nil {
			structFieldValue.Set(reflect.ValueOf(value))
		}

	// ?????????????????????
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		a := reflect.Value{}
		v := reflect.ValueOf(value)
		if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
			if v.Len() > 0 {
				a = reflect.MakeSlice(structFieldValue.Type(), v.Len(), v.Len())
				t := a.Index(0).Type()
				for i := 0; i < v.Len(); i++ {
					if t.Kind() == reflect.Ptr {
						e := reflect.New(t.Elem()).Elem()
						if err := Struct(v.Index(i).Interface(), e); err != nil {
							e.Set(reflect.ValueOf(v.Index(i).Interface()))
						}
						a.Index(i).Set(e.Addr())
					} else {
						e := reflect.New(t).Elem()
						if err := Struct(v.Index(i).Interface(), e); err != nil {
							e.Set(reflect.ValueOf(v.Index(i).Interface()))
						}
						a.Index(i).Set(e)
					}
				}
			}
		} else {
			a = reflect.MakeSlice(structFieldValue.Type(), 1, 1)
			t := a.Index(0).Type()
			if t.Kind() == reflect.Ptr {
				e := reflect.New(t.Elem()).Elem()
				if err := Struct(value, e); err != nil {
					e.Set(reflect.ValueOf(value))
				}
				a.Index(0).Set(e.Addr())
			} else {
				e := reflect.New(t).Elem()
				if err := Struct(value, e); err != nil {
					e.Set(reflect.ValueOf(value))
				}
				a.Index(0).Set(e)
			}
		}
		structFieldValue.Set(a)

	// ?????????????????????
	case reflect.Ptr:
		e := reflect.New(structFieldValue.Type().Elem()).Elem()
		if err := Struct(value, e); err != nil {
			e.Set(reflect.ValueOf(value))
		}
		structFieldValue.Set(e.Addr())

	case reflect.Interface:
		if value == nil {
			structFieldValue.Set(reflect.ValueOf((*interface{})(nil)))
		} else {
			structFieldValue.Set(reflect.ValueOf(value))
		}

	default:
		return errors.New(
			fmt.Sprintf(`cannot convert to type "%s"`,
				structFieldValue.Type().String(),
			),
		)
	}
	return nil
}
