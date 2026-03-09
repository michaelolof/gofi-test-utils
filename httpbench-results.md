# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **157275.02** | 68164.76 | 4811.59 | 30213.34 | 54742.48 | **Gofi** |
| Single Param | `GET /user/gordon` | **132292.16** | 51171.90 | 12145.52 | 31805.63 | 59193.86 | **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **136188.74** | 43795.45 | 12027.37 | 27545.77 | 56479.41 | **Gofi** |
| Middleware Chain | `GET /middlewares` | **139916.51** | 40774.42 | 12217.49 | 27574.39 | 56273.13 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **133824.12** | 33785.24 | 11141.20 | 27722.92 | 53176.02 | **Gofi** |
| JSON Bind (Small) | `POST /json` | **79233.99** | 26649.38 | 10841.51 | 24198.73 | 44882.09 | **Gofi** |
| JSON Response (Small) | `GET /json-response` | **121542.11** | 19335.51 | 9717.33 | 18723.23 | 26957.55 | **Gofi** |
| JSON Bind (Large) | `POST /json-large` | **2694.44** | 1495.07 | 839.91 | 1798.06 | 2000.28 | **Gofi** |
| JSON Response (Large) | `GET /json-response-large` | **116173.53** | 3354.56 | 2215.66 | 4599.58 | 4390.24 | **Gofi** |
| JSON Validate (Small) | `POST /json-validate-small` | **83332.77** | 14785.48 | 10322.01 | 22768.03 | 37093.43 | **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **109016.55** | 14783.75 | 15789.51 | 28232.69 | 46891.62 | **Gofi** |
| JSON Validate (Large) | `POST /json-validate-large` | **6686.59** | 858.20 | 939.82 | 1221.53 | 1982.40 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **99401.37** | 2064.39 | 3138.28 | 4064.56 | 4971.35 | **Gofi** |
| Multipart Bind | `POST /multipart` | **58624.67** | 6823.53 | 11606.38 | 16998.27 | 23529.04 | **Gofi** |
| FormData Bind | `POST /formdata` | **98024.90** | 9628.87 | 17980.85 | 23079.86 | 32214.72 | **Gofi** |
