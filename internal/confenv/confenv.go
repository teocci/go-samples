// Package confenv
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-12
package confenv

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	CharHyphen     = "-"
	CharUnderscore = "_"
	CharComa       = ","
	CharEqual      = "="
	BoolYes        = "yes"
	BoolNo         = "no"
	BoolTrue       = "true"
	BoolFalse      = "false"
	TagYAML        = "yaml"
)

func load(environment map[string]string, prefix string, reflectValue reflect.Value) error {
	reflectType := reflectValue.Type()

	if reflectType == reflect.TypeOf(time.Duration(0)) {
		if ev, ok := environment[prefix]; ok {
			d, err := time.ParseDuration(ev)
			if err != nil {
				return fmt.Errorf("%s: %s", prefix, err)
			}
			reflectValue.Set(reflect.ValueOf(d))
		}
		return nil
	}

	switch reflectType.Kind() {
	case reflect.String:
		if ev, ok := environment[prefix]; ok {
			reflectValue.SetString(ev)
		}
		return nil

	case reflect.Int:
		if ev, ok := environment[prefix]; ok {
			iv, err := strconv.ParseInt(ev, 10, 64)
			if err != nil {
				return fmt.Errorf("%s: %s", prefix, err)
			}
			reflectValue.SetInt(iv)
		}
		return nil

	case reflect.Uint64:
		if ev, ok := environment[prefix]; ok {
			iv, err := strconv.ParseUint(ev, 10, 64)
			if err != nil {
				return fmt.Errorf("%s: %s", prefix, err)
			}
			reflectValue.SetUint(iv)
		}
		return nil

	case reflect.Bool:
		if ev, ok := environment[prefix]; ok {
			switch strings.ToLower(ev) {
			case BoolYes, BoolTrue:
				reflectValue.SetBool(true)

			case BoolNo, BoolFalse:
				reflectValue.SetBool(false)

			default:
				return fmt.Errorf("%s: invalid value '%s'", prefix, ev)
			}
		}
		return nil

	case reflect.Slice:
		if reflectType.Elem().Kind() == reflect.String {
			if ev, ok := environment[prefix]; ok {
				newValue := reflect.Zero(reflectType)
				for _, sv := range strings.Split(ev, CharComa) {
					newValue = reflect.Append(newValue, reflect.ValueOf(sv))
				}
				reflectValue.Set(newValue)
			}
			return nil
		}

	case reflect.Map:
		for k := range environment {
			mapPrefix := prefix + CharUnderscore
			if !strings.HasPrefix(k, mapPrefix) {
				continue
			}

			mapKey := strings.Split(k[len(mapPrefix):], CharUnderscore)[0]
			if len(mapKey) == 0 {
				continue
			}

			// allow only keys in uppercase
			if mapKey != strings.ToUpper(mapKey) {
				continue
			}

			// initialize only if there's at least one key
			if reflectValue.IsNil() {
				reflectValue.Set(reflect.MakeMap(reflectType))
			}

			mapKeyLower := strings.ToLower(mapKey)
			nv := reflectValue.MapIndex(reflect.ValueOf(mapKeyLower))
			zero := reflect.Value{}
			if nv == zero {
				nv = reflect.New(reflectType.Elem().Elem())
				reflectValue.SetMapIndex(reflect.ValueOf(mapKeyLower), nv)
			}

			err := load(environment, mapPrefix+mapKey, nv.Elem())
			if err != nil {
				return err
			}
		}
		return nil

	case reflect.Struct:
		numField := reflectType.NumField()
		for index := 0; index < numField; index++ {
			field := reflectType.Field(index)

			// load only public fields
			if field.Tag.Get(TagYAML) == CharHyphen {
				continue
			}

			newPrefix := prefix + CharUnderscore + strings.ToUpper(field.Name)
			err := load(environment, newPrefix, reflectValue.Field(index))
			if err != nil {
				return err
			}
		}
		return nil
	}

	return fmt.Errorf("unsupported type: %v", reflectType)
}

// Load fills a structure with data from the environment.
func Load(prefix string, value interface{}) error {
	environment := make(map[string]string)
	for _, envString := range os.Environ() {
		envKV := strings.SplitN(envString, CharEqual, 2)
		environment[envKV[0]] = envKV[1]
	}

	return load(environment, prefix, reflect.ValueOf(value).Elem())
}
