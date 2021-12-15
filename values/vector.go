package values

import (
	"regexp"

	"github.com/influxdata/flux/semantic"
)

type Vector interface {
	Value
	Vector() Vector
	Get(i int) Value
	Set(i int, value Value)
	Len() int
}

type vector struct {
	t        semantic.MonoType
	elements []Value
}

func (v *vector) Vector() Vector {
	return v
}
func (v *vector) Get(i int) Value {
	return v.elements[i]
}
func (v *vector) Set(i int, value Value) {
	v.elements[i] = value
}
func (v *vector) Len() int {
	return len(v.elements)
}
func (v *vector) Range(f func(i int, v Value)) {
	for i, n := 0, v.Len(); i < n; i++ {
		f(i, v.Get(i))
	}
}

func (v *vector) Type() semantic.MonoType {
	return v.t
}
func (v *vector) IsNull() bool {
	return false
}
func (v *vector) Str() string {
	var s string
	return s
}
func (v *vector) Bytes() []byte {
	var b []byte
	return b
}
func (v *vector) Int() int64 {
	var i int64
	return i
}
func (v *vector) UInt() uint64 {
	var u uint64
	return u
}
func (v *vector) Float() float64 {
	var f float64
	return f
}
func (v *vector) Bool() bool {
	var b bool
	return b
}
func (v *vector) Time() Time {
	var t Time
	return t
}
func (v *vector) Duration() Duration {
	var d Duration
	return d
}
func (v *vector) Regexp() *regexp.Regexp {
	var r *regexp.Regexp
	return r
}
func (v *vector) Array() Array {
	var a Array
	return a
}
func (v *vector) Object() Object {
	var o Object
	return o
}
func (v *vector) Function() Function {
	var f Function
	return f
}
func (v *vector) Dict() Dictionary {
	var d Dictionary
	return d
}
func (v *vector) Equal(Value) bool {
	return false
}
