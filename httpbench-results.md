# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **161218.08** | 78474.83 | 55302.25 | 59022.07 | 116747.33 | **Gofi** |
| Single Param | `GET /user/gordon` | **137274.30** | 67605.17 | 43653.68 | 51926.20 | 93024.04 | **Gofi** |
| Multi Param | `GET /users/123/posts/456` | **137068.24** | 62645.71 | 39202.71 | 45436.23 | 89068.79 | **Gofi** |
| Middleware Chain | `GET /middlewares` | **137237.45** | 62012.75 | 40406.59 | 45252.20 | 89961.78 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **128683.14** | 55324.47 | 37534.90 | 43941.38 | 83514.71 | **Gofi** |
| JSON Bind (Small) | `POST /json` | **85964.96** | 47571.56 | 33761.37 | 37810.41 | 71260.70 | **Gofi** |
| JSON Response (Small) | `GET /json-response` | **121691.71** | 38704.63 | 25933.52 | 29540.61 | 44244.99 | **Gofi** |
| JSON Bind (Large) | `POST /json-large` | 3103.64 | 3042.09 | 2368.47 | 2752.63 | **3423.59** | **Fiber** |
| JSON Response (Large) | `GET /json-response-large` | **112637.53** | 8554.07 | 6532.91 | 6896.01 | 8096.33 | **Gofi** |
| JSON Validate (Small) | `POST /json-validate-small` | **83453.99** | 39219.77 | 31050.75 | 36272.58 | 64080.56 | **Gofi** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | **108917.76** | 45677.19 | 39760.77 | 41689.63 | 78385.25 | **Gofi** |
| JSON Validate (Large) | `POST /json-validate-large` | **16873.80** | 2871.12 | 2411.09 | 1825.23 | 3148.74 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | **111319.64** | 7583.70 | 6797.15 | 6749.77 | 7786.03 | **Gofi** |
| Multipart Bind | `POST /multipart` | **68032.25** | 24755.07 | 22176.17 | 26209.44 | 37690.87 | **Gofi** |
| FormData Bind | `POST /formdata` | **110929.68** | 34862.55 | 33099.65 | 36744.42 | 59271.42 | **Gofi** |
