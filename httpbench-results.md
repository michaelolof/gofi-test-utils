# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | 182131.88 | 177706.98 | 175234.97 | 175758.81 | **182935.51** | **Fiber** |
| Single Param | `GET /user/gordon` | 183237.41 | 175923.25 | 176035.16 | 179415.92 | **183499.07** | **Fiber** |
| Multi Param | `GET /users/123/posts/456` | 183125.89 | 176020.85 | 176024.08 | 176326.91 | **183517.12** | **Fiber** |
| Middleware Chain | `GET /middlewares` | **183677.39** | 177522.46 | 172735.82 | 176519.11 | 182817.48 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **184329.83** | 173712.36 | 173112.45 | 173002.99 | 183220.54 | **Gofi** |
| JSON Bind (Small) | `POST /json` | 180900.63 | 174090.14 | 170566.35 | 172370.15 | **182529.67** | **Fiber** |
| JSON Response (Small) | `GET /json-response` | **180966.01** | 160047.07 | 163378.28 | 146454.19 | 171507.35 | **Gofi** |
| JSON Bind Only (Large) | `POST /json-large` | **42618.77** | 26243.05 | 26202.90 | 25363.80 | 38403.40 | **Gofi** |
| JSON Response (Large) | `GET /json-response-large` | **84071.51** | 61253.39 | 63976.06 | 45612.15 | 57794.97 | **Gofi** |
| JSON Bind + Validate (Small) | `POST /json-validate-small` | **182736.70** | 172762.39 | 171504.30 | 172405.05 | 182728.84 | **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | 183160.20 | 172393.20 | 174860.28 | 174757.00 | **183250.88** | **Fiber** |
| JSON Validate Only (Large) | `POST /json-validate-large` | **37629.88** | 15714.32 | 15510.45 | 15545.23 | 20142.59 | **Gofi** |
| JSON Bind + Validate (Large) | `POST /json-bind-validate-large` | **26044.74** | 15529.13 | 15340.26 | 15499.48 | 19825.00 | **Gofi** |
| JSON Validation Failure (Early) | `POST /json-bind-validate-large` | **84953.71** | 24455.83 | 24206.00 | 15437.90 | 34758.37 | **Gofi** |
| JSON Validation Failure (Late/Deep) | `POST /json-bind-validate-large` | **25889.97** | 15404.25 | 15241.92 | 15447.74 | 19810.71 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **47902.23** | 21903.47 | 21638.38 | 19729.21 | 24240.71 | **Gofi** |
| Multipart Bind | `POST /multipart` | **178818.37** | 135447.76 | 136259.99 | 138347.61 | 173715.65 | **Gofi** |
| FormData Bind | `POST /formdata` | 182676.40 | 166608.60 | 164727.40 | 166477.66 | **182939.97** | **Fiber** |
