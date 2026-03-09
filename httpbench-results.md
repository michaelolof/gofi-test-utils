# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **174533.56** | 79398.90 | 55374.32 | 59907.01 | 125927.21 | **Gofi** |
| Single Param | `GET /user/gordon` | **150968.12** | 73591.61 | 42351.46 | 53414.28 | 103337.03 | **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **142353.57** | 70630.60 | 38593.72 | 46235.69 | 94202.66 | **Gofi** |
| Middleware Chain | `GET /middlewares` | **144973.10** | 67328.84 | 36221.03 | 45185.79 | 95732.69 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **138379.26** | 59591.50 | 35085.03 | 43764.45 | 92876.34 | **Gofi** |
| JSON Bind (Small) | `POST /json` | **94183.99** | 50703.92 | 31128.73 | 38596.53 | 78355.11 | **Gofi** |
| JSON Response (Small) | `GET /json-response` | **133920.73** | 40036.89 | 27036.95 | 31058.90 | 48116.21 | **Gofi** |
| JSON Bind (Large) | `POST /json-large` | 3204.82 | 3294.55 | 2237.18 | 2795.31 | **3520.00** | **Fiber** |
| JSON Response (Large) | `GET /json-response-large` | **122048.60** | 9028.79 | 6483.00 | 7000.73 | 7919.58 | **Gofi** |
| JSON Validate (Small) | `POST /json-validate-small` | **96417.15** | 42212.09 | 29573.28 | 36384.48 | 70593.35 | **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **116639.62** | 47283.93 | 37081.04 | 47535.10 | 85565.04 | **Gofi** |
| JSON Validate (Large) | `POST /json-validate-large` | **17850.29** | 2821.03 | 2309.14 | 1965.18 | 3555.96 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **121944.20** | 7382.03 | 6678.53 | 7213.75 | 8628.13 | **Gofi** |
| Multipart Bind | `POST /multipart` | **70513.38** | 24294.75 | 20870.79 | 28382.68 | 42762.16 | **Gofi** |
| FormData Bind | `POST /formdata` | **124613.36** | 34513.31 | 31287.95 | 38296.17 | 67915.24 | **Gofi** |
