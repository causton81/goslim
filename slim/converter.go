package slim

import (
	"fmt"
	"github.com/causton81/goslim/lib"
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

type intConv struct{}

func (intConv) Type() reflect.Type {
	return reflect.TypeOf(int(0))
}

func (intConv) In(s string) reflect.Value {
	n, err := strconv.Atoi(s)
	lib.Must(err)
	return reflect.ValueOf(n)
}

func (intConv) Out(value reflect.Value) string {
	return fmt.Sprintf("%d", value.Int())
}

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

type sliceIntConv struct {
}

func (sliceIntConv) Type() reflect.Type {
	return reflect.TypeOf([]int{})
}

func (sliceIntConv) In(s string) reflect.Value {
	parts := strings.Split(s, ",")
	rval := make([]int, len(parts))
	for i, e := range parts {
		var err error
		rval[i], err = strconv.Atoi(e)
		lib.Must(err)
	}

	return reflect.ValueOf(rval)
}

func (sliceIntConv) Out(value reflect.Value) string {
	numbers := value.Interface().([]int)
	strs := make([]string, len(numbers))
	for i, elem := range numbers {
		strs[i] = strconv.Itoa(elem)
	}
	return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
	//return fmt.Sprintf("%v", numbers)
}

func RegisterConverter(c Converter) {
	converters[c.Type()] = c
}

//type Scoper interface {
//	GetPrefix() string
//}

var typePrefix = ""

func SetTypePrefix(p string) {
	typePrefix = p
}

func convertArguments(funcTyp reflect.Type, args slimList) []reflect.Value {
	ret := make([]reflect.Value, funcTyp.NumIn())
	for i := 0; i < funcTyp.NumIn(); i++ {
		argTyp := funcTyp.In(i)
		log.Printf("arg typ %s", argTyp)
		conv := getConverterForType(argTyp)
		ret[i] = conv.In(args[i].(slimString).String())
	}

	return ret
}

func getConverterForType(argTyp reflect.Type) Converter {
	conv, found := converters[argTyp]
	if !found {
		panic(fmt.Errorf("message:<<NO_CONVERTER_FOR_ARGUMENT_NUMBER %s%s.>>", typePrefix, argTyp.Name()))
	}
	return conv
}
