# 🥥 Coco
Golang WebAssembly Framework

app.go

```go
type App struct {
    Counter int
}

func (a App) OnTestClick() {
    a.Counter++
    coco.Set(a)
}
```

app.html

```html
<link href="app.css" rel="stylesheet">

<div class="test">
  {{ .Counter }}
</div>
```

app.css

```css
.test {
  width: 160px;
}
```

## Installation

```sh
go get github.com/mrmiguu/coco
```

## Running the example

You should use [go-wasm-cli](https://github.com/mfrachet/go-wasm-cli) to run this example.

```sh
cd $GOPATH/src/github.com/mrmiguu/coco/example
go-wasm start
```
