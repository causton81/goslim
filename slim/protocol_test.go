package slim

import (
	"bytes"
	"testing"
)

func TestEncodeString(t *testing.T) {
	assertEqual(t, "000000:", slimString("").Slim())
	//assertEqual(t, "000000:", slimString(nil).Slim())
	assertEqual(t, "000006:Hello!", slimString("Hello!").Slim())
}

func TestEncodeList(t *testing.T) {
	//assertEqual(t, expected, slimList{slimString("hello"), slimString("world")}.Slim())
	assertEqual(t, "000035:[000002:000005:hello:000005:world:]", asList("hello", "world").Slim())
}

func TestLoadSlim(t *testing.T) {
	slim := loadSlim(bytes.NewBufferString("000000:"))
	assertEqual(t, "", string(slim.(slimString)))

	slim = loadSlim(bytes.NewBufferString("000001::"))
	assertEqual(t, ":", string(slim.(slimString)))

	slim = loadSlim(bytes.NewBufferString("000009:[000000:]"))
	l := slim.(slimList)
	if 0 != len(l) {
		t.Fatal()
	}

	slim = loadSlim(bytes.NewBufferString("000035:[000002:000005:hello:000005:world:]"))
	l = slim.(slimList)
	if 2 != len(l) {
		t.Fatal()
	}

	assertEqual(t, "hello", l[0].String())
	assertEqual(t, "world", l[1].String())
}

func TestStreams(t *testing.T) {
	w := new(bytes.Buffer)
	ps := newPrefixedStream(w, "SOUT")
	ps.write([]byte(""))
	assertEqual(t, "SOUT :\n", w.String())

	w.Reset();
	ps = newPrefixedStream(w, "SOUT")
	ps.write([]byte("x"))
	assertEqual(t, "SOUT :x\n", w.String())

	w.Reset();
	ps = newPrefixedStream(w, "SOUT")
	ps.write([]byte(""))
	ps.write([]byte("x"))
	assertEqual(t, "SOUT :\nSOUT.:x\n", w.String())

	w.Reset();
	ps = newPrefixedStream(w, "SOUT")
	ps.write([]byte("one\ntwo"))
	assertEqual(t, "SOUT :one\nSOUT :two\n", w.String())

	w.Reset();
	ps = newPrefixedStream(w, "SOUT")
	ps.write([]byte("one\n"))
	ps.write([]byte("two\n"))
	assertEqual(t, "SOUT :one\nSOUT :two\n", w.String())

	w.Reset();
	ps = newPrefixedStream(w, "SOUT")
	ps.write([]byte("3 "))
	ps.write([]byte("4\n"))
	assertEqual(t, "SOUT :3 \nSOUT.:4\n", w.String())
}

func assertEqual(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("%v does not equal expected value %v", actual, expected)
	}
}
