package main

import (
	"bytes"
	"fmt"
	"html/template"
	"syscall/js"
)

type App struct {
	Name     string
	Click    <-chan bool
	DblClick <-chan bool
	H1
	Test
}

type H1 struct {
	Text     string
	Click    <-chan bool
	DblClick <-chan bool
	Div
	string
}

type Div struct {
	Text  string
	Click <-chan bool
	Span
}

type Span struct {
	Text  string
	Click <-chan bool
}

type Test struct {
	Msg   string
	Focus <-chan bool
	Blur  <-chan bool
}

func main() {

	h1 := H1{
		Text: "Howdy",
		Div: Div{
			Text: "Partner",
			Span: Span{
				Text: "Alex",
			},
		},
		string: `
		<div	class="outer-div"
				style="	border: 1px dotted black;
						cursor: pointer"
				{{.Click}}
				{{.DblClick}}>

			{{.Div.Span.Text}}
			{{if true}}

				<div	class="inner-div"
						style="border: 1px solid black"
						{{.Div.Click}}>

					{{.Div.Text}}

					<span	class="inner-span"
							style="color: white; background: black"
							{{.Div.Span.Click}}>

						{{.Span.Text}}
						{{.Text}}
					</span>
				</div>

			{{else}}

				<span>
					{{.Div.Text}}
				</span>
				
			{{end}}
		</div>
		`,
	}

	html := h1.string

	t := htmlToTemplate(html)
	// el := templateToElement(t)
	html = t.Get("innerHTML").String()

	b, err := origBindings(h1.string, html)
	if err != nil {
		panic(err)
	}

	fmt.Println(b)

	skel := setupSkeleton(html)

	fmt.Println(skel)

	fillSkeleton(skel, h1)

	fmt.Println(skel)

	// inject
	tmpl := template.New("x")

	tmpl, err = tmpl.Parse(html)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, skel)
	if err != nil {
		panic(err)
	}

	println(buf.String())

	t = htmlToTemplate(buf.String())
	el := templateToElement(t)

	scrubInject(el, skel, b, &h1)

	js.Global().Get("document").Get("body").Call("appendChild", el)

	for {
		select {
		case <-h1.Click:
			println("<-h1.Click!!!")
		case <-h1.DblClick:
			println("<-h1.DblClick!!!")
		case <-h1.Span.Click:
			println("<-h1.Span.Click!!!")
		case <-h1.Div.Click:
			println("<-h1.Div.Click!!!")
		}
	}
}
