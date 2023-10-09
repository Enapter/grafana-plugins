export const setupMirage = () => {
  const { createServer } = require('miragejs');
  const server = createServer({
    namespace: 'api',
    routes() {
      this.post(
        '/ds/query',
        () => {
          return {
            queries: [
              {
                refId: 'A',
                fields: [
                  {
                    state: 'succeeded',
                  },
                  {
                    state: 'error',
                    errors: [
                      {
                        code: 'invalid_header',
                        message: 'Authentication token is invalid.',
                        details: {
                          header: 'X-Enapter-Auth-Token',
                        },
                      },
                    ],
                  },
                  {
                    state: 'platform_error',
                    errors: [
                      {
                        code: 'invalid_header',
                        message: 'Authentication token is invalid.',
                        details: {
                          header: 'X-Enapter-Auth-Token',
                        },
                      },
                    ],
                  },
                ][Math.floor(Math.random() * 3)],
              },
            ],
          };
        },
        { timing: 1500 }
      );

      this.passthrough();
    },
  });

  (window as any).__mirage = server;
  return server;
};
