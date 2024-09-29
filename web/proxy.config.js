module.exports = {
  "/api": {
    "changeOrigin": true,
    "target": "http://localhost:8080",
    "secure": false,
    "pathRewrite": {
      "^/api": ""
    },
    "logLevel": "debug",
    bypass: function(req, res, opts) {

    },
    onProxyRes: (proxyRes, req, res) => {
      proxyRes.headers['x-added'] = 'foobar'; // add new header to response
    }
  }
};
