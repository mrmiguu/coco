package main

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"syscall/js"
)

// regTags is a regular expression that matches Coco template binding tags.
var regTags = regexp.MustCompile(`{{\s*((\.[A-Za-z]+)+)\s*}}`)

// stringErr is a general purpose string-error tuple.
type stringErr struct {
	string
	error
}

// field
type field struct {
	reflect.Value
	reflect.StructField
}

// skeleton is a wireframe representation of a struct component.
type skeleton map[string]interface{}

// bindings is a mapping of current lowercase attributes to their original title case tags.
type bindings map[string]string

// name reflects on the component to get its name.
func name(comp interface{}) string {
	v := reflect.ValueOf(comp)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.Type().Name()
}

// origBindings maps attributes to their original title case bindings.
func origBindings(raw, tmpl string) (bindings, error) {
	raw = sanitize(raw)
	pre := regTags.FindAllString(raw, -1)
	post := regTags.FindAllString(tmpl, -1)

	if len(pre) != len(post) {
		return nil, errors.New("origBindings: tmpl is not raw, processed")
	}

	b := make(bindings)

	for i := 0; i < len(pre); i++ {
		if strings.ToLower(pre[i]) == post[i] {
			b[post[i]] = pre[i]
		}
	}

	return b, nil
}

// attrPath converts a dot-separated binding tag into a path.
func attrPath(attr string, b bindings) []string {
	title := b["{"+attr+"}"]
	title = strings.Replace(title, "{{.", "", -1)
	title = strings.Replace(title, "}}", "", -1)

	println(title)

	return strings.Split(title, ".")
}

// reflBinding travels up a path, reflecting on a struct hierarchy.
func reflBinding(path []string, b bindings, rv reflect.Value) (field, error) {
	if len(path) == 0 {
		return field{}, nil
	}

	name, tiers := path[0], path[1:]
	rfv := rv.FieldByName(name)

	if len(tiers) > 0 {
		return reflBinding(tiers, b, rfv)
	}

	rft, ok := rv.Type().FieldByName(name)
	if !ok {
		return field{}, errors.New("reflBinding: StructField reflection failed")
	}

	return field{rfv, rft}, nil
}

// setChan creates a channel of the base type for the field, setting
// the new one and returning it and its base type for usage.
func setChan(f field) (reflect.Value, reflect.Type) {
	name := f.StructField.Name
	chE := f.StructField.Type.Elem()

	println("field " + name + ": " + chE.Name() + " channel")

	chT := reflect.ChanOf(reflect.BothDir, chE)
	ch := reflect.MakeChan(chT, 0)

	f.Value.Set(ch)

	return ch, chE
}

// sanitize readies raw HTML for the template element to parse.
func sanitize(html string) string {
	html = strings.Replace(html, "\n", "", -1)
	html = strings.Trim(html, " \t")
	reg := regexp.MustCompile(`>\s+<`)
	html = reg.ReplaceAllString(html, "><")

	return html
}

// htmlToTemplate converts raw HTML into a JS template element.
func htmlToTemplate(html string) js.Value {
	t := js.Global().Get("document").Call("createElement", "template")
	t.Set("innerHTML", sanitize(html))
	return t
}

// templateToElement unwraps the base JS element from the template element.
func templateToElement(t js.Value) js.Value {
	return t.Get("content").Get("firstChild")
}

// attributes parses a JS element's attributes into a key-value map.
func attributes(el js.Value) map[string]string {
	if el.Get("hasAttributes") == js.Undefined() || !el.Call("hasAttributes").Bool() {
		return nil
	}

	m := make(map[string]string)

	attrs := el.Get("attributes")
	for i := 0; i < attrs.Get("length").Int(); i++ {
		attr := attrs.Index(i)
		m[attr.Get("name").String()] = attr.Get("value").String()
	}

	return m
}

// define wraps a raw HTML component in a define tag.
func define(name, html string) string {
	return `{{define "` + name + `"}}` + html + "{{end}}"
}

// fetch grabs a file from the server.
func fetch(file string) <-chan stringErr {
	c := make(chan stringErr)

	js.Global().Call("fetch", file).Call("then", js.NewEventCallback(0, func(e js.Value) {

		ok := e.Get("ok").Bool()
		if !ok {
			c <- stringErr{
				"",
				errors.New(file + " not found"),
			}
			return
		}

		e.Call("text").Call("then", js.NewEventCallback(0, func(e js.Value) {

			println(e.String())

			c <- stringErr{
				e.String(),
				nil,
			}

		})).Call("catch", js.NewEventCallback(0, func(e js.Value) {
			println(e.Get("message").String())
		}))

	})).Call("catch", js.NewEventCallback(0, func(e js.Value) {
		println(e.Get("message").String())
	}))

	return c
}
