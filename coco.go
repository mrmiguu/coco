package coco

import (
	"bytes"
	"errors"
	"html/template"
	"reflect"
	"strings"
	"syscall/js"
)

var (
	vdom = VDOM{
		make(map[string]string),
	}
)

// VDOM is the virtual DOM for this package.
type VDOM struct {
	cache map[string]string
}

// Render renders the root component as the body.
func Render(root interface{}) {
	vdom.Render(root)
}

// Render renders the root component as the body.
func (v *VDOM) Render(root interface{}) {
	fnMap := make(map[string]reflect.Value)
	for _, comp := range append(bfsEmbedded(root), root) {
		for fn, v := range getFuncs(comp) {
			fnMap[fn] = v
		}
	}

	bin, err := v.compile(root)
	if err != nil {
		panic(err)
	}

	updateDOM(bin)

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
	for _, comp := range append(bfsEmbedded(comp), comp) {
		for fn, v := range getFuncs(comp) {
			fnMap[fn] = v
		}
	}

	bin, err := v.compile(comp)
	if err != nil {
		panic(err)
	}

	elements := js.Global().Get("document").Call("getElementsByClassName", name)
	for i := 0; i < elements.Get("length").Int(); i++ {
		element := elements.Index(i)
		parent := element.Get("parentElement")
		parent.Call("replaceChild", htmlToElement(bin), element)
	}

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
}

func (v *VDOM) compile(comp interface{}) (string, error) {
	bin := ""
	for _, c := range bfsEmbedded(comp) {
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

func htmlToElement(html string) js.Value {
	template := js.Global().Get("document").Call("createElement", "template")
	html = strings.Trim(html, " ")
	template.Set("innerHTML", html)
	return template.Get("content").Get("firstChild")
}

func updateDOM(html string) {
	js.Global().Get("document").Get("body").Set("innerHTML", html)
}

func define(name, html string) string {
	return `{{define "` + name + `"}}` + `<link href="` + name + `.css" rel="stylesheet">` + html + "{{end}}"
}

func getName(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

func bfsEmbedded(v interface{}) []interface{} {
	a := make([]interface{}, 0)
	x := reflect.ValueOf(v)
	y := reflect.TypeOf(v)
	for i := 0; i < x.NumField(); i++ {
		if y.Field(i).Name != y.Field(i).Type.Name() {
			continue
		}
		a = append(a, x.Field(i).Interface())
	}
	s := []string{}
	a2 := make([]interface{}, 0)
	for _, emb := range a {
		a2 = append(a2, bfsEmbedded(emb)...)
	}
	a2 = append(a2, a...)
	for _, emb := range a2 {
		s = append(s, reflect.TypeOf(emb).Name())
	}
	return a2
}

func getEmbedded(v interface{}) []interface{} {
	a := make([]interface{}, 0)
	x := reflect.ValueOf(v)
	y := reflect.TypeOf(v)
	for i := 0; i < x.NumField(); i++ {
		if y.Field(i).Name != y.Field(i).Type.Name() {
			continue
		}
		a = append(a, x.Field(i).Interface())
	}
	return a
}

func getFields(v interface{}) []string {
	a := []string{}
	x := reflect.ValueOf(v)
	y := reflect.TypeOf(v)
	for i := 0; i < x.NumField(); i++ {
		a = append(a, y.Field(i).Type.Name())
	}
	return a
}

func getFuncs(v interface{}) map[string]reflect.Value {
	m := make(map[string]reflect.Value)
	x := reflect.ValueOf(v)
	y := reflect.TypeOf(v)
	for i := 0; i < x.NumMethod(); i++ {
		method := x.Method(i)
		m[y.Method(i).Name] = method
	}
	return m
}
