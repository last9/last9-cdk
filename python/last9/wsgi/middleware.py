import os
import sys
import time
import socket
import inspect

from flask import Flask, request, current_app
from werkzeug.wrappers import Request, Response
from timeit import default_timer

from prometheus_client import Counter, Histogram
from prometheus_client import generate_latest


def get_program():
    return os.path.basename(sys.argv[0])


def get_hostname():
    return socket.gethostname()


http_requests_duration = Histogram(
        "http_requests_duration_milliseconds", "HTTP requests duration per path",
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
    try:
        request.environ["last9_route"] = request.url_rule.rule
    except AttributeError:
        request.environ["last9_route"] = "/unmatched"


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

        return response, (end-start)

    return inner


class WsgiMiddleware(object):
    def __init__(self, app, grouper=None, hostname=None, program=None):
        self.app = app
        self.grouper = grouper or per_path
        self.hostname = hostname or get_hostname()
        self.program = program or get_program()

    def metric_labels(self, req, response):
        return {
                "per": self.grouper(req),
                "hostname": self.hostname,
                "program": self.program,
                "status": response.status_code,
                "domain": req.environ.get("HTTP_HOST"),
                "method": req.environ.get("REQUEST_METHOD")
                }

    def __call__(self, environ, start_response):
        req = Request(environ)

        # decorate the self.app with a timed_function
        # the assumption that response may be null needs to be handled.
        fn = timed_function(self.app)

        response, duration = fn(environ, start_response)

        # Observe the metrics
        labels = self.metric_labels(req, Response(environ))
        # Observe the latency in a bucket
        http_requests_duration.labels(**labels).observe(duration*1000)

        return response


class FlaskMiddleware(WsgiMiddleware):
    def serve_metrics(self):
        return generate_latest()

    def __init__(self, app, grouper=None, hostname=None, program=None):
        self.app = app.wsgi_app
        self.grouper = grouper or per_rule
        self.hostname = hostname or get_hostname()
        self.program = program or get_program()

        # bind events on app for the lifecycle of a request.
        app.teardown_request(teardown_request_func)

        # Bind the route to serve metrics
        app.route("/metrics")(self.serve_metrics)


def per_rule(req):
    return (
            req.environ.get("last9_route") or ""
            ).replace("<", "{").replace(">", "}")


def per_path(req):
    return req.environ.get("PATH_INFO")


class RedMiddleware(object):
    """
    RED Middleware
    """

    def __new__(self, app, grouper=None, *a, **kw):
        if isinstance(app, Flask):
            return FlaskMiddleware(app, grouper,
                                   hostname=kw.get("hostname"),
                                   program=kw.get("program"))

        elif callable(app):
            return WsgiMiddleware(app, grouper,
                                  hostname=kw.get("hostname"),
                                  program=kw.get("program"))

        raise Exception("Unsupported type")
