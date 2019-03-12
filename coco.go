package coco

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"syscall/js"
)

var (
	vdom = VDOM{
		cache: make(map[string]string),
	}
)

// VDOM is the virtual DOM for this package.
type VDOM struct {
	cache   map[string]string
	globals interface{}
}

// Globals registers the global store component.
func Globals(comp interface{}) {
	vdom.Globals(comp)
}

// Globals registers the global store component.
func (v *VDOM) Globals(comp interface{}) {
	println("registering globals (" + getName(comp) + ")")
	v.globals = comp
}

// Render renders the root component as the body.
func Render(root interface{}) {
	vdom.Render(root)
}

// Render renders the root component as the body.
func (v *VDOM) Render(root interface{}) {
	fnMap := make(map[string]reflect.Value)
	for _, comp := range append(v.getAllEmbed(root, v.isComp), root) {
		for fn, v := range getFuncs(comp) {
			fnMap[fn] = v
		}
	}

	bin, err := v.compile(root)
	if err != nil {
		panic(err)
	}

	setBody(bin)

	for fn, v := range fnMap {
		cls := strings.Replace(fn, "On", "", -1)
		cls = strings.Replace(cls, "Click", "", -1)
		elements := js.Global().Get("document").Call("getElementsByClassName", cls)

		v := v
		for i := 0; i < elements.Get("length").Int(); i++ {
			elements.Index(i).Call("addEventListener", "click", js.NewCallback(func(args []js.Value) {
				v.Call([]reflect.Value{})
			}))
		}
	}

	select {}
}

// Set sets the diff component and patches the * tree.
func Set(comp interface{}) {
	vdom.Set(comp)
}

// Set sets the diff component and patches the * tree.
func (v *VDOM) Set(comp interface{}) {
	name := getName(comp)

	fnMap := make(map[string]reflect.Value)
	for _, comp := range append(v.getAllEmbed(comp, v.isComp), comp) {
		for fn, val := range getFuncs(comp) {
			fnMap[fn] = val
		}
	}

	glName := getName(v.globals)
	newGl := reflect.ValueOf(comp).FieldByName(glName)
	println(glName+" = ", newGl.FieldByName("Cocos").Len())

	bin, err := v.compile(comp)
	if err != nil {
		panic(err)
	}

	elements := js.Global().Get("document").Call("getElementsByClassName", name)
	for i := 0; i < elements.Get("length").Int(); i++ {
		element := elements.Index(i)
		parent := element.Get("parentElement")
		parent.Call("replaceChild", ParseDOM(bin), element)
	}

	for fn, val := range fnMap {
		// println(fn + " :: " + val.String())

		cls := strings.Replace(fn, "On", "", -1)
		cls = strings.Replace(cls, "Click", "", -1)
		elements := js.Global().Get("document").Call("getElementsByClassName", cls)

		val := val
		for i := 0; i < elements.Get("length").Int(); i++ {
			println(fn + " SET loops (inner)")
			elements.Index(i).Call("addEventListener", "click", js.NewCallback(func(args []js.Value) {
				val.Call([]reflect.Value{})
			}))
		}
	}
}

// Patch compares the component and the DOM element and patches if necessary.
func Patch(comp interface{}) {
	vdom.patch(comp)
}

func (v *VDOM) patch(comp interface{}) {
	doms := js.Global().Get("document").Call("getElementsByClassName", getName(comp))
	println(doms.Get("length").Int())
	for i := 0; i < doms.Get("length").Int(); i++ {
		dom := doms.Index(i)
		v.patchPair(comp, dom)
	}
}

func (v *VDOM) patchPair(comp interface{}, dom js.Value) {
	el, err := MustParseDOM(v.compile(comp))
	if err != nil {
		panic(err)
	}
	println(el.Get("children").Get("length").Int())
}

// Compile compiles the component down into raw HTML.
func Compile(comp interface{}) (string, error) {
	return vdom.compile(comp)
}

func (v *VDOM) compile(comp interface{}) (string, error) {
	bin := ""
	for _, c := range v.getAllEmbed(comp, v.isLocalComp) {
		name := getName(c)
		b, err := v.compile(c)
		if err != nil {
			return "", err
		}
		bin += define(name, b)
	}

	name := reflect.TypeOf(comp).Name()
	t := template.New(name + ".coco")

	raw, ok := v.cache[name]
	if !ok {
		println("fetching " + name + ".html")
		r, e := fetch(name + ".html")
		if err := <-e; err != nil {
			return "", err
		}
		raw = <-r
		v.cache[name] = raw
	}

	t, err := t.Parse(bin + raw)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)

	err = t.Execute(buf, comp)
	if err != nil {
		return "", err
	}

	oldGl := fmt.Sprint(reflect.ValueOf(v.globals).Interface())
	newGl := fmt.Sprint(reflect.ValueOf(comp).FieldByName(getName(v.globals)).Interface())

	println("compile: oldGl", oldGl)
	println("compile: newGl", newGl)

	b := oldGl == newGl

	println("compile: oldGlobals == newGlobals? " + strconv.FormatBool(b))

	return buf.String(), nil
}

