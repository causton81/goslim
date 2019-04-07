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
runLoop:
	for running {
		slimmer := loadSlim(slimIn)

		switch s := slimmer.(type) {
		case slimString:
			if "bye" == s {
				running = false
				break runLoop
			}
			log.Fatalf("unexpected string from fitnessse: '%s'", s)

		case slimList:
			slimResults := make(slimList, len(s))
			log.Println("Slim Instructions:")
			for idx, inst := range s {
				if processInstruction(inst.(slimList), slimResults, idx) {
					slimResults = slimResults[0 : idx+1]
					break
				}
			}
			log.Println("Slim Response:")
			for _, item := range slimResults {
				log.Println(item)
			}
			fmt.Fprintf(os.Stderr, "%s", slimResults.Slim())
			fmt.Fprintf(slimOut, "%s", slimResults.Slim())
		}
	}
}

var instanceTypes = make(map[string]string)

var symbols = make(map[string]interface{})

func processInstruction(inst slimList, slimResults slimList, idx int) (stop bool) {
	log.Println(inst)
	id := inst[0].String()
	defer func() {
		err := recover()
		if nil != err {
			typ := reflect.TypeOf(err)
			if strings.Contains(typ.Name(), "StopTest") {
				message := ""
				if str, ok := err.(fmt.Stringer); ok {
					message = str.String()
				}

				if 0 < len(message) {
					slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:ABORT_SLIM_TEST:message:<<%s>>", message))
				} else {
					slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:ABORT_SLIM_TEST:%s", err))
				}
				stop = true
			} else {
				message := ""
				switch s := err.(type) {
				case string:
					message = s
				case fmt.Stringer:
					message = s.String()
				}

				if 0 < len(message) {
					slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<%s>>", message))
				} else {
					slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:%s", err))
				}
			}
		}
	}()

	op := inst[1].String()
	switch op {
	case "make":
		instanceName := inst[2].String()
		className := inst[3].String()
		var args slimList
		if 4 < len(inst) {
			args = inst[4:]
		}
		numArgs := len(args)

		typ, found := fixtureTypes[className]
		if !found {
			panic(fmt.Errorf("message:<<COULD_NOT_INVOKE_CONSTRUCTOR %s[%d]>>", className, numArgs))
		} else {
			initMethod, hasInit := reflect.PtrTo(typ).MethodByName("Init")
			log.Printf("has init %v, args %v", hasInit, args)
			if hasInit {
				numArgsWithReceiver := 1 + numArgs
				if numArgsWithReceiver != initMethod.Type.NumIn() {
					panic(fmt.Errorf("message:<<COULD_NOT_INVOKE_CONSTRUCTOR %s[%d]>>", className, numArgs))
				}
			}
			instances[instanceName] = reflect.New(typ)
			//TODO: maybe make this better
			instanceTypes[instanceName] = className
			slimResults[idx] = asList(id, "OK")
		}

	// [decisionTable_0_1 call decisionTable_0 table [[numerator denominator quotient?] [10 2 5.0] [12.6 3 4.2] [22 7 ~=3.14] [9 3 <5] [11 2 4<_<6] [100 4 33]]]
	case "call", "callAndAssign":
		var rval slimmer = slimString("/__VOID__/")
		off := 0
		symbolName := ""
		if "callAndAssign" == op {
			symbolName = inst[2].String()
			off = 1
		}
		instanceName := inst[2+off].String()
		slimMethodName := inst[3+off].String()
		goMethodName := strings.Title(slimMethodName)
		instance, found := instances[instanceName]
		numFields := len(inst)
		var args slimList
		hasArguments := (4 + off) < numFields
		if hasArguments {
			args = inst[4+off : numFields]
		}
		if !found {
			slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<NO_INSTANCE %s>>", instanceName))
		} else {
			m := instance.MethodByName(goMethodName)
			slimArgCount := len(args)
			if !m.IsValid() {
				slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<NO_METHOD_IN_CLASS %s[%d] %s.>>", slimMethodName, slimArgCount, instanceTypes[instanceName]))
			} else if m.Type().NumIn() != slimArgCount {
				slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<%s expects exactly %d arguments, but received %d>>", goMethodName, m.Type().NumIn(), slimArgCount))
			} else {
				convertedValues := convertArguments(m.Type(), args)
				res := m.Call(convertedValues)
				switch len(res) {
				case 0:
					// empty
				case 1:
					v := res[0]
					if "callAndAssign" == op {
						symbols[symbolName] = v.Interface()
					}
					if reflect.Ptr == v.Kind() && v.IsNil() {
						rval = slimString("null")
					} else if reflect.Slice == v.Kind() {
						elemType := v.Type().Elem()
						conv := getConverterForType(elemType)
						list := make(slimList, v.Len())
						for i := 0; i < v.Len(); i++ {
							list[i] = slimString(conv.Out(v.Index(i)))
						}

						rval = list
					} else {
						conv := getConverterForType(v.Type())
						rval = slimString(conv.Out(v))
					}
				default:
					rval = slimString("__EXCEPTION__:multi-value return is not supported")
				}

				slimResults[idx] = asList(id, rval)
			}
		}
	//case "callAndAssign":

	default:
		slimResults[idx] = asList(id, fmt.Sprintf("__EXCEPTION__:message:<<MALFORMED_INSTRUCTION %s>>", op))
	}

	return
}

func init() {
	RegisterConverter(stringConv{})
	RegisterConverter(intConv{})
	RegisterConverter(float64conv{})
	RegisterConverter(sliceIntConv{})
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
		case slimList:
			l = append(l, e)
		case slimString:
			l = append(l, e)
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
	//buf := bufio.NewReader(io.TeeReader(r, os.Stderr))
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
	replacer := []byte(ps.replacer)
	sep := ' '
	if '\n' != ps.lastByte {
		sep = '.'
	}
	fmt.Fprintf(ps.outStream, "%s%c:", replacer[1:5], sep)
	ps.outStream.Write(bytes.ReplaceAll(in, newline, replacer))
	ps.outStream.Write(newline)

	ps.lastByte = trailingByte
}
