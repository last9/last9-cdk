import ResponseTime from 'response-time';
import prom from 'prom-client';
import { getParsedPathname, getSanitizedPath } from '../utils';

import type { Router, Application, Request } from 'express';
import type { CounterConfiguration, HistogramConfiguration } from 'prom-client';
import type { IncomingMessage } from 'http';

interface InstrumentExpressOptions {
  requestsCounterConfig?: CounterConfiguration<string>;
  requestDurationHistogramConfig?: HistogramConfiguration<string>;
}

const instrumentExpress = (
  expressInstance: {
    Router: Router;
    application: Application;
  },
  options?: InstrumentExpressOptions
) => {
  /////////// Prometheus configuration
  const requestsCounter = new prom.Counter(
    options?.requestsCounterConfig ?? {
      name: 'http_requests_total',
      help: 'Total number of requests',
      labelNames: ['path', 'method', 'status']
    }
  );
  const defaultBuckets = prom.exponentialBuckets(0.25, 1.5, 31);

  const requestsDurationHistogram = new prom.Histogram(
    options?.requestDurationHistogramConfig ?? {
      name: 'http_requests_duration_milliseconds',
      help: 'Duration of HTTP requests in milliseconds',
      labelNames: ['path', 'method', 'status'],
      buckets: defaultBuckets
    }
  );

  prom.collectDefaultMetrics({
    gcDurationBuckets: defaultBuckets
  });
  ////////////////////////////////////////

  ///////// REDmiddleware
  const REDmiddleware = ResponseTime(
    (req: IncomingMessage & Request, res, time) => {
      const sanitizePathname = getSanitizedPath(req.originalUrl ?? '/');
      const parsedPathname = getParsedPathname(sanitizePathname, undefined);
      const labels = {
        path: parsedPathname,
        status: res.statusCode.toString(),
        method: req.method as string
      };

      requestsCounter?.labels(labels.path, labels.method, labels.status).inc();
      requestsDurationHistogram
        ?.labels(labels.path, labels.method, labels.status)
        .observe(time);
    }
  );
  //////////////////////////

  // Add the middleware to the application
  let hasMiddlewareMounted = false;
  type AppUseType = typeof expressInstance.application.use;
  const originalUse = expressInstance.application.use;

  // Overiding the use function in application and adding the express middleware
  expressInstance.application.use = function use() {
    // Avoid mounting the middleware twice
    if (!hasMiddlewareMounted) {
      // @ts-ignore
      originalUse.apply(this, [REDmiddleware]);
      hasMiddlewareMounted = true;
    }
    // @ts-ignore
    originalUse.apply(this, arguments);
  } as AppUseType;

  // Add the middleware to the application
  // type RouterUseType = typeof expressInstance.Router.use;
  // const originalRouterUse = expressInstance.Router.use;

  // // Overiding the use function in router and adding the express middleware
  // expressInstance.Router.use = function use() {
  //   // Avoid mounting the middleware twice
  //   if (!hasMiddlewareMounted) {
  //     // @ts-ignore
  //     originalRouterUse.apply(this, [REDmiddleware]);
  //     hasMiddlewareMounted = true;
  //   }
  //   // @ts-ignore
  //   originalRouterUse.apply(this, arguments);
  // } as RouterUseType;
};

export default instrumentExpress;
