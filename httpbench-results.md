# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **141332.22** | 68329.86 | 58302.22 | 61392.13 | 57092.95 | **Gofi** |
| Single Param | `GET /user/gordon` | **112395.78** | 52291.67 | 47728.93 | 54295.24 | 79252.09 | **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **105706.25** | 48480.21 | 36796.50 | 42286.80 | 81717.65 | **Gofi** |
| Middleware Chain | `GET /middlewares` | **110025.93** | 50568.93 | 35022.34 | 46435.84 | 80550.87 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **109942.90** | 47509.03 | 42856.23 | 33165.00 | 68965.49 | **Gofi** |
| JSON Bind (Small) | `POST /json` | **70413.51** | 44200.96 | 38075.01 | 31825.67 | 53300.48 | **Gofi** |
| JSON Response (Small) | `GET /json-response` | 16081.01 | **34941.64** | 33898.32 | 20676.61 | 33968.22 | **Chi** |
| JSON Bind Only (Large) | `POST /json-large` | **8392.36** | 2970.93 | 2709.71 | 1569.07 | 2279.83 | **Gofi** |
| JSON Response (Large) | `GET /json-response-large` | 3129.07 | **7754.84** | 6896.08 | 3678.61 | 6456.05 | **Chi** |
| JSON Bind + Validate (Small) | `POST /json-validate-small` | **66418.09** | 36722.44 | 34033.99 | 22562.12 | 60081.61 | **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **88274.83** | 43642.14 | 44935.27 | 30525.16 | 78093.73 | **Gofi** |
| JSON Validate Only (Large) | `POST /json-validate-large` | **6217.02** | 1833.66 | 1863.60 | 1329.64 | 1594.23 | **Gofi** |
| JSON Bind + Validate (Large) | `POST /json-bind-validate-large` | **3992.22** | 1802.60 | 1689.13 | 1475.38 | 1366.42 | **Gofi** |
| JSON Validation Failure (Early) | `POST /json-bind-validate-large` | **17388.21** | 2471.84 | 2186.93 | 1581.07 | 1736.78 | **Gofi** |
| JSON Validation Failure (Late/Deep) | `POST /json-bind-validate-large` | **3969.21** | 1539.31 | 1479.64 | 1587.38 | 1002.82 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | 2173.98 | **6810.10** | 6429.45 | 5180.94 | 4571.70 | **Chi** |
| Multipart Bind | `POST /multipart` | **49733.53** | 24716.94 | 25160.94 | 15925.07 | 20612.76 | **Gofi** |
| FormData Bind | `POST /formdata` | **84185.71** | 32410.88 | 36514.42 | 22264.71 | 37633.94 | **Gofi** |
