# Australia Post 4-state barcode
This is a minimalistic library written in golang that generates the 4-state barcode used by Australia Post
![Australia Post barcode](./examples/barcode.png)

## Why
The main reason I've written this library is because couldn't find anything that would satisfy the following criterias:
* a small binary that can be run from a lambda function.
* simple to understand code with the minimum functionality and configuration options
* generates barcode as SVG file

4-state-barcode library does only one thing &mdash; generates a 4 state barcode as SVG file.
 
Additionally, it allows to pass an optional text and change a few configuration options.
The guidelines how to generate 4 state barcode required by Australia Post are taken from [customer-barcode-technical-specifications-aug2012.pdf](https://auspost.com.au/.../customer-barcode-technical-specifications-aug2012.pdf) file.

## Usage
The library can used in three different ways.
### As code
```go
package main

import (
    "os"
    "github.com/b-b3rn4rd/4-state-barcode/src/australiapost"
)

func main()  {
    file, _ := os.Create("barcode.svg")
    generator := australiapost.NewFourStateBarcode("5956439111ABA 9", file, "hello world")
        
    err := generator.Generate()
    if err != nil {
        panic(err)
    } 
}
```

There are several optional configuration functions available:
* `OptionPadding(padding int) option`
* `OptionLogger(logger Logger) option`
* `OptionRatio(ratio int) option`
* `OptionFontSize(fontSize int) option`
* `OptionBackgroundColor(color string) option`
* `OptionFontColor(color string) option`
* `OptionalEncoder(encoder Encoder) option`

The following example generates barcode's content into a variable

```go
package main

import (
    "os"
    "bytes"
    "fmt"
    "github.com/b-b3rn4rd/4-state-barcode/src/australiapost"
)

func main()  {
	b := bytes.Buffer{}
    
    generator := australiapost.NewFourStateBarcode("5956439111ABA 9", &b, "hello world", 
    	australiapost.OptionPadding(10),
        australiapost.OptionBackgroundColor("blue"),
        australiapost.OptionFontSize(12),
    )
        
    err := generator.Generate()
    if err != nil {
        panic(err)
    }
    
    s := b.String()
    
    fmt.Println(s)
}
```

## As CLI

```bash
$ 4-state-barcode --help
Australia Post 4 state barcode generator
Generates a SVG image containing barcode with an optional additional text
Example: 
4-state-barcode -b "5956439111ABA 9" -f barcode.svg

Usage:
  4-state-barcode [flags]

Flags:
  -b, --barcode string    Barcode value
  -f, --filename string   Output filename
  -h, --help              help for 4-state-barcode
  -t, --text string       Optional barcode text
      --version           version for 4-state-barcode

```