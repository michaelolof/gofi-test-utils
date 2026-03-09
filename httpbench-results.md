# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **164049.36** | 51392.31 | 28080.83 | 47795.39 | 109097.41 | **Gofi** |
| Single Param | `GET /user/gordon` | **103889.11** | 40731.20 | 21611.24 | 44584.24 | 85414.43 | **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **126116.83** | 40336.15 | 18398.76 | 39286.86 | 77671.25 | **Gofi** |
| Middleware Chain | `GET /middlewares` | **128225.40** | 33169.48 | 16538.30 | 37314.10 | 77501.30 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **126361.62** | 30487.48 | 20287.32 | 35612.75 | 68762.14 | **Gofi** |
| JSON Bind (Small) | `POST /json` | **75164.84** | 32040.78 | 13758.06 | 30205.51 | 58023.06 | **Gofi** |
| JSON Response (Small) | `GET /json-response` | **120821.04** | 26297.93 | 12489.40 | 24360.26 | 36524.41 | **Gofi** |
| JSON Bind (Large) | `POST /json-large` | 2545.44 | 1972.63 | 1240.63 | 2137.63 | **3019.68** | **Fiber** |
| JSON Response (Large) | `GET /json-response-large` | **98193.41** | 4825.36 | 4071.13 | 5670.27 | 6837.08 | **Gofi** |
| JSON Validate (Small) | `POST /json-validate-small` | **77564.03** | 23761.88 | 19306.52 | 32319.30 | 54143.83 | **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **106625.04** | 24806.84 | 27612.75 | 41042.02 | 64104.96 | **Gofi** |
| JSON Validate (Large) | `POST /json-validate-large` | **6648.12** | 1708.91 | 1702.60 | 1682.21 | 2841.15 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **95677.27** | 4033.09 | 4878.62 | 5894.58 | 6902.25 | **Gofi** |
| Multipart Bind | `POST /multipart` | **49474.37** | 14597.69 | 16749.32 | 21083.19 | 29958.92 | **Gofi** |
| FormData Bind | `POST /formdata` | **78102.73** | 20093.20 | 23473.41 | 28930.78 | 48143.06 | **Gofi** |
