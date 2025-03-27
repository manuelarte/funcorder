[![Go Report Card](https://goreportcard.com/badge/github.com/manuelarte/funcorder)](https://goreportcard.com/report/github.com/manuelarte/funcorder)
![version](https://img.shields.io/github/v/release/manuelarte/funcorder)
- [🧐 FuncOrder](#-funcorder)
    * [⬇️ Getting Started](#-getting-started)
    * [🚀 Features](#-features)
        + [Check exported methods are placed before non-exported methods](#check-exported-methods-are-placed-before-non-exported-methods)
        + [Check `Constructors` functions are placed after struct declaration](#check-constructors-functions-are-placed-after-struct-declaration)
    * [Resources](#resources)

# 🧐 FuncOrder
Go Linter to check Functions/Methods Order.

## ⬇️ Getting Started

Install FuncOrder linter using

> go install github.com/manuelarte/funcorder@latest

And then use it with

> funcorder ./...

## 🚀 Features

### Check exported methods are placed before non-exported methods

This rule checks that the exported method are placed before the non-exported ones, e.g:

<table>
<thead><tr><th>❌ Bad</th><th>✅ Good</th></tr></thead>
<tbody>
<tr><td>

```go
type MyStruct struct {
	Name string
}

// ❌ non-exported method 
// placed before exported method
func (m MyStruct) lenName() int { 
	return len(m.Name)
}

func (m MyStruct) GetName() string {
	return m.Name
}
...
```

</td><td>

```go
type MyStruct struct {
	Name string
}

// ✅ exported methods before 
// non-exported methods
func (m MyStruct) GetName() string {
	return m.Name
}

func (m MyStruct) lenName() int {
    return len(m.Name)
}
...
```

</td></tr>

</tbody>
</table>

### Check `Constructors` functions are placed after struct declaration

This rule checks that the `Constructor` functions are placed after the struct declaration and before the struct's methods.

<details>
  <summary>Constructor function</summary>

> [!NOTE]  
> This linter considers a Constructor function a function that has the prefix *New*, or *Must*, and returns 1 or 2 types.
> Where the 1st return type is a struct declared in the same file.

</details>

<table>
<thead><tr><th>❌ Bad</th><th>✅ Good</th></tr></thead>
<tbody>
<tr><td>

```go
// ❌ constructor "NewMyStruct" placed 
// before the struct declaration
func NewMyStruct() MyStruct {
    return MyStruct{Name: "John"}
}

type MyStruct struct {
    Name string
}

...
```

</td><td>

```go
type MyStruct struct {
    Name string
}

// ✅ `constructor "NewMyStruct" placed 
// after the struct declaration 
// and before the struct's methods`
func NewMyStruct() MyStruct {
    return MyStruct{Name: "John"}
}

// other MyStruct's methods
...
```

</td></tr>

</tbody>
</table>

## Resources

+ Following Uber Style Guidelines about [function-grouping-and-ordering](https://github.com/uber-go/guide/blob/master/style.md#function-grouping-and-ordering) 
