// Package main
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-12
package main

import (
	"fmt"
	"github.com/teocci/go-samples/internal/logger"
	"reflect"
	"time"
)
type Identiflier int

const (
	_ Identiflier = iota // a = 0
	a        // iota is incremented  to 1
	b        // b = 2
)

func parseLogLevel(level string) (logger.Level, error) {
	var parsedLevel = logger.Info
	switch level {
	case logger.WarnString:
		return logger.Warn, nil
	case logger.InfoString:
		return logger.Info, nil
	case logger.DebugString:
		return logger.Debug, nil
	default:
		return parsedLevel, fmt.Errorf("unsupported log level: %s", level)
	}
}

func main() {

	i := 2
	fmt.Print("Write ", i, " as ")
	switch i {
	case 1:
		fmt.Println("one")
	case 2:
		fmt.Println("two")
	case 3:
		fmt.Println("three")
	}

	switch time.Now().Weekday() {
	case time.Saturday, time.Sunday:
		fmt.Println("It's the weekend")
	default:
		fmt.Println("It's a weekday")
	}

	t := time.Now()
	switch {
	case t.Hour() < 12:
		fmt.Println("It's before noon")
	default:
		fmt.Println("It's after noon")
	}

	whatAmI := func(i interface{}) {
		switch t := i.(type) {
		case bool:
			fmt.Println("I'm a bool")
		case int:
			fmt.Println("I'm an int")
		default:
			fmt.Printf("Don't know type %T\n", t)
		}
	}
	whatAmI(true)
	whatAmI(1)
	whatAmI("hey")
	whatAmI(int64(4))

	level, _ := parseLogLevel("level")

	fmt.Printf("Don't know type %d\n", reflect.ValueOf(level))

	var logLevel string
	if logLevel == "" {
		var foo logger.Level
		logLevel = logger.GetLevelName(foo)
	}
	fmt.Printf("Default log level: %s\n", logLevel)

	fmt.Printf("a: %d, b: %d\n", a, b)
	var foo Identiflier
	fmt.Printf("Identiflier: %d\n", reflect.ValueOf(foo))
}