func fetch(file string) (<-chan string, <-chan error) {
	c := make(chan string, 1)
	e := make(chan error, 1)

	js.Global().Call("fetch", file).Call("then", js.NewCallback(func(args []js.Value) {

		ok := args[0].Get("ok").Bool()
		if !ok {
			c <- ""
			e <- errors.New(file + " not found.")
			return
		}

		args[0].Call("text").Call("then", js.NewCallback(func(args []js.Value) {

			c <- args[0].String()
			e <- nil

		})).Call("catch", js.NewCallback(func(args []js.Value) {
			println(args[0].Get("message").String())
		}))

	})).Call("catch", js.NewCallback(func(args []js.Value) {
		println(args[0].Get("message").String())
	}))

	return c, e
}

type strErrC struct {
	c <-chan string
	e <-chan error
}

func newStrErrC(c <-chan string, e <-chan error) strErrC {
	return strErrC{c, e}
}

// MustParseDOM converts raw HTML into an element.
func MustParseDOM(html string, err error) (js.Value, error) {
	return ParseDOM(html), err
}

// ParseDOM converts raw HTML into an element.
func ParseDOM(html string) js.Value {
	t := js.Global().Get("document").Call("createElement", "template")
	html = strings.Trim(html, " ")
	t.Set("innerHTML", html)
	return t.Get("content").Get("firstChild")
}

// Root grabs the anchor element for the index.html page.
func Root() js.Value {
	return js.Global().Get("document").Get("body")
}

// Append appaends the child onto the parent node.
func Append(parent, child js.Value) {
	parent.Call("appendChild", child)
}

func setBody(html string) {
	js.Global().Get("document").Get("body").Set("innerHTML", html)
}

func define(name, html string) string {
	return `{{define "` + name + `"}}` + `<link href="` + name + `.css" rel="stylesheet">` + html + "{{end}}"
}

func getName(comp interface{}) string {
	t := reflect.TypeOf(comp)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

func (v *VDOM) isComp(comp interface{}, i int) bool {
	y := reflect.TypeOf(comp)
	return y.Field(i).Name == y.Field(i).Type.Name()
}

func (v *VDOM) isLocalComp(comp interface{}, i int) bool {
	y := reflect.TypeOf(comp)
	g := getName(v.globals)
	name := y.Field(i).Type.Name()
	return y.Field(i).Name == name && g != name
}

func (v *VDOM) getAllEmbed(comp interface{}, keep func(interface{}, int) bool) []interface{} {
	a := make([]interface{}, 0)
	x := reflect.ValueOf(comp)
	for i := 0; i < x.NumField(); i++ {
		if !keep(comp, i) {
			continue
		}
		a = append(a, x.Field(i).Interface())
	}
	s := []string{}
	a2 := make([]interface{}, 0)
	for _, emb := range a {
		a2 = append(a2, v.getAllEmbed(emb, keep)...)
	}
	a2 = append(a2, a...)
	for _, emb := range a2 {
		s = append(s, reflect.TypeOf(emb).Name())
	}
	return a2
}

func getEmbed(comp interface{}) []interface{} {
	a := make([]interface{}, 0)
	x := reflect.ValueOf(comp)
	y := reflect.TypeOf(comp)
	for i := 0; i < x.NumField(); i++ {
		if y.Field(i).Name != y.Field(i).Type.Name() {
			continue
		}
		a = append(a, x.Field(i).Interface())
	}
	return a
}

func getFields(comp interface{}) []string {
	a := []string{}
	x := reflect.ValueOf(comp)
	y := reflect.TypeOf(comp)
	for i := 0; i < x.NumField(); i++ {
		a = append(a, y.Field(i).Type.Name())
	}
	return a
}

func getFuncs(comp interface{}) map[string]reflect.Value {
	m := make(map[string]reflect.Value)
	x := reflect.ValueOf(comp)
	y := reflect.TypeOf(comp)
	for i := 0; i < x.NumMethod(); i++ {
		method := x.Method(i)
		m[y.Method(i).Name] = method
	}
	return m
}
