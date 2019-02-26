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