package codegen

import (
	"net/http"
)

// HeaderQueryPathSchema tests basic request parameter extraction and binding.
type HeaderQueryPathSchema struct {
	Request struct {
		Header struct {
			Authorization string `json:"Authorization" validate:"required"`
			XRequestID    string `json:"X-Request-Id" default:"auto-generated"`
		}
		Query struct {
			Page   int    `json:"page" default:"1"`
			Sort   string `json:"sort"`
			Active bool   `json:"active"`
		}
		Path struct {
			ID       int     `json:"id" validate:"required"`
			Category string  `json:"category" validate:"required"`
			Rating   float64 `json:"rating"`
		}
	}
	Ok struct {
		Body struct {
			Message string `json:"message"`
		}
	}
}

// CookieSchema tests cookie extraction including http.Cookie types.
type CookieSchema struct {
	Request struct {
		Cookie struct {
			SessionID string      `json:"session_id" validate:"required"`
			Tracking  http.Cookie `json:"tracking"`
		}
	}
	Ok struct {
		Cookie struct {
			AuthToken http.Cookie `json:"auth_token"`
		}
		Body struct {
			Authenticated bool `json:"authenticated"`
		}
	}
}

// JSONBodySchema tests JSON body parsing with struct body.
type JSONBodySchema struct {
	Request struct {
		Body struct {
			Name  string  `json:"name" validate:"required"`
			Email string  `json:"email" validate:"required"`
			Age   int     `json:"age" validate:"min=1,max=150"`
			Score float64 `json:"score"`
		} `validate:"required"`
	}
	Ok struct {
		Body struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
	}
	BadRequest struct {
		Body struct {
			Message string `json:"message"`
		}
	}
}

// ResponseHeaderSchema tests response header writing.
type ResponseHeaderSchema struct {
	Request struct {
		Path struct {
			ID int `json:"id" validate:"required"`
		}
	}
	Ok struct {
		Header struct {
			XRequestID string `json:"X-Request-Id"`
			XVersion   string `json:"X-Version"`
		}
		Body struct {
			ID int `json:"id"`
		}
	}
}

// CustomSpecSchema tests custom spec encoding and decoding.
type CustomSpecSchema struct {
	Request struct {
		Query struct {
			Amount int `json:"amount" spec:"currency" validate:"required"`
		}
	}
	Ok struct {
		Header struct {
			Price int `json:"X-Price" spec:"currency"`
		}
		Body string
	}
}

// InlineValidationsSchema tests inline rule evaluation like min, max, required, gte, lte
type InlineValidationsSchema struct {
	Request struct {
		Query struct {
			Code string `json:"code" validate:"required,min=4,max=8"`
			Age  int    `json:"age" validate:"gte=18,lte=65"`
		}
	}
	Ok struct {
		Body string
	}
}
