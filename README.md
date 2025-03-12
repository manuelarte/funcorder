# gofuncor
Golang Linter to check Functions Order

## Features

- Check exported methods before not exported methods
- `New(MyStruct)` is after struct declaration but before methods declaration

Following [uber guidelines](https://github.com/uber-go/guide/blob/master/style.md#function-grouping-and-ordering) 
