import * as os from 'os';
import http from 'http';
import prom from 'prom-client';

import type {
  Counter,
  CounterConfiguration,
  Histogram,
  HistogramConfiguration
} from 'prom-client';
import type { Server } from 'http';

import { getHostIpAddress, getPackageJson, getSanitizedPath } from './utils';
import instrumentExpress from './clients/express';

export interface CDKOptions {
  /** Route where the metrics will be exposed
   * @default "/metrics"
   */
  path?: string;
  /** Port for the metrics server
   * @default 9097
   */
  metricsServerPort?: number;
  /** Application environment
   * @default 'production'
   */
  environment?: string;
  /** Any default labels you want to include */
  defaultLabels?: Record<string, string>;
}

const packageJson = getPackageJson();

export class CDK {
  private path: string;
  private metricsServerPort: number;
  private environment: string;
  private defaultLabels?: Record<string, string>;

  private requestsCounter?: Counter;
  private requestsDurationHistogram?: Histogram;
  public metricsServer?: Server;

  constructor(options?: CDKOptions) {
    // Initializing all the options
    this.path = options?.path ?? '/metrics';
    this.metricsServerPort = options?.metricsServerPort ?? 9097;
    this.environment = options?.environment ?? 'production';
    this.defaultLabels = options?.defaultLabels;

    this.initiateMetricsRoute();
    this.initiatePromClient();
  }

  private initiatePromClient = () => {
    // Setting default Labels
    prom.register.setDefaultLabels({
      environment: this.environment,
      program: packageJson.name,
      version: packageJson.version,
      host: os.hostname(),
      ip: getHostIpAddress(),
      ...this.defaultLabels
    });
  };

  private initiateMetricsRoute = () => {
    // Creating native http server
    this.metricsServer = http.createServer(async (req, res) => {
      // Sanitize the path
      const path = getSanitizedPath(req.url ?? '/');
      if (path === this.path && req.method === 'GET') {
        res.setHeader('Content-Type', prom.register.contentType);
        return res.end(await prom.register.metrics());
      } else {
        res.statusCode = 404;
        res.end('404 Not found');
      }
    });

    // Start listening at the given port defaults to 9097
    this.metricsServer.listen(this.metricsServerPort, () => {
      console.log(`Metrics server running at ${this.metricsServerPort}`);
    });
  };

  // Function overloads for all supported modules
  public instrument<E>(moduleName: 'express', express: E): void;
  public instrument<MS>(moduleName: 'mysql', mysql: MS): void;
  public instrument(moduleName: string, module: any): void {
    if (moduleName === 'express') {
      instrumentExpress(module);
    }
    return;
  }
}

export default CDK;
