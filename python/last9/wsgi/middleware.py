import time
import inspect
from flask import Flask, request, current_app
from werkzeug.wrappers import Request, Response
from timeit import default_timer

from prometheus_client import Counter, Histogram
from prometheus_client import generate_latest

http_requests_total = Counter(
        "http_requests_total", "Total HTTP requests per path",
        ["per", "hostname", "domain", "method", "program", "status"]
        )

http_requests_duration = Histogram(
        "http_requests_duration", "HTTP requests duration per path",
        ["per", "hostname", "domain", "method", "program", "status"]
        )

def teardown_request_func(error=None):
    """
    If it's a flask request, last9_route MUST point to the matched rule
    and not the original path. This additional trick prevvents
    cardinality explosion of the emitted metrics.

    Each framework has its unique way of handling this, so just like Flask
    if a new framework is being monitored AND we want to emit metrics
    for the matched pattern, a similar trick would have to be performed.

    In absence of this, the emitter will fallback to per-request URI metrics.
    """
    request.environ["last9_route"] = request.url_rule.rule


def timed_function(fn):
    """
    A decorator to return an additional response argument
    i.e the total duration of the function run. Any exceptions
    are snubbed.

    TODO: Maybe exception and duration can be packed
    together as a second response argument.
    """
    def inner(*a, **kw):
        start = default_timer()
        try:
            response = fn(*a, **kw)
        finally:
            end = default_timer()

        return (response, (end-start))

    return inner

class WsgiMiddleware(object):
    def __init__(self, app, grouper=None):
        self.app = app
        self.grouper = grouper or per_path

    def metric_labels(self, request, response):
        return {
                "per": self.grouper(request),
                "hostname": "hostname",
                "program": "my program",
                "status": response.status_code,
                "domain": request.environ.get("HTTP_HOST"),
                "method": request.environ.get("REQUEST_METHOD")
                }

    def __call__(self, environ, start_response):
        request = Request(environ)

        # decorate the self.app with a timed_function
        # the assumption that response may be null needs to be handled.
        fn = timed_function(self.app)

        response, duration = fn(environ, start_response)

        # Observe the metrics
        labels = self.metric_labels(request, Response(environ))
        # Incrrment the request counter
        http_requests_total.labels(**labels).inc()
        # Observe the latency in a bucket
        http_requests_duration.labels(**labels).observe(duration*1000)

        return response


class FlaskMiddleware(WsgiMiddleware):
    def serve_metrics(self):
        return generate_latest()

    def __init__(self, app, grouper=None):
        self.app = app.wsgi_app
        self.grouper = grouper or per_rule
        # bind events on app for the lifecycle of a request.
        app.teardown_request(teardown_request_func)

        # Bind the route to serve metrics
        app.route("/metrics")(self.serve_metrics)

def per_rule(request):
    return (
            request.environ.get("last9_route") or ""
            ).replace("<", "{").replace(">", "}")

def per_path(request):
    return request.environ.get("PATH_INFO")

class RedMiddleware(object):
    '''
    RED Middleware
    '''
    def __new__(self, app, grouper=None, *a, **kw):
        if isinstance(app, Flask):
            return FlaskMiddleware(app, grouper)

        elif callable(app):
            return WsgiMiddleware(app, grouper)

        raise Exception("Unsupported type")

