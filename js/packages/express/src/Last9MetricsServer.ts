import { Last9MetricsServerConfig } from './types';
import { getPromClient } from '@last9/cdk-core';

class Last9MetricsServer {
  private programForPromClient;
  private appVersionForPromClient;

  public promClient;

  constructor(config?: Last9MetricsServerConfig) {
    this.promClient = getPromClient(config?.promClientConfig);
  }
}
