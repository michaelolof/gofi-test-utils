# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **157530.49** | 79836.47 | 62819.42 | 63284.74 | 124215.59 | **Gofi** |
| Single Param | `GET /user/gordon` | **145365.81** | 63788.99 | 49817.93 | 60210.97 | 100184.32 | **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **134095.88** | 60242.73 | 45659.12 | 52943.40 | 95929.04 | **Gofi** |
| Middleware Chain | `GET /middlewares` | **138207.99** | 59914.77 | 44942.52 | 44947.42 | 97649.51 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **134956.49** | 55574.55 | 42214.90 | 41640.73 | 92097.22 | **Gofi** |
| JSON Bind (Small) | `POST /json` | **82164.74** | 48473.78 | 36825.01 | 42401.31 | 74025.89 | **Gofi** |
| JSON Response (Small) | `GET /json-response` | **125944.75** | 39302.69 | 29620.71 | 32628.27 | 46499.43 | **Gofi** |
| JSON Bind (Large) | `POST /json-large` | 2217.41 | 3118.21 | 2536.14 | 2683.10 | **3560.18** | **Fiber** |
| JSON Response (Large) | `GET /json-response-large` | **119141.03** | 8430.25 | 7119.18 | 6963.54 | 8561.31 | **Gofi** |
| JSON Validate (Small) | `POST /json-validate-small` | **86312.05** | 36687.49 | 35551.72 | 42351.59 | 69329.65 | **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **110221.42** | 40144.96 | 43985.32 | 50780.62 | 79549.77 | **Gofi** |
| JSON Validate (Large) | `POST /json-validate-large` | **7250.95** | 2741.27 | 2833.64 | 2095.14 | 3444.75 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **110811.71** | 7655.24 | 8083.36 | 7374.14 | 8828.94 | **Gofi** |
| Multipart Bind | `POST /multipart` | **63514.33** | 24795.42 | 27109.06 | 27011.26 | 43477.69 | **Gofi** |
| FormData Bind | `POST /formdata` | **111096.34** | 36580.09 | 37893.30 | 37934.41 | 66186.76 | **Gofi** |
