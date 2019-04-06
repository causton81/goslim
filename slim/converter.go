package slim

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
)

type Converter interface {
	Type() reflect.Type
	In(s string) reflect.Value
	Out(value reflect.Value) string
}

var converters = make(map[reflect.Type]Converter)

type float64conv struct{}

func (float64conv) Type() reflect.Type {
	return reflect.TypeOf(float64(0))
}

func (float64conv) Out(value reflect.Value) string {
	rval := fmt.Sprintf("%g", value.Float())
	if !strings.Contains(rval, ".") {
		rval += ".0"
	}

	return rval
}

func (float64conv) In(s string) reflect.Value {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}

	return reflect.ValueOf(v)
}

func RegisterConverter(c Converter) {
	converters[c.Type()] = c
}

type stringConv struct{}

func (stringConv) Type() reflect.Type {
	return reflect.TypeOf(string(""))
}

func (stringConv) In(s string) reflect.Value {
	return reflect.ValueOf(s)
}

func (stringConv) Out(value reflect.Value) string {
	return value.Interface().(string)
}

func convertArguments(funcTyp reflect.Type, args slimList) []reflect.Value {
	ret := make([]reflect.Value, funcTyp.NumIn())
	for i := 0; i < funcTyp.NumIn(); i++ {
		argTyp := funcTyp.In(i)
		log.Printf("arg typ %s", argTyp)
		conv, found := converters[argTyp]
		if !found {
			panic("no converter")
		}
		ret[i] = conv.In(args[i].(slimString).String())
	}

	return ret
}
