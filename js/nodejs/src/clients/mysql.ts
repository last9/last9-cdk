// Default metrics
// Transaction count metrics
// Transaction Duration metric
import prom from 'prom-client';
import type { createConnection, QueryFunction } from 'mysql';

const instrumentMysql = (mysqlInstance: {
  createConnection: typeof createConnection;
}) => {
  mysqlInstance.createConnection = new Proxy(mysqlInstance.createConnection, {
    apply: (target, prop, receiver) => {
      const connection = Reflect.apply(target, prop, receiver);
      return new Proxy(connection, {
        get: (target, prop, receiver) => {
          if (prop === 'query') {
            return function (...args: Parameters<QueryFunction>) {
              if (typeof args[0] === 'string') {
                console.log(args[0]);
              } else if (typeof args[0].sql === 'string') {
                console.log(args[0].sql);
              }
              // @ts-ignore
              return Reflect.apply(target[prop], this, args);
            } as QueryFunction;
          }
          return Reflect.get(target, prop, receiver);
        }
      });
    }
  });
};

export default instrumentMysql;
