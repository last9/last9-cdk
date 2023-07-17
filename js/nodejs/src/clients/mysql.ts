// Default metrics
// Transaction count metrics
// Transaction Duration metric
import prom from 'prom-client';
import type { createConnection } from 'mysql';

const instrumentMysql = (mysqlInstance: {
  createConnection: typeof createConnection;
}) => {
  mysqlInstance.createConnection = new Proxy(mysqlInstance.createConnection, {
    apply: (target, props, args) => {
      const connection = Reflect.apply(target, props, args);
      return new Proxy(connection, {
        get: (target, props, args) => {
          if (props === 'query') {
            console.log(args);
          }
          return Reflect.get(target, props, args);
        }
      });
    }
  });
};

export default instrumentMysql;
