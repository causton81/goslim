package slim

import (
	"fmt"
	"github.com/causton81/goslim/lib"
	"log"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// I tried implementing this with user-defined basic types (`type stringConv string`) instead of empty structs, but I
// could not figure out how to get the underlying type when reflecting (reflection returns the user-defined type which
// I don't want to force users to use). The current implementation with a Type() method is the best way I know to find
// the right converter.
// TODO consider splitting into two interfaces because returned slices will never call Out()
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
	//numbers := value.Interface().([]int)
	//strs := make([]string, len(numbers))
	//for i, elem := range numbers {
	//	strs[i] = strconv.Itoa(elem)
	//}
	//return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
	////return fmt.Sprintf("%v", numbers)
	panic("Out should not be called on returned slices")
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

var symbolExpr = regexp.MustCompile(`\$([[:alpha:]]+)`)

func convertArguments(funcTyp reflect.Type, args slimList) []reflect.Value {
	ret := make([]reflect.Value, funcTyp.NumIn())
	for i := 0; i < funcTyp.NumIn(); i++ {
		argString := args[i].(slimString).String()
		argTyp := funcTyp.In(i)
		log.Printf("arg typ %s: %s", argTyp, argString)
		conv := getConverterForType(argTyp)
		symMatch := symbolExpr.FindAllStringSubmatch(argString, -1)
		log.Printf("sym match: '%s'", symMatch)
		anyMatches := 0 < len(symMatch)
		if anyMatches {
			wholeArgIsSymbol := symMatch[0][0] == argString
			if wholeArgIsSymbol {
				symVal := symbols[symMatch[0][1]]
				if reflect.TypeOf(symVal).AssignableTo(argTyp) {
					ret[i] = reflect.ValueOf(symVal)
					continue
				}
			} else {
				for _, m := range symMatch {
					symVal := symbols[m[1]]
					log.Printf("sym %s = '%s'", m[0], symVal)

					symString := fmt.Sprintf("%s", symVal)
					argString = strings.ReplaceAll(argString, m[0], symString)
				}
			}
		}
		ret[i] = conv.In(argString)
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
