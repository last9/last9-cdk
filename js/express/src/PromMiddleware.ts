import * as os from "os";
import express from "express";
import ResponseTime from "response-time";
import promClient, {
  CounterConfiguration,
  HistogramConfiguration,
} from "prom-client";

import type { IncomingMessage, ServerResponse } from "http";
import { getHostIpAddress, getPackageJson, getParsedPathname } from "./utils";

export interface PromMiddlewareOptions {
  /** Route where the metrics will be exposed */
  path: string;
  /** Metrics server, By default the middleware uses existing Express app for the metrics route.
   * This option helps to run the metrics route on a different server running on different port
   */
  metricsServerPort?: number;
  /** Application environment
   * @default 'production'
   */
  environment?: string;
  /** Any default labels you want to include */
  defaultLabels?: Record<string, string>;
  /** Accepts configuration for Prometheus Counter  */
  requestsCounterConfig?: CounterConfiguration<string>;
  /** Accepts configuration for Prometheus Histogram */
  requestDurationHistogramConfig?: HistogramConfiguration<string>;
}

const packageJson = getPackageJson();

const promMiddleware = (options: PromMiddlewareOptions) => {
  // Options with the default set for the optional keys
  const {
    path,
    environment = "production",
    metricsServerPort = 9097,
    defaultLabels,
    requestsCounterConfig = {
      name: "http_requests_total",
      help: "Total number of requests",
      labelNames: ["path", "method", "status"],
    },
    requestDurationHistogramConfig = {
      name: "http_requests_duration_milliseconds",
      help: "Duration of HTTP requests in milliseconds",
      labelNames: ["path", "method", "status"],
      buckets: promClient.exponentialBuckets(0.25, 1.5, 31),
    },
  } = options;

  promClient.register.setDefaultLabels({
    environment,
    program: packageJson.name,
    version: packageJson.version,
    host: os.hostname(),
    ip: getHostIpAddress(),
    ...defaultLabels,
  });

  const collectDefaultMetrics = promClient.collectDefaultMetrics;
  collectDefaultMetrics({
    gcDurationBuckets: requestDurationHistogramConfig.buckets,
  });

  // Prometheus counter for the number of requests
  const requestsCounter = new promClient.Counter(requestsCounterConfig);

  // Prometheus histogram for request duration
  const requestsDurationHistogram = new promClient.Histogram(
    requestDurationHistogramConfig
  );

  // RED Middleware
  const redMiddleware = ResponseTime(
    (
      req: IncomingMessage,
      res: ServerResponse<IncomingMessage>,
      time: number
    ) => {
      if (path !== req.url) {
        const parsedPathname = getParsedPathname(req.url ?? "/", undefined);
        const labels = {
          path: parsedPathname,
          status: res.statusCode.toString(),
          method: req.method as string,
        };

        requestsCounter.labels(labels.path, labels.method, labels.status).inc();
        requestsDurationHistogram
          .labels(labels.path, labels.method, labels.status)
          .observe(time);
      }
    }
  );

  const metricsServer = express();
  // Use the RED middleware
  metricsServer.use(redMiddleware);
  metricsServer.get(path, async (req, res, next) => {
    // Adding Content-Type header
    res.set("Content-Type", promClient.register.contentType);
    return res.end(await promClient.register.metrics());
  });

  metricsServer.listen(metricsServerPort);

  return metricsServer;
};

export default promMiddleware;
