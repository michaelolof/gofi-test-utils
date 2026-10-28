# HTTP Load Test Results

Tests run with `bombardier` for 5s with 125 concurrent connections.

| Endpoint | Method | Gofi | Chi | Echo | Gin | Fiber | Winner |
|---|---|---|---|---|---|---|---|
| Static Route | `GET /` | **183482.30** | 176701.84 | 175154.15 | 175827.79 | 182662.57 | **Gofi** |
| Single Param | `GET /user/gordon` | **183883.02** | 176055.42 | 176123.81 | 179194.77 | 182661.71 | **Gofi** |
| Multi Param | `GET /users/123/posts/456` | 183031.66 | 175497.80 | 175491.64 | 175773.62 | **183659.83** | **Fiber** |
| Middleware Chain | `GET /middlewares` | **183837.66** | 177806.13 | 175750.74 | 176163.59 | 183015.17 | **Gofi** |
| Query Processing | `GET /query?q=searchterm&limit=10` | **183402.66** | 173711.74 | 173788.03 | 169407.78 | 182200.37 | **Gofi** |
| JSON Bind (Small) | `POST /json` | 181703.38 | 173938.66 | 171072.35 | 171914.02 | **182680.17** | **Fiber** |
| JSON Response (Small) | `GET /json-response` | 86988.86 | 160833.83 | 164403.78 | 148352.05 | **172712.90** | **Fiber** |
| JSON Bind Only (Large) | `POST /json-large` | **43026.26** | 26438.29 | 26457.29 | 25634.28 | 38902.83 | **Gofi** |
| JSON Response (Large) | `GET /json-response-large` | 18827.80 | 62061.17 | **63716.05** | 46334.14 | 62884.56 | **Echo** |
| JSON Bind + Validate (Small) | `POST /json-validate-small` | 182675.22 | 173819.36 | 170556.35 | 172389.71 | **183441.79** | **Fiber** |
| JSON Validate Response (Small) | `GET /json-response-validate-small` | 182904.74 | 173250.08 | 175253.01 | 175518.49 | **183540.00** | **Fiber** |
| JSON Validate Only (Large) | `POST /json-validate-large` | **37426.21** | 15861.35 | 15591.46 | 15864.84 | 20068.15 | **Gofi** |
| JSON Bind + Validate (Large) | `POST /json-bind-validate-large` | **25556.15** | 15774.91 | 15640.91 | 15868.30 | 20062.60 | **Gofi** |
| JSON Validation Failure (Early) | `POST /json-bind-validate-large` | **85706.76** | 24668.39 | 24516.60 | 15818.00 | 34793.43 | **Gofi** |
| JSON Validation Failure (Late/Deep) | `POST /json-bind-validate-large` | **26211.28** | 15466.58 | 15441.76 | 15609.22 | 19809.29 | **Gofi** |
| JSON Validate Response (Large) | `GET /json-response-validate-large` | 15932.32 | 61484.77 | **62157.16** | 45185.66 | 58795.93 | **Echo** |
| Multipart Bind | `POST /multipart` | **178506.50** | 135954.81 | 137357.17 | 138073.11 | 172044.16 | **Gofi** |
| FormData Bind | `POST /formdata` | **183490.78** | 165997.45 | 163199.06 | 166049.95 | 182909.93 | **Gofi** |
