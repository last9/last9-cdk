import type { PromClientDefaultConfig } from '@last9/cdk-core';

export interface Last9MetricsServerConfig {
  promClientConfig: PromClientDefaultConfig & {
    port: number;
  };
}
