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

App.html

```html
<div class="Test">
  {{.Counter}}
</div>
```

App.css

```css
.Test {
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
