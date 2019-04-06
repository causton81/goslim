package slim

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/causton81/goslim/lib"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

var slimIn *os.File
var slimOut *os.File

func redirectStdoutAndStderr() {
	realStderr := os.Stderr
	//slimOut = os.Stdout

	stdoutReadSide, stdoutWriteSide, err := os.Pipe()
	lib.Must(err)

	stderrReadSide, stderrWriteSide, err := os.Pipe()
	lib.Must(err)

	psOut := newPrefixedStream(realStderr, "SOUT")
	psErr := newPrefixedStream(realStderr, "SERR")

	go func() {
		buf := make([]byte, 4096)

		for {
			fds := []unix.PollFd{
				{Fd: int32(stdoutReadSide.Fd()), Events: unix.POLLIN},
				{Fd: int32(stderrReadSide.Fd()), Events: unix.POLLIN},
			}
			n, err := unix.Poll(fds, 100)
			if err != nil {
				panic(err)
			} else if 0 < n {
				if 0 < unix.POLLIN&fds[0].Revents {
					numRead, err := stdoutReadSide.Read(buf)
					lib.Must(err)
					data := buf[0:numRead]
					psOut.write(data)
					//final := bytes.ReplaceAll(data, newline, stdoutStart)
					//realStderr.Write(final)
				}
				if 0 < unix.POLLIN&fds[1].Revents {
					numRead, err := stderrReadSide.Read(buf)
					lib.Must(err)
					data := buf[0:numRead]
					psErr.write(data)
					//final := bytes.ReplaceAll(data, newline, stdoutStart)
					//realStderr.Write(final)
				}
			}
		}
	}()

	os.Stdout = stdoutWriteSide
	os.Stderr = stderrWriteSide
}

func RegisterFixture(fix interface{}) {
	//log.Println(reflect.TypeOf(fix).String())
	RegisterFixtureWithName(fix, reflect.TypeOf(fix).String())
}

var fixtureTypes = make(map[string]reflect.Type)

func RegisterFixtureWithName(fix interface{}, scriptAlias string) {
	fixtureTypes[scriptAlias] = reflect.TypeOf(fix)
}

var instances = make(map[string]reflect.Value)

func ListenAndServe() {
	slimIn = os.Stdin
	slimOut = os.Stdout
	redirectStdoutAndStderr()

	slimOut.WriteString("Slim -- V0.5\n")

	running := true

	for running {
		slimmer := loadSlim(slimIn)

		if s, ok := slimmer.(slimString); ok {
			if "bye" == s {
				running = false
				break
			}

			log.Fatalf("unexpected string from fitnessse: '%s'", s)
		}
		instructions := slimmer.(slimList)
		slimResults := make(slimList, len(instructions))
		log.Println("Slim Instructions:")
		for idx, inst := range instructions {
			log.Println(inst)
			inst := inst.(slimList)
			id := inst[0].String()
			op := inst[1].String()

			switch op {
			case "make":
				instanceName := inst[2].String()
				className := inst[3].String()
				typ, found := fixtureTypes[className]
				if !found {
					slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<NO_CLASS %s>>", className))
				} else {
					instances[instanceName] = reflect.New(typ)
					slimResults[idx] = asList(id, "OK")
				}

			// [decisionTable_0_1 call decisionTable_0 table [[numerator denominator quotient?] [10 2 5.0] [12.6 3 4.2] [22 7 ~=3.14] [9 3 <5] [11 2 4<_<6] [100 4 33]]]
			case "call":
				returnString := "/__VOID__/"
				instanceName := inst[2].String()
				methodName := strings.Title(inst[3].String())
				instance, found := instances[instanceName]
				numFields := len(inst)
				var args slimList
				hasArguments := 4 < numFields
				if hasArguments {
					args = inst[4:numFields]
				}
				if !found {
					slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<NO_INSTANCE %s>>", instanceName))
				} else {
					m := instance.MethodByName(methodName)
					if !m.IsValid() {
						slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<NO_METHOD_IN_CLASS %s %s>>", methodName, instance.Type()))
					} else if m.Type().NumIn() != len(args) {
						slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<%s expects exactly %d arguments, but received %d>>", methodName, m.Type().NumIn(), len(args)))
					} else {
						//slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:TODO"))
						convertedValues := convertArguments(m.Type(), args)
						res := m.Call(convertedValues)
						switch len(res) {
						case 0:
							// empty
						case 1:
							c := converters[res[0].Type()]
							returnString = c.Out(res[0])
						default:
							returnString = "__EXCEPTION__:multi-value return is not supported"
						}

						slimResults[idx] = asList(id, returnString)
					}
				}
			default:
				slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<MALFORMED_INSTRUCTION %s>>", op))
			}
		}

		fmt.Fprintf(slimOut, "%s", slimResults.Slim())
	}
}

