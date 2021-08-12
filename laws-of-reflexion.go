// Package reflexcion
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-12
package main

import (
	"fmt"
	"reflect"
)

// 1. Reflection goes from interface value to reflection object.
// TypeOf returns the reflection Type of the value in the interface{}.
// func TypeOf(i interface{}) Type
func fromInterfaceToObject() {
	defer fmt.Println()

	fmt.Println("1. Reflection goes from interface value to reflection object.")

	var x float64 = 3.4
	fmt.Println("type:", reflect.TypeOf(x))
	fmt.Println("value:", reflect.ValueOf(x).String())
	fmt.Println()

	xValue := reflect.ValueOf(x)
	fmt.Println("type:", xValue.Type())
	fmt.Println("kind is float64:", xValue.Kind() == reflect.Float64)
	fmt.Println("value:", xValue.Float())
	fmt.Println()

	var y uint8 = 'y'
	yValue := reflect.ValueOf(y)
	fmt.Println("type:", yValue.Type())                            // uint8.
	fmt.Println("kind is uint8: ", yValue.Kind() == reflect.Uint8) // true.
	y = uint8(yValue.Uint())
}

// 2. Reflection goes from reflection object to interface value.
// Interface returns v's value as an interface{}.
// func (v Value) Interface() interface{}
func fromObjectToInterface() {
	defer fmt.Println()

	fmt.Println("2. Reflection goes from reflection object to interface value.")

	var x float64 = 3.4
	xValue := reflect.ValueOf(x)
	xNew := xValue.Interface().(float64) // y will have type float64.
	fmt.Println("xNew:", xNew)
	fmt.Println("xValue interface:", xValue.Interface())
	fmt.Printf("value is %7.1e\n", xValue.Interface())
}

// 3. To modify a reflection object, the value must be settable.
// xValue := reflect.ValueOf(x)
// xValue.SetFloat(7.1) // Error: will panic.
// panic: reflect.Value.SetFloat using unaddressable value
func modifyObjectIfSettable() {
	defer fmt.Println()

	fmt.Println("3. To modify a reflection object, the value must be settable.")

	var x float64 = 3.4
	xValue := reflect.ValueOf(x)
	fmt.Println("settability of wValue:", xValue.CanSet()) // settability of wValue: false
	fmt.Println()

	xPointer := reflect.ValueOf(&x) // Note: take the address of x.
	fmt.Println("type of wPointer:", xPointer.Type())
	fmt.Println("settability of wPointer:", xPointer.CanSet())

	elem := xPointer.Elem()
	fmt.Println("settability of elem:", elem.CanSet()) // settability of elem: true
	elem.SetFloat(7.1)
	fmt.Println("elem value:", elem.Interface())
	fmt.Println("x value:", x)
	fmt.Println()

	type T struct {
		A int
		B string
	}
	tObject := T{23, "skidoo"}
	tElem := reflect.ValueOf(&tObject).Elem()
	tType := tElem.Type()
	for i := 0; i < tElem.NumField(); i++ {
		f := tElem.Field(i)
		fmt.Printf("%d: %s %s = %v\n", i, tType.Field(i).Name, f.Type(), f.Interface())
	}

	fmt.Println()
	tElem.Field(0).SetInt(77)
	tElem.Field(1).SetString("Sunset Strip")
	fmt.Println("t is now", tObject)
}

func main() {
	fromInterfaceToObject()
	fromObjectToInterface()
	modifyObjectIfSettable()
}
