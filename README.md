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

## Disclaimer 

The uses of this module are under discovery. There's no stability promise. This means that interfaces could change between commits. No other versioning is intended at this moment. 

## Motivation

Software developers often work with similar patterns over and over again. This isn't exclusive regarding production code, but also testing
patterns and tools. 

Go promotes the creation of small and cohesive tools. This repo is a space for that kind of small, reusable packages
across projects.

Its good to mention, that sometimes its preferable to copy some code than depending in third party libs. Feel free to just use this repo as an example.

## How to use this library

In order to use any of the packages of this Go module, use the following import url:

```
go get go.eloylp.dev/kit
```

## Table of contents

- [Go kit](#go-kit)
	- [Introduction](#introduction)
	- [Disclaimer](#disclaimer)
	- [Motivation](#motivation)
	- [How to use this library](#how-to-use-this-library)
	- [Table of contents](#table-of-contents)
		- [Archive tools](#archive-tools)
		- [Parallelization helpers](#parallelization-helpers)
		- [Data Fanout](#data-fanout)
		- [HTTP Middlewares](#http-middlewares)
			- [Auth](#auth)
			- [Logger](#logger)
			- [Metrics](#metrics)
			- [Default headers](#default-headers)
			- [Panic handling](#panic-handling)

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

### Parallelization helpers

Some times its needed to parallelize a task during a certain time and gracefully wait until all the tasks are done.

Imagine that we have a service called `MultiAPIService`, which has 2 APIs. An `HTTP` one and a `TCP` one. We want to do a stress test on data-fanoutthis service with the [Go race detector](https://go.dev/blog/race-detector) enabled, in order to catch some nasty data races. Lets see a code example on how we can stress such APIs:

```go
package main_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.eloylp.dev/kit/exec"
)

func TestMain(t *testing.T) {

	// Service initialization logic ...

	const MAX_CONCURRENT = 10 // Max concurrent jobs per API.

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // Test time.
	defer cancel()

	wg := &sync.WaitGroup{}

	// Launch all the parallelization for both APIs with exec.Parallelize
	go exec.Parallelize(ctx, wg, MAX_CONCURRENT, doHTTPRequest)
	exec.Parallelize(ctx, wg, MAX_CONCURRENT, doTCPRequest)

	wg.Wait() // Waiting for all tasks to end.

	// Service shutdown logic ...
}

func doHTTPRequest() {
	time.Sleep(100 * time.Millisecond)
}

func doTCPRequest() {
	time.Sleep(50 * time.Millisecond)
}
```

### Data Fanout

Current implementation of Go channels does not allow
to broadcast same values to all consumers. This fanout solution comes to rescue:

```go
package main

import (
	"fmt"
	"io"
	"sync"

	"go.eloylp.dev/kit/flow"
)

func main() {

	// A fanout of integers with buffer size 10.
	fanout := flow.NewFanout[int](10)

	var wg sync.WaitGroup

	// Bring up consumer 1
	wg.Add(1)
	consume1, cancel1 := fanout.Subscribe()
	go func() {
		defer cancel1()
		for {
			elem, err := consume1()
			if err == io.EOF {
				break
			}

			fmt.Printf("consumer 1, received: %v\n", elem.Elem)
		}
		wg.Done()
	}()

	// Same way, bring up consumer 2
	wg.Add(1)
	consume2, cancel2 := fanout.Subscribe()

	go func() {
		defer cancel2()
		for {
			elem, err := consume2()
			if err == io.EOF {
				break
			}

			fmt.Printf("consumer 2, received: %v\n", elem.Elem)
		}
		wg.Done()
	}()

	// Push some elements to the fanout.
	fanout.Publish(1)
	fanout.Publish(2)
	fanout.Publish(3)

	// Shutdown consumers
	cancel1()
	cancel2()

	// Wait for everything
	wg.Wait()
}
```

Result:

```bash
consumer 2, received: 1
consumer 2, received: 2
consumer 2, received: 3
consumer 1, received: 1
consumer 1, received: 2
consumer 1, received: 3
```

If the buffer size of an specific consumer its exceeded, the oldest element will be discarded. This can cause slow consumers to loose data. This could be configurable in the future, though.

See internal code documentation for complete API and other details.

### HTTP Middlewares

HTTP middlewares allow us to execute common logic before all our handlers,
providing to all of them the same pre-processing/post-processing logic.

All this middlewares respect the standard library interfaces, so it should 
not be a problem to use them with your favorite's HTTP lib also.
#### Auth

The [basic Auth](https://www.rfc-editor.org/rfc/rfc7617) middleware provides a way 
to authenticate a specific set of `paths` and `methods` with a given `auth configuration`. 
It also allows multiple sets of configurations. Lets do a walk-through with a full example.

```go
package main

import (
	"net/http"

	"go.eloylp.dev/kit/http/middleware"
)

func main() {

	// Setup the configuration function, that will return all
	// configurations. For this basic case we only need one.
	cfg := middleware.NewAuthConfig().
		WithPathRegex("^/protected.*").
		WithMethods(middleware.AllMethods()).
		// The following hash can be generated like: https://unix.stackexchange.com/questions/307994/compute-bcrypt-hash-from-command-line
		WithAuth(middleware.Authorization{
			"user": "$2y$10$mAx10mlJ/UNbQJCgPp2oLe9n9jViYl9vlT0cYI3Nfop3P3bU1PDay", // unencrypted is user:password.
		})

	cfgs := []*middleware.AuthConfig{cfg}
	cfgFunc := middleware.AuthConfigFunc(func() []*middleware.AuthConfig {
		return cfgs
	})

	// Configure the middleware with already setup config function.
	authMiddleware := middleware.AuthChecker(cfgFunc)

	// Prepare the router. In this example, we will use the standard library.
	// More advanced routers would allow us to define this middleware for
	// all routes globally.
	mux := http.NewServeMux()
	mux.Handle("/", handler())
	mux.Handle("/protected", middleware.For(handler(), authMiddleware))

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != http.ErrServerClosed {
		panic(err)
	}
	// If we visit 0.0.0.0:8080/ we will see "Hello !"
	// 
	// If we visit 0.0.0.0:8080/protected (or any subpath of it) without setting 
	// the proper auth heather, we will see "Unauthorized".
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello !"))

	}
}
```
Please visit code documentation for more clarification about each specific type/helper.

#### Logger

The logger middleware will print general request information in the standard output. It
currently assumes the use of the `logrus` logger.

```go
package main

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"go.eloylp.dev/kit/http/middleware"
)

func main() {

	logger := logrus.NewEntry(logrus.StandardLogger())
	mid := middleware.RequestLogger(logger, logrus.InfoLevel)
	handler := middleware.For(handler(), mid)

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != http.ErrServerClosed {
		panic(err)
	}
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello !"))
	}
}
```

Heres the output in a terminal:

```bash
INFO[0001] intercepted request                           duration="15.645µs" headers="map[Accept:[text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8] Accept-Encoding:[gzip, deflate, br] Accept-Language:[en-GB,en] Cache-Control:[max-age=0] Connection:[keep-alive] Cookie:[redirect_to=%2F; Goland-976d74e5=8331d2d1-9f8e-48d8-a86a-6586446a99e0; Goland-976d74e6=81adf854-3ffb-4839-aa60-e7dbb375c58c] Sec-Ch-Ua:[\"Not?A_Brand\";v=\"8\", \"Chromium\";v=\"108\", \"Brave\";v=\"108\"] Sec-Ch-Ua-Mobile:[?0] Sec-Ch-Ua-Platform:[\"Linux\"] Sec-Fetch-Dest:[document] Sec-Fetch-Mode:[navigate] Sec-Fetch-Site:[none] Sec-Fetch-User:[?1] Sec-Gpc:[1] Upgrade-Insecure-Requests:[1] User-Agent:[Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36]]" ip="[::1]:51110" method=GET path=/ response_size=7
```

#### Metrics

The metrics middlewares allows instrumenting an application really fast 
with [Prometheus](https://prometheus.io) metrics !

Lets see a working example:

```go
package main

import (
	"net/http"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"go.eloylp.dev/kit/http/middleware"
)

// We need an endpoint mapper implementation to avoid cardinality problems in metrics.
// This is just a naive implementation for the example.

// This type should wrap an efficient implementation, like https://github.com/hashicorp/go-immutable-radix
type EndpointMapper struct {
	productIDPathReg *regexp.Regexp
}

func NewEndpointMapper() *EndpointMapper {
	return &EndpointMapper{
		productIDPathReg: regexp.MustCompile(`^/product/\d+$`),
	}
}

func (m *EndpointMapper) Map(url string) string {
	if m.productIDPathReg.MatchString(url) {
		return "/product/{ID}"
	}
	return url
}

func main() {

	// An endpoint mapper its needed.
	endpointMapper := NewEndpointMapper()

	// Configure middlewares with Prometheus defaults.
	durationObserver := middleware.RequestDurationObserver(prometheus.DefaultRegisterer, prometheus.DefBuckets, endpointMapper)
	sizeBuckets := []float64{1, 5, 10, 20, 40, 80} // This is very specific to your application.
	sizeObserver := middleware.ResponseSizeObserver(prometheus.DefaultRegisterer, sizeBuckets, endpointMapper)

	// Apply the middlewares to the product handler.
	productHandler := middleware.For(productHandler(), durationObserver, sizeObserver)

	// Configure the router, enabling metrics.
	mux := http.NewServeMux()
	mux.Handle("/product/1", productHandler) // In an advanced router, placeholders should be available. This is just for the example.
	mux.Handle("/metrics", promhttp.Handler())

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != http.ErrServerClosed {
		panic(err)
	}

	// Reach http://0.0.0.0:8080/product/1 , multiple times.
	// Then Reach http://0.0.0.0:8080/metrics, we should see all metrics there.
	// Note the placeholder {ID} instead of the real product id should be listed.
}

func productHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Heres the product one !"))
	}
}
```

The above example partially results in:

```bash
# HELP http_request_duration_seconds 
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="0.005"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="0.01"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="0.025"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="0.05"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="0.1"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="0.25"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="0.5"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="1"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="2.5"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="5"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="10"} 2
http_request_duration_seconds_bucket{code="200",endpoint="/product/{ID}",method="GET",le="+Inf"} 2
http_request_duration_seconds_sum{code="200",endpoint="/product/{ID}",method="GET"} 9.969499999999999e-05
http_request_duration_seconds_count{code="200",endpoint="/product/{ID}",method="GET"} 2
# HELP http_response_size 
# TYPE http_response_size histogram
http_response_size_bucket{code="200",endpoint="/product/{ID}",method="GET",le="1"} 0
http_response_size_bucket{code="200",endpoint="/product/{ID}",method="GET",le="5"} 0
http_response_size_bucket{code="200",endpoint="/product/{ID}",method="GET",le="10"} 0
http_response_size_bucket{code="200",endpoint="/product/{ID}",method="GET",le="20"} 0
http_response_size_bucket{code="200",endpoint="/product/{ID}",method="GET",le="40"} 2
http_response_size_bucket{code="200",endpoint="/product/{ID}",method="GET",le="80"} 2
http_response_size_bucket{code="200",endpoint="/product/{ID}",method="GET",le="+Inf"} 2
http_response_size_sum{code="200",endpoint="/product/{ID}",method="GET"} 46
http_response_size_count{code="200",endpoint="/product/{ID}",method="GET"} 2
```

The introduced placeholder `{ID}` avoids cardinality problems, allowing us to properly aggregate metrics and create awesome [Grafana](https://grafana.com) dashboards.

This 2 middlewares should be enough for a standard HTTP application, as [Prometheus](https://prometheus.io) histograms already provides `*_sum` and `*_count` metrics. So counters and averages calculations are already available.

#### Default headers

This middleware allows users to set a default set of headers that will be added on every response.

The following handlers have always the last responsibility on wether to override the job done by this middleware.

```go
package main

import (
	"net/http"

	"go.eloylp.dev/kit/http/middleware"
)

func main() {

	// Define the default headers
	defaultHeaders := middleware.DefaultHeaders{}
	defaultHeaders.Set("Server", "random-server")

	// Apply the middleware to the handler chain.
	handler := middleware.For(handler(), middleware.ResponseHeaders(defaultHeaders))

	mux := http.NewServeMux()
	mux.Handle("/", handler)

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != http.ErrServerClosed {
		panic(err)
	}

	// Visiting / shoud show us the following headers:
	// Server: random-server
	// X-Custom-Id: 09AF
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Id", "09AF")
		w.Write([]byte("Hello !"))
	}
}
```

#### Panic handling

Its almost mandatory to protect HTTP servers from panics. If not, one handler can cause
an entire service outage.

A common practice is to define a general panic handler, that will act as protection for
the entire system. Later on, downstream code could define their own panic handlers if
they wish to manage each situation differently.

The following middleware should be defined at very first of the handler chain:

```go
package main

import (
	"log"
	"net/http"

	"go.eloylp.dev/kit/http/middleware"
)

func main() {

	// Defining our panic handler function.
	handlerFunc := middleware.PanicHandlerFunc(func(v interface{}) {
		log.Printf("panic detected: %v\n", v)
	})

	// Configuring the handler chain.
	mid := middleware.PanicHandler(handlerFunc)
	handler := middleware.For(handler(), mid)

	mux := http.NewServeMux()
	mux.Handle("/panic", handler)

	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != http.ErrServerClosed {
		panic(err)
	}
}

func handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		panic("I was about to say hello, but i panicked !")
		w.Write([]byte("Hello !"))
	}
}
```

Visiting `/panic` should show us a similar terminal output like:

```bash
$ 2023/01/09 18:10:04 panic detected: I was about to say hello, but i panicked !
```

And allowing further operations to continue, without crashing the entire server.

In a production scenario, we should instrument our handler function function with some kind of alerting.
