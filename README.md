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

## Tag ##
The structopt provides severals pre-define tag and use to identify the field:

| Tag      | Value    | Description                                                              |
|----------|----------|--------------------------------------------------------------------------|
| -        |          | Ignore parse the field                                                   |
| name     |          | The customied field name of the field                                    |
| short    |          | The customied rune of the field as a shortcut                            |
| help     |          | The description or help message                                          |
| callback |          | The callback function defined and execute when set value                 |
| choice   |          | Pre-defined value that only can be set in the field (separate by spece)  |
| default  |          | The default value of the field                                           |
|----------|----------|--------------------------------------------------------------------------|
| option   |          | The customied option that used to set the propertied (separate by comma) |
|          | skip     | Same as '-' and skip process the field                                   |
|          | flag     | Force set the field as the flag                                          |
|          | trunc    | The value can be truncated when set, usually set in the INT and UINT     |
|          | required | Force required the field cannot be empty value                           |

[0]: https://golang.org/ref/spec#Struct_types
