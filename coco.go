package coco

import (
	"bytes"
	"html/template"
	"reflect"
	"strings"
	"syscall/js"
)

func getType(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
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

// Set sets the diff component and patches the VDOM tree.
func Set(comp interface{}) {
	render(comp)
}

// Render renders the root component as the body.
func Render(root interface{}) {
	render(root)
	select {}
}

func render(root interface{}) {
	html := strings.ToLower(getType(root)) + ".html"

	js.Global().Call("fetch", html).Call("then", js.NewCallback(func(args []js.Value) {
		args[0].Call("text").Call("then", js.NewCallback(func(args []js.Value) {

			t := template.New(html)

			fns := getFuncs(root)
			fnMap := make(template.FuncMap)
			for fn, v := range fns {
				fn := fn
				v := v
				fnMap[fn] = func(args ...interface{}) string {
					v.Call([]reflect.Value{})
					return "_X_"
				}
			}
			t = t.Funcs(fnMap)

			t, err := t.Parse(args[0].String())
			if err != nil {
				panic(err)
			}

			buf := new(bytes.Buffer)

			t.Execute(buf, root)
			document := js.Global().Get("document")
			document.Get("body").Set("innerHTML", buf.String())

			for fn := range fns {
				fn := fn
				cls := strings.Replace(fn, "On", "", -1)
				cls = strings.Replace(cls, "Click", "", -1)
				cls = strings.ToLower(cls)
				elements := document.Call("getElementsByClassName", cls)
				for i := 0; i < elements.Get("length").Int(); i++ {
					elements.Index(i).Call("addEventListener", "click", js.NewCallback(func(args []js.Value) {
						fnMap[fn].(func(...interface{}) string)()
					}))
				}
			}

		}))
	}))
}
