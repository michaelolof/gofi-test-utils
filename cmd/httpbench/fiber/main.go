package main

import (
	"fmt"
	"log"
	"os"

	"github.com/michaelolof/gofi-test-utils/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = validator.New()

func main() {
	port := "8084"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	data := make([]utils.SmallPayload, 100)
	for i := 0; i < 100; i++ {
		data[i] = utils.SmallPayload{ID: i, Name: fmt.Sprintf("Item %d", i)}
	}
	largeData := utils.GenerateLargeData()
	largeDataValidate := utils.GenerateLargeDataValidate()

	app := fiber.New(fiber.Config{DisableStartupMessage: true})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/user/:name", func(c *fiber.Ctx) error {
		return c.SendString(c.Params("name"))
	})

	app.Get("/users/:userID/posts/:postID", func(c *fiber.Ctx) error {
		return c.SendString(c.Params("userID") + ":" + c.Params("postID"))
	})

	app.Post("/json", func(c *fiber.Ctx) error {
		var p utils.SmallPayload
		if err := c.BodyParser(&p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})

	app.Get("/json-response", func(c *fiber.Ctx) error {
		return c.JSON(data)
	})

	app.Post("/json-large", func(c *fiber.Ctx) error {
		var p []utils.LargePayload
		if err := c.BodyParser(&p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})

	app.Get("/json-response-large", func(c *fiber.Ctx) error {
		return c.JSON(largeData)
	})

	app.Post("/json-validate-small", func(c *fiber.Ctx) error {
		var p utils.SmallPayloadValidate
		if err := c.BodyParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := validate.Struct(p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})

	app.Get("/json-response-validate-small", func(c *fiber.Ctx) error {
		return c.JSON(utils.SmallPayloadValidate{ID: 1, Name: "test"})
	})

	app.Post("/json-validate-large", func(c *fiber.Ctx) error {
		var p []utils.LargePayloadValidate
		if err := c.BodyParser(&p); err != nil {
			return c.SendStatus(400)
		}
		for _, item := range p {
			if err := validate.Struct(item); err != nil {
				return c.SendStatus(400)
			}
		}
		return c.SendString("OK")
	})

	app.Get("/json-response-validate-large", func(c *fiber.Ctx) error {
		return c.JSON(largeDataValidate)
	})

	app.Post("/multipart", func(c *fiber.Ctx) error {
		var p utils.SmallPayloadValidate
		if err := c.BodyParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := validate.Struct(p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})

	app.Post("/formdata", func(c *fiber.Ctx) error {
		var p utils.SmallPayloadValidate
		if err := c.BodyParser(&p); err != nil {
			return c.SendStatus(400)
		}
		if err := validate.Struct(p); err != nil {
			return c.SendStatus(400)
		}
		return c.SendString("OK")
	})

	mw := func(c *fiber.Ctx) error {
		return c.Next()
	}
	app.Get("/middlewares", mw, mw, mw, mw, mw, func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/query", func(c *fiber.Ctx) error {
		q := c.Query("q")
		limit := c.Query("limit")
		return c.SendString(q + limit)
	})

	log.Printf("Fiber listening on :%s\n", port)
	log.Fatal(app.Listen(":" + port))
}
