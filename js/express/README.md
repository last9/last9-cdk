# @last9/cdk-express-js

## Setup locally

Make sure you are in the express directory

- Install packages

```
npm install
```

- Build package

  - This will build the package and store the JS and type declaration files in
    the `dist` folder.

```
npm run build
```

## Usage

In the example below, the metrics will be served on `localhost:9097/metrics`. To
change the port, you can update it through the options
([See the options documentation](#options)).

```js
const express = require('express)
const { CDK } = require('@last9/cdk-express-js')

const app = express();
const cdk = new CDK();

app.use(cdk.REDMiddleware);

// ...

app.listen(3000)

```

## Options

### Usage

```js
const cdk = new CDK({
  // Options go here
});
```

1. `path`: The path at which the metrics will be served. For ex. `/metrics`
2. `metricsServerPort`: (Optional) The port at which the metricsServer will run.
3. `environment`: (Optional) The application environment. Defaults to
   `production`
4. `defaultLabels`: (Optional) Any default labels to be included.
5. `requestsCounterConfig`: (Optional) Requests counter configuration, same as
   [Counter](https://github.com/siimon/prom-client#counter) in `prom-client`.
   Defaults to
   ```js
   {
      name: 'http_requests_total',
      help: 'Total number of requests',
      labelNames: ['path', 'method', 'status'],
    }
   ```
6. `requestDurationHistogramConfig`: (Optional) Requests Duration histogram
   configuration, same as
   [Histogram](https://github.com/siimon/prom-client#histogram) in
   `prom-client`. Defaults to
   ```js
    {
        name: 'http_requests_duration_milliseconds',
        help: 'Duration of HTTP requests in milliseconds',
        labelNames: ['path', 'method', 'status'],
        buckets: promClient.exponentialBuckets(0.25, 1.5, 31),
      }
   ```
