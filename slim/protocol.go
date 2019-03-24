package slim

import (
	"bufio"
	"bytes"
	"fmt"
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
	Must(err)

	stderrReadSide, stderrWriteSide, err := os.Pipe()
	Must(err)

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
					Must(err)
					data := buf[0:numRead]
					psOut.write(data)
					//final := bytes.ReplaceAll(data, newline, stdoutStart)
					//realStderr.Write(final)
				}
				if 0 < unix.POLLIN&fds[1].Revents {
					numRead, err := stderrReadSide.Read(buf)
					Must(err)
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

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func RegisterFixture(fix interface{}) {
	//log.Println(reflect.TypeOf(fix).String())
	RegisterFixtureWithName(fix, reflect.TypeOf(fix).String())
}

var fixtureTypes = make(map[string]reflect.Type)

func RegisterFixtureWithName(fix interface{}, scriptAlias string) {
	fixtureTypes[scriptAlias] = reflect.TypeOf(fix)
}

func ListenAndServe() {
	slimIn = os.Stdin
	slimOut = os.Stdout
	redirectStdoutAndStderr()

	slimOut.WriteString("Slim -- V0.5\n")
	instructions := loadSlim(slimIn)
	slimResults := slimList{}
	log.Println("Slim Instructions:")
	for _, inst := range instructions.(slimList) {
		log.Println(inst)
		inst := inst.(slimList)
		id := inst[0].String()
		op := inst[1].String()

		switch op {
		case "make":
			//instanceName := inst[2].String()
			className := inst[3].String()
			//typ, found := fixtureTypes[className]
			slimResults = append(slimResults, asList(id, fmt.Sprintf("__EXCEPTION__:message:<<NO_CLASS %s>>", className)))
		default:
			slimResults = append(slimResults, asList(id, fmt.Sprintf("__EXCEPTION__:message:<<MALFORMED_INSTRUCTION %s>>", op)))
		}
	}

	fmt.Fprintf(slimOut, "%s", slimResults.Slim())
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
		Must(err)

		isList := '[' == peekBytes[0]
		if isList {
			//return slimString("TODO")
			buf.Discard(1)
			numElems := parseLength(buf)
			l := make(slimList, numElems)
			for i := range l {
				l[i] = loadSlim(buf)
				nextByte, err := buf.ReadByte()
				Must(err)

				if ':' != nextByte {
					log.Fatalf("expected next byte to be :")
				}
			}

			nextByte, err := buf.ReadByte()
			Must(err)
			if ']' != nextByte {
				log.Fatalf("expected next byte to be ]")
			}

			return l
		} else {
			bld := new(strings.Builder)
			bld.Grow(length)
			_, err := io.CopyN(bld, buf, int64(length))
			Must(err)

			return slimString(bld.String())
		}
	}
}

func parseLength(buf *bufio.Reader) int {
	sizeField, err := buf.ReadString(':')
	Must(err)
	size, err := strconv.ParseInt(sizeField[0:len(sizeField)-1], 10, 0)
	Must(err)
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