//type converter func(s slimString) reflect.Value
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
	//return strconv.FormatFloat(value.Float(), 'f', 1, 64)
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

func init() {
	RegisterConverter(stringConv{})
	RegisterConverter(float64conv{})
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

type slimmer interface {
	Slim() string
	String() string
}

type slimString string

func (s slimString) String() string {
	return string(s)
}

func (s slimString) Slim() string {
	return fmt.Sprintf("%06d:%s", len(s), s)

}

type slimList []slimmer

func (l slimList) String() string {
	return fmt.Sprint([]slimmer(l))
}

func asList(items ...interface{}) slimList {
	l := slimList{}
	for _, e := range items {
		switch e := e.(type) {
		case string:
			l = append(l, slimString(e))
		}
	}

	return l
}

func (l slimList) Slim() string {
	numElem := len(l)
	sb := new(strings.Builder)
	fmt.Fprintf(sb, "[%06d:", numElem)
	for _, e := range l {
		fmt.Fprintf(sb, "%s:", e.Slim())
	}
	fmt.Fprint(sb, "]")
	return slimString(sb.String()).Slim()
}

func loadSlim(r io.Reader) slimmer {
	buf := bufio.NewReader(r)
	length := parseLength(buf)

	if length < 1 {
		return slimString("")
	} else {
		peekBytes, err := buf.Peek(1)
		lib.Must(err)

		isList := '[' == peekBytes[0]
		if isList {
			//return slimString("TODO")
			buf.Discard(1)
			numElems := parseLength(buf)
			l := make(slimList, numElems)
			for i := range l {
				l[i] = loadSlim(buf)
				nextByte, err := buf.ReadByte()
				lib.Must(err)

				if ':' != nextByte {
					log.Fatalf("expected next byte to be :")
				}
			}

			nextByte, err := buf.ReadByte()
			lib.Must(err)
			if ']' != nextByte {
				log.Fatalf("expected next byte to be ]")
			}

			return l
		} else {
			bld := new(strings.Builder)
			bld.Grow(length)
			_, err := io.CopyN(bld, buf, int64(length))
			lib.Must(err)

			return slimString(bld.String())
		}
	}
}

func parseLength(buf *bufio.Reader) int {
	sizeField, err := buf.ReadString(':')
	lib.Must(err)
	size, err := strconv.ParseInt(sizeField[0:len(sizeField)-1], 10, 0)
	lib.Must(err)
	return int(size)
}

type prefixedStream struct {
	outStream io.Writer
	replacer  []byte
	lastByte  byte
}

func newPrefixedStream(w io.Writer, prefix string) *prefixedStream {
	return &prefixedStream{
		outStream: w,
		replacer:  []byte(fmt.Sprintf("\n%s :", prefix)),
		lastByte:  '\n',
	}
}

func (ps *prefixedStream) write(in []byte) {
	trailingByte := byte(0)
	if 0 < len(in) {
		trailingByte = in[len(in)-1]
	}
	if '\n' == trailingByte {
		in = in[0 : len(in)-1]
	}
	newline := []byte("\n")
	replacer := []byte("\nSOUT :")
	sep := ' '
	if '\n' != ps.lastByte {
		sep = '.'
	}
	fmt.Fprintf(ps.outStream, "%s%c:", replacer[1:5], sep)
	ps.outStream.Write(bytes.ReplaceAll(in, newline, replacer))
	ps.outStream.Write(newline)

	ps.lastByte = trailingByte
}
