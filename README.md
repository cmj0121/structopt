# Struct Opt #

![Action](https://github.com/cmj0121/structopt/workflows/pipeline/badge.svg)

The Go-based struct-based command-line argument parser. The structopt will process
the tag in the [struct][0] and generate the related command-line interface tool.

By-default the boolean type in struct field will be consider as the flip option,
the general type is the flag option and the pointer would be argument or sub-command,
depenends on the type is struct or not.

| Type       | Usage       | Description                          |
|------------|-------------|--------------------------------------|
| bool       | flip        | store the true/false value           |
| Type       | flag        | store the pre-defined optional value |
| \*Type     | argument    | store as the necessary argument      |
| \*Struct   | sub-command | as the sub-command                   |


```go
package main

import (
	"github.com/cmj0121/structopt"
)

type Example struct {
	structopt.Help
}

func main() {
	example := Example{}
	parser := structopt.MustNew(&example)
	parser.Run()
}
```

[0]: https://golang.org/ref/spec#Struct_types
