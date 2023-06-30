import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import * as promClient from 'prom-client';

const getPackageJson = () => {
  const packageJsonPath = path.join(process.cwd(), 'package.json');
  try {
    const packageJson = fs.readFileSync(packageJsonPath, 'utf-8');
    return JSON.parse(packageJson);
  } catch (error) {
    console.error('Error parsing package.json');
    return null;
  }
};

const getHostIpAddress = () => {
  const networkInterfaces = os.networkInterfaces();

  // Iterate over network interfaces to find a non-internal IPv4 address
  for (const interfaceName in networkInterfaces) {
    const interfaces = networkInterfaces[interfaceName];
    if (interfaces) {
      for (const iface of interfaces) {
        // Skip internal and non-IPv4 addresses
        if (!iface.internal && iface.family === 'IPv4') {
          return iface.address;
        }
      }
    }
  }

  // Return null if no IP address is found
  return null;
};

export interface PromClientDefaultConfig {
  environment?: string;
}

export const getPromClient = async ({ environment }: PromClientDefaultConfig) => {
  const { version, name } = getPackageJson();
  const ip = getHostIpAddress();
  promClient.register.setDefaultLabels({
    program: name,
    version,
    environment,
    host: os.hostname(),
    ip,
  });
  return promClient;
};
