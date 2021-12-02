### Last9 Python CDK

Last9 Python CDK is a way to emit commonly aggregated metrics straight from your python application that serves HTTP traffic.

### Why

Traditional metrics emitters emit, per-path or handler-based metrics. Both these approaches have their drawbacks.

- Per-path-based metrics can explode the cardinality because `/user/foo` `/user/bar` and so on can result in millions of endpoints
- handler-based metrics can absorb details of the route details.

### What metrics does CDK emit?

#### RED Metrics

The very first metrics to be emitted are Rate, Error, and Duration. The following two metrics are enough to populate that.

```
http_requests_duration_milliseconds = Histogram(
	"http_requests_duration", "HTTP requests duration per path",
  ["per", "hostname", "domain", "method", "program", "status",
  "tenant", "cluster"]
)
```

### How are metrics exposed?

CDK emits metrics in ~~Openmetrics~~ Prometheus Exposition format. The metric port and endpoint expose the freshest metrics to be pulled by a Prometheus.

You can read more about the Prometheus exposition format on the [link](https://github.com/Showmax/prometheus-docs/blob/master/content/docs/instrumenting/exposition_formats.md)

### Python Support

The most common frameworks in Python are WSGI-based. While WSGI-based implementation can have a generic enough way to capture duration, status_code, method, domain name, etc.
It cannot produce the path pattern used for the handler invocation, that information only resides with the Mux involved.

Hence, there may be parts to the CDK which are activated ONLY when supported frameworks are detected or declared.


### Example Usage

1. ```pip install last9```
2. Sample flask application using last9

```import time
from flask import Flask

# Import RedMiddleware from last9
from last9.wsgi.middleware import RedMiddleware

app = Flask(__name__)

# Only line that needs to be changed
app.wsgi_app = RedMiddleware(app)

@app.route("/name/<name>")
def hello_world(name):
    time.sleep(1)
    return "<p>Hello, %s!</p>" % name

@app.route("/static")
def hello_static():
    return "<p>Static</p>"

if __name__ == "__main__":
    app.run('127.0.0.1', '5000', debug=True)

```

 3. Generate sample traffic by calling `curl -s -k 'GET' 'http://localhost:5000/name/[1-10]'`
 4. Metrics will be exposed on `/metrics` endpoint.


### Framework Support

| Name   | Supported  | Mux Supported  |
|---|---|---|
| Flask  | Yes   | Yes  |
| Tornado |  - | -  |
| Django  |  - | -  |
| Pyramid  |  - | -  |
| Bottle  |  - | -  |
| FastAPI  |  - | -  |
| Falcon  |  - | -  |

### Next Release Plan

[ ] Lineage detection in environments where a forward proxy like Envoy, etc is absent.

[ ] Concurrency Metrics.
