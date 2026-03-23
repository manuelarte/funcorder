# 🧐 FuncOrder

[![Go Report Card](https://goreportcard.com/badge/github.com/manuelarte/funcorder)](https://goreportcard.com/report/github.com/manuelarte/funcorder)
![version](https://img.shields.io/github/v/release/manuelarte/funcorder)

- [🧐 FuncOrder](#-funcorder)
  - [⬇️ Getting Started](#️-getting-started)
    - [As A Golangci-lint linter](#as-a-golangci-lint-linter)
    - [Standalone application](#standalone-application)
  - [🚀 Features](#-features)
    - [Check exported methods are placed before unexported methods](#check-exported-methods-are-placed-before-unexported-methods)
    - [Check `Constructors` functions are placed after struct declaration](#check-constructors-functions-are-placed-after-struct-declaration)
    - [Check Constructors/Methods are sorted alphabetically](#check-constructorsmethods-are-sorted-alphabetically)
    - [Check exported functions are placed before unexported functions](#check-exported-functions-are-placed-before-unexported-functions)
  - [Resources](#resources)

Go Linter to check Functions/Methods Order.

## ⬇️ Getting Started

### As a golangci-lint linter

Define the rules in your `golangci-lint` configuration file, e.g:

```yaml
linters:
  enable:
    - funcorder
    ...

  settings:
    funcorder:
      # Checks that constructors are placed after the structure declaration.
      # Default: true
      constructor: false
      # Checks if the exported methods of a structure are placed before the unexported ones.
      # Default: true
      struct-method: false
      # Checks if the constructors and/or structure methods are sorted alphabetically.
      # Default: false
      alphabetical: true
      # Checks that exported functions are placed before unexported functions.
      # Default: false
      function: true
```

### Standalone application

Install FuncOrder linter using

```bash
go install github.com/manuelarte/funcorder@latest
```

And then use it with

```
funcorder [-constructor=true|false] [-struct-method=true|false] [-alphabetical=true|false] [-function=true|false] ./...
```

Parameters:

- `constructor`: `true|false` (default `true`) Checks that constructors are placed after the structure declaration.
- `struct-method`: `true|false` (default `true`) Checks if the exported methods of a structure are placed before the unexported ones.
- `alphabetical`: `true|false` (default `false`) Checks if the constructors and/or structure methods are sorted alphabetically.
- `function`: `true|false` (default `false`) Checks that exported functions are placed before unexported functions.

## 🚀 Features

### Check exported methods are placed before unexported methods

This rule checks that the exported method are placed before the unexported ones, e.g:

<table>
<thead><tr><th>❌ Bad</th><th>✅ Good</th></tr></thead>
<tbody>
<tr><td>

```go
type MyStruct struct {
 Name string
}

// ❌ unexported method 
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
// unexported methods
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

### Check Constructors/Methods are sorted alphabetically

This rule checks:

- `Constructor` functions are sorted alphabetically (if `constructor` setting/parameter is `true`).
- `Methods` are sorted alphabetically (if `struct-method` setting/parameter is `true`) for each group (exported and unexported).

<table>
<thead><tr><th>❌ Bad</th><th>✅ Good</th></tr></thead>
<tbody>
<tr><td>

```go
type MyStruct struct {
    Name string
}

func NewMyStruct() MyStruct {
    return MyStruct{Name: "John"}
}

// ❌ constructor "NewAMyStruct" should be placed 
// before "NewMyStruct"
func NewAMyStruct() MyStruct {
    return MyStruct{Name: "John"}
}

func (m MyStruct) GoodMorning() string {
    return "good morning"
}

// ❌ method "GoodAfternoon" should be placed 
// before "GoodMorning"
func (m MyStruct) GoodAfternoon() string {
    return "good afternoon"
}

func (m MyStruct) hello() string {
 return "hello"
}

// ❌ method "bye" should be placed 
// before "hello"
func (m MyStruct) bye() string {
    return "bye"
}

...
```

</td><td>

```go
type MyStruct struct {
    Name string
}

// ✅ constructors sorted alphabetically
func NewAMyStruct() MyStruct {
    return MyStruct{Name: "John"}
}

func NewMyStruct() MyStruct {
    return MyStruct{Name: "John"}
}

// ✅ exported methods sorted alphabetically
func (m MyStruct) GoodAfternoon() string {
    return "good afternoon"
}

func (m MyStruct) GoodMorning() string {
    return "good morning"
}

// ✅ unexported methods sorted alphabetically
func (m MyStruct) bye() string {
    return "bye"
}

func (m MyStruct) hello() string {
    return "hello"
}

...
```

</td></tr>

</tbody>
</table>

### Check exported functions are placed before unexported functions

This rule checks that exported functions (those with no receiver) are placed before unexported ones within each file. The `init` function is excluded from this rule.

<table>
<thead><tr><th>❌ Bad</th><th>✅ Good</th></tr></thead>
<tbody>
<tr><td>

```go
// ❌ unexported function placed
// before exported function
func helper() string {
    return "helper"
}

func PublicFunc() string {
    return "public"
}
```

</td><td>

```go
// ✅ exported function placed
// before unexported function
func PublicFunc() string {
    return "public"
}

func helper() string {
    return "helper"
}
```

</td></tr>

</tbody>
</table>

> **Note:** When this change is merged upstream and consumed by golangci-lint, enable this rule by adding `function: true` under `linters.settings.funcorder` in your golangci-lint configuration.

## Resources

- Following Uber Style Guidelines about [function-grouping-and-ordering](https://github.com/uber-go/guide/blob/master/style.md#function-grouping-and-ordering)
