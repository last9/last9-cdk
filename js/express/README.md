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

In the example below, the metrics will be served on `localhost:3000/metrics`. To
serve the metrics on different server use `metricsServer` option
([See the options documentation](#options)).

```js
const express = require('express)
const { promMiddleware } = require('@last9/cdk-express-js')

const app = express();

app.use(promMiddleware({
  path: '/metrics,
}))

app.listen(3000)

```

## Options

1. `path`: The path at which the metrics will be served. For ex. `/metrics`
2. `pathsRegexp`: This is a mandatory option that needs to be passed to the
   promMiddleware middleware. This allows the middleware to replace any
   parameteres in the pathname with a `replacementString`. The `pathsRegexp` is
   a function that returns an array of the regexp for the paths.
   - Example Express v4
   ```js
   const pathsRegexp = () => {
     // Use the express app here to get the regexp
     return app._router.stack.map((router) => router.regexp);
   };
   ```
   - Example Express v3
   ```js
   const pathsRegexp = () => {
     // Use the express app here to get the
     return app.routes.map((router) => router.regexp);
   };
   ```
3. `replacementString`: The replacementString is a filler string for replacing
   the params in the pathname.
4. `metricsServer`: (Optional) A express server which will be exposed on a
   different port than the app. If it is not exposed, the metrics will be served
   at the application port.
5. `environment`: (Optional) The application environment. Defaults to
   `production`
6. `defaultLabels`: (Optional) Any default labels to be included.
7. `requestsCounterConfig`: (Optional) Requests counter configuration, same as
   [Counter](https://github.com/siimon/prom-client#counter) in `prom-client`.
   Defaults to
   ```js
   {
      name: 'http_requests_total',
      help: 'Total number of requests',
      labelNames: ['path', 'method', 'status'],
    }
   ```
8. `requestDurationHistogramConfig`: (Optional) Requests Duration histogram
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
