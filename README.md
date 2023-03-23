# Last9 CDK for HTTP

Last9 Golang CDK is a way to emit commonly aggregated metrics straight from your Golang application that serves HTTP traffic.

### Why?

Traditional metrics emitters emit, per-path or handler-based metrics. Both these approaches have their drawbacks.

- Per-path-based metrics can explode the cardinality because `/user/foo` `/user/bar` and so on can result in millions of endpoints.
- handler-based metrics can absorb details of the route details.

### What metrics does CDK emit?

### RED Metrics

The very first metrics to be emitted are Rate, Error, and Duration.

```go
httpRequestsDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_requests_duration_milliseconds",
			Help: "HTTP requests duration per path",
		},
		[]string{
			"per", "hostname", "domain", "method", "program", "status",
			"tenant", "cluster",
		},
	)
```

## Labels Explained

| Name | Description |
| --- | --- |
| hostname | current Hostname where the metric is emitted from |
| program | Binary / Process name where the metric is emitted from |
| per | This is the "main" label, which contains the pathname or an identifier that you emit per request.
By default, it is the path pattern if the Mux is one of the supported List and the entire URL Path where Mux does not have a path parameter.
Optionally, this can ALSO be a custom string (Read below for details) |
| domain | domain at which the request was received |
| method | HTTP method |
| status | HTTP Status returned |
| tenant | Optional Field for a multi-tenant application |
| cluster | Optional Field for a multi-cluster deployment |

---

## Say more about Tenancy

Most modern SaaS applications end up being multi-tenant or multi-clustered where it's crucial to identify the behavior across each, separately.

CDK honors this need as a first-class property and has reserved two label fields for this purpose. These two are:

- tenant
- cluster

**Features**

- Simple configuration for multi-tenancy or multi-cluster using an additional label.
- Cross-tenant aggregation or segregation later via PromQL.
- Allow data with no tenant or cluster information to be written or queried.

---

### How are metrics exposed?

CDK emits metrics in ~~Openmetrics~~ Prometheus Exposition format. The metric port and endpoint expose the freshest metrics to be pulled by a Prometheus.

You can read more about the Prometheus exposition format on the [link](https://github.com/Showmax/prometheus-docs/blob/master/content/docs/instrumenting/exposition_formats.md)

### Golang Support

There may be parts to the CDK which are activated ONLY when supported frameworks are detected or declared.

```go
import (
	"github.com/last9/last9-cdk/go/httpmetrics"
)

// given you have a http.Handler or a Router already
// Only line that needs to be changed
httpmetrics.REDHandler(handler)
```

| Name | Supported | Mux Supported |
| --- | --- | --- |
| http.ServeMux | Yes | Yes |
| Gorilla | Yes | Yes |
| Pat | Yes | Yes |
| Chi | Yes | Yes |

---

## An Example Golang Application

go get the httpmetrics

```bash
go get -v github.com/last9/last9-cdk/go/httpmetrics
```

Assuming a basic main.go

```go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/last9/last9-cdk/go/httpmetrics"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {
	http.Handle("/golang", httpmetrics.REDHandler(http.HandlerFunc(handler)))
	go httpmetrics.ServeMetrics(9091)
	log.Fatal(http.ListenAndServe(":9090", nil))
}
```

Once set up, you can send a couple of test requests on your application

```bash
➜  example curl -XPOST http://localhost:9090/golang -d '{}'
Hi there, I love golang!%                                                                                                                                                        ➜  example curl -XGET http://localhost:9090/golang -d '{}' 
Hi there, I love mount!
```

> Visit [localhost:9091/metrics](http://localhost:9091/metrics) to see your RED metrics
> 

---

## Custom Labels

If you want to change the default path being emitted, it is extremely easy to do so.

Say, my application handler is something like this

```go
m := mux.NewRouter()
m.Handle("/api/category/{category}/item/{id}", itemHandler())
m.Use(REDHandler)
```

This will emit metrics where the `per` label will look like `/api/category/{category}/item/{id}`

But you want the category to NOT be abstracted. For situations like these, you can use the 

`REDHandlerWithLabelMaker` function to assist the label-making process.

```go
// REDHandlerWithLabelMaker accepts a function that in-turn accepts
// both the request and the mux.
m.Use(REDHandlerWithLabelMaker(
	func(r *http.Request, m http.Handler) map[string]string {

		// Gorilla exposes the variables using a request local mux
		vars := mux.Vars(r)
		category := vars["category"]

		return map[string]string{
			"per":     strings.Replace(r.URL.Path, "{category}", category),
			"tenant":  "possible_override", // You may also override other labels
		}

	},
))
```

Voila! That's it

> The above example is for Gorilla Mux, but it's extremely straightforward to draw inspiration for other mux like Pat, etc.
> 

---

# About Last9

This project is sponsored and maintained by [Last9](https://last9.io). Last9 builds reliability tools for SRE and DevOps.

<a href="https://last9.io"><img src="https://last9.github.io/assets/email-logo-green.png" alt="" loading="lazy" height="40px" /></a>
