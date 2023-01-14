[![CodeQL](https://github.com/slashdevops/go-table/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/slashdevops/go-table/actions/workflows/codeql-analysis.yml)
[![Run Gosec](https://github.com/slashdevops/go-table/actions/workflows/gosec.yml/badge.svg)](https://github.com/slashdevops/go-table/actions/workflows/gosec.yml)
[![Unit Test](https://github.com/slashdevops/go-table/actions/workflows/main.yml/badge.svg)](https://github.com/slashdevops/go-table/actions/workflows/main.yml)
[![Release](https://github.com/slashdevops/go-table/actions/workflows/release.yml/badge.svg)](https://github.com/slashdevops/go-table/actions/workflows/release.yml)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/slashdevops/go-table?style=plastic)
[![license](https://img.shields.io/github/license/slashdevops/go-table.svg)](https://github.com/slashdevops/go-table/blob/main/LICENSE)
[![codecov](https://codecov.io/gh/slashdevops/go-table/branch/main/graph/badge.svg?token=UNTP5C1P6C)](https://codecov.io/gh/slashdevops/go-table)
[![release](https://img.shields.io/github/release/slashdevops/go-table/all.svg)](https://github.com/slashdevops/go-table/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/slashdevops/go-table.svg)](https://pkg.go.dev/github.com/slashdevops/go-table)

**go-table** is a simple `Table` implementation in golang.

This is focused on `usability and simplicity`.

## Overview

Taking advantage of the [text/tabwriter](https://pkg.go.dev/text/tabwriter) package, this library provides a simple way to create and manage tables.

### Documentation

Official documentation is available on [pkg.go.dev -> slashdevops/go-table](https://pkg.go.dev/github.com/slashdevops/go-table)

## Installing

Latest release:

```bash
go get -u github.com/slashdevops/go-table@latest
```

Specific release:

```bash
go get -u github.com/slashdevops/go-table@vx.y.z
```

Adding it to your project:

```go
import "github.com/slashdevops/go-table"
```

## Examples

EXAMPLE 1: This example use the `Builder` to create a table with a header and rows before building it.  Here you can see how to mix the `Builder` with the `Table` native methods.  Also is used the `Debug` flag to show the table with the `Debug` format (| symbol) at the end of each row.

```go
package main

import (
  "fmt"

  "github.com/slashdevops/go-table"
)

func main() {
  out := new(bytes.Buffer)
  t := NewBuilder(out).
   WithFlags(Debug)

   // add header and rows before building
  t.WithHeader([]string{"Name", "Age", "City"})
  t.WithRow([]string{"John", "30", "New York"}...)
  t.WithRow([]string{"Jane", "20", "London"}...)
  t.WithRow([]string{"Jack", "40", "Paris"}...)
  t.WithRow([]string{"Christian", "47", "Barcelona"}...)

  // build the table
  table := t.Build()

  // add a new row after building
  table.AddRow([]string{"Ely", "50", "Barcelona"}...)

  // render the table
  table.Render()

  fmt.Println(out.String())
}
```

Output:

```bash
Name       |Age  |City
John       |30   |New York
Jane       |20   |London
Jack       |40   |Paris
Christian  |47   |Barcelona
Ely        |50   |Barcelona
```

EXAMPLE 2: Show the same table with a different separator (,) CSV style

```go
package main

import (
  "fmt"

  "github.com/slashdevops/go-table"
)

func main() {
  buf := new(bytes.Buffer)
  t := New(buf, WithSep(","))
  t.SetHeader([]string{"Name", "Age", "City"})
  t.AddRow([]string{"John", "30", "New York"}...)
  t.AddRow([]string{"Jane", "20", "London"}...)
  t.AddRow([]string{"Jack", "40", "Paris"}...)

  // render the table into the buffer
  t.Render()

 fmt.Println(buf.String())
}
```

Output:

```bash
Name,Age,City
John,30,New York
Jane,20,London
Jack,40,Paris
```

## License

go-table is released under the BSD 3-Clause License. See the bundled LICENSE file for details.

* [The 3-Clause BSD License](https://opensource.org/licenses/BSD-3-Clause)
