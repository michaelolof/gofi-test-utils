package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/michaelolof/gofi-test-utils/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

func main() {
	port := "8082"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	data := make([]utils.SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = utils.SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	largeData := utils.GenerateLargeData()
	largeDataValidate := utils.GenerateLargeDataValidate()

	e := echo.New()
	e.HideBanner = true
	e.Validator = &CustomValidator{validator: validator.New()}

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/user/:name", func(c echo.Context) error {
		return c.String(http.StatusOK, c.Param("name"))
	})

	e.GET("/users/:userID/posts/:postID", func(c echo.Context) error {
		return c.String(http.StatusOK, c.Param("userID")+":"+c.Param("postID"))
	})

	e.POST("/json", func(c echo.Context) error {
		var p utils.SmallPayload
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/json-response", func(c echo.Context) error {
		return c.JSON(http.StatusOK, data)
	})

	e.POST("/json-large", func(c echo.Context) error {
		var p []utils.LargePayload
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/json-response-large", func(c echo.Context) error {
		return c.JSON(http.StatusOK, largeData)
	})

	e.POST("/json-validate-small", func(c echo.Context) error {
		var p utils.SmallPayloadValidate
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		if err := c.Validate(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/json-response-validate-small", func(c echo.Context) error {
		return c.JSON(http.StatusOK, utils.SmallPayloadValidate{ID: 1, Name: "test"})
	})

	e.POST("/json-validate-large", func(c echo.Context) error {
		var p []utils.LargePayloadValidate
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		for _, item := range p {
			if err := c.Validate(&item); err != nil {
				return c.String(http.StatusBadRequest, "Bad")
			}
		}
		return c.String(http.StatusOK, "OK")
	})

	e.GET("/json-response-validate-large", func(c echo.Context) error {
		return c.JSON(http.StatusOK, largeDataValidate)
	})

	e.POST("/multipart", func(c echo.Context) error {
		var p utils.SmallPayloadValidate
		p.Name = c.FormValue("name")
		if id, err := strconv.Atoi(c.FormValue("id")); err == nil {
			p.ID = id
		}
		if err := c.Validate(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		return c.String(http.StatusOK, "OK")
	})

	e.POST("/formdata", func(c echo.Context) error {
		var p utils.SmallPayloadValidate
		if err := c.Bind(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		if err := c.Validate(&p); err != nil {
			return c.String(http.StatusBadRequest, "Bad")
		}
		return c.String(http.StatusOK, "OK")
	})

	mw := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
	e.GET("/middlewares", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}, mw, mw, mw, mw, mw)

	e.GET("/query", func(c echo.Context) error {
		q := c.QueryParam("q")
		limit := c.QueryParam("limit")
		return c.String(http.StatusOK, q+limit)
	})

	log.Printf("Echo listening on :%s\n", port)
	log.Fatal(e.Start(":" + port))
}
