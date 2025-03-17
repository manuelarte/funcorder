# 🧐gofuncor
Golang Linter to check Functions Order.

## 🚀 Features

### Check exported methods are placed before not exported methods

This rule checks that the exported method are placed before the unexported ones, e.g:

```go
type MyStruct struct {
	Name string
}

// ❌ `unexported method "lenName" for struct "MyStruct" should be placed after the exported method "GetName"`
func (m MyStruct) lenName() int { 
	return len(m.Name)
}

func (m MyStruct) GetName() string {
	return m.Name
}
...
```

### Check `Constructors` methods are placed after struct declaration

This rule checks that the `Consturctor` functions are placed after the struct declaration and before the struct's methods.

<details>
  <summary>Constructor method</summary>

> [!NOTE]
> This linter considers a constructor function a function that has the prefix *New*, or *Must*, and returns 1 or 2 types.
> Where the 1st return type is an struct declared in the same file.
</details>

```go
// ❌ `constructor "NewMyStruct" should be placed after the struct declaration`
func NewMyStruct() MyStruct {
    return MyStruct2{Name: "John"}
}

type MyStruct struct {
	Name string
}

// `unexported method "lenName" for struct "MyStruct" should be placed after the exported method "GetName"`
func (m MyStruct) lenName() int { 
	return len(m.Name)
}

func (m MyStruct) GetName() string {
	return m.Name
}
...
```

## Resources

Following [uber guidelines](https://github.com/uber-go/guide/blob/master/style.md#function-grouping-and-ordering) 
