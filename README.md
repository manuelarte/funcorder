# 🧐gofuncor
Golang Linter to check Functions/Methods Order.

## 🚀 Features

### Check exported methods are placed before not exported methods

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
// placed after exported method
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

This rule checks that the `Consturctor` functions are placed after the struct declaration and before the struct's methods.

<details>
  <summary>Constructor function</summary>

> [!NOTE]  
> This linter considers a constructor function a function that has the prefix *New*, or *Must*, and returns 1 or 2 types.
> Where the 1st return type is an struct declared in the same file.

</details>

<table>
<thead><tr><th>❌ Bad</th><th>✅ Good</th></tr></thead>
<tbody>
<tr><td>

```go
// ❌ constructor "NewMyStruct" placed 
// before the struct declaration
func NewMyStruct() MyStruct {
    return MyStruct2{Name: "John"}
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
    return MyStruct2{Name: "John"}
}

// other MyStruct's methods
...
```

</td></tr>

</tbody>
</table>

## Resources

+ Following [uber guidelines](https://github.com/uber-go/guide/blob/master/style.md#function-grouping-and-ordering) 
