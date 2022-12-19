# Go kit

<p align="center">
<img src="art/building-blocks.png" alt="go-kit" width="300"/>
</p>

<p align="right" style="color:silver">
Gophers by the great site gopherize.me
</p>

## Introduction

Welcome gophers !

This is a higher level, personal standard lib. A little space in the Git universe where its intended
the [unix philosophy](https://en.wikipedia.org/wiki/Unix_philosophy) to take place.

### Disclaimer 

The uses of this module are under discovery. There's no stability promise. This means that interfaces could change between commits. No other versioning is intended at this moment. 

## Motivation

Software developers often work with similar patterns over and over again. This isn't exclusive regarding production code, but also testing
patterns and tools. Go promotes the creation of small and cohesive tools. This repo is a space for that kind of small, reusable packages
across projects.

## How to use this library

In order to use any of the packages of this Go module, use the following import url:

```
go get go.eloylp.dev/kit
```

## Table of contents

1. [Archive tools](#archive-tools)


### Archive tools

We can create `tar.gz` file given any number of paths, which can be files or directories:

```go
package main

import (
	"fmt"

	"go.eloylp.dev/kit/archive"
)

func main() {
	b, err := archive.TARGZ("/tmp/test.tar.gz", "/home/user/Pictures", "/home/user/notes.txt")
	if err != nil {
		panic(fmt.Errorf("error creating tar: %v", err))
	}
	fmt.Printf("tar created. %v bytes of content were written\n", b)
}
```

The above feels familiar to how we would use the original `tar` command.

In order to decompress the `tar.gz` file we can just use the extraction function:

```go
package main

import (
	"fmt"

	"go.eloylp.dev/kit/archive"
)

func main() {
	b, err := archive.ExtractTARGZ("/home/user/destination", "/tmp/test.tar.gz")
	if err != nil {
		panic(fmt.Errorf("error extracting tar: %v", err))
	}
	fmt.Printf("tar extracted. %v bytes of content were written\n", b)
}
```

Finally, if you need a **stream based** interface, take a look to the `Stream` functions in the same package.

### Contributing

Before a piece of code reaches this repository, this can be a summary of the previous steps:

1. Code is created and tested in applications.

2. A pattern is detected in other project with similar needs.

3. Acceptable maturity of public interfaces is reached.

4. The piece of software and the tests, are promoted to this repo.

** idea : ite preferable to cover many use cases with an abstraction than making it too obtuse. In that case, a low level abstraction and a high level one should be provided.