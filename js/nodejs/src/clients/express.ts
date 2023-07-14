import type { Router } from 'express';

const instrumentExpress = (module: { Router: Router }) => {
  module.Router['use'] = new Proxy(module.Router['use'], {
    apply: (target, props, receiver) => {
      const result = Reflect.apply(target, props, receiver);
      const newResult = {
        ...result,
        stack: [...result.stack]
      };

      console.log(newResult);
      return newResult;
    }
  });
};

export default instrumentExpress;
