package main

import (
	"html/template"
	"reflect"
	"strings"
	"syscall/js"
)

// setupSkeleton crafts a component subtree skeleton based on
// the template passed in, allowing later for attribute sanitation.
func setupSkeleton(html string) skeleton {
	skel := make(skeleton)

	for _, sub := range regTags.FindAllStringSubmatch(html, -1) {
		match := sub[1]
		tiers := strings.Split(match, ".")[1:]

		s := skel
		for i, tier := range tiers {
			if i >= len(tiers)-1 {
				s[tier] = nil
				continue
			}

			si, ok := s[tier]
			if !ok {
				sii := make(skeleton)
				si = sii
				s[tier] = si
				s = sii
			}

			s = si.(skeleton)
		}
	}

	return skel
}

// fillSkeleton recurses the skeleton subtree, plugging in mirrored
// component values and setting a template-safe context for inline
// events (as DOM attributes) to later be wired up as event handlers.
func fillSkeleton(s skeleton, comp interface{}, ctx ...string) {
	rv := reflect.ValueOf(comp)

	for fld, inf := range s {
		rfv := rv.FieldByName(fld)

		// this field in our skeleton is not an event subtree
		if rfv.IsValid() {
			switch rfv.Kind() {

			case reflect.Struct:
				skel, ok := inf.(skeleton)
				if !ok {
					panic(fld + " is a struct but not a skeleton")
				}

				fillSkeleton(skel, rfv.Interface())

			default:
				s[fld] = rfv.Interface()
			}
		} else {
			// comp isn't used after this point; event subtree only

			skel, ok := inf.(skeleton)
			if ok {
				fillSkeleton(skel, comp, append(ctx, fld)...)
				continue
			}

			s[fld] = template.HTMLAttr("{." + strings.Join(append(ctx, fld), ".") + "}")
		}
	}
}

// scrubInject strips preprocessed template patterns and injects event
// listeners in the form of channels, wiring up the original component.
func scrubInject(el js.Value, s skeleton, b bindings, compPtr interface{}) {
	kids := el.Get("childNodes")

	for i := 0; i < kids.Get("length").Int(); i++ {
		scrubInject(kids.Index(i), s, b, compPtr)
	}

	for attr := range attributes(el) {
		if !strings.Contains(attr, "{") {
			continue
		}

		ev := strings.Replace(attr, "{.", "", -1)
		ev = strings.Replace(ev, "}", "", -1)
		ctx := strings.Split(ev, ".")
		ev = ctx[len(ctx)-1]

		println("scrubbing " + attr + ", injecting " + ev)

		path := attrPath(attr, b)

		f, err := reflBinding(path, b, reflect.ValueOf(compPtr).Elem())
		if err != nil {
			panic(err)
		}

		ch, chE := setChan(f)

		// attr := attr
		el.Call("addEventListener", ev, js.NewEventCallback(js.StopPropagation, func(e js.Value) {
			// println(attr + " fired!")
			go func() {
				ch.Send(reflect.Zero(chE))
			}()
		}))
		el.Call("removeAttribute", attr)
	}
}
