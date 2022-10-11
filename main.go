package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/joho/godotenv"
	"github.com/settermjd/url-shortener/urlshortener"
)

// defaultRouteHandler handles GET requests to the default or main route of the
// application.
func defaultRouteHandler(context *fiber.Ctx) error {
	return context.Render("default", fiber.Map{})
}

// shortenURLHandler handles POST requests to the default route in the application.
// This is when a shortened URL is created and link in the database against the
// original URL submitted in the request's POST data.
func shortenURLHandler(context *fiber.Ctx, shortener *urlshortener.URLShortener) error {
	longURL := context.FormValue("url")

	if _, err := url.ParseRequestURI(longURL); err != nil {
		return context.Render(
			"default",
			fiber.Map{
				"error": "Sorry, but that is not a valid URL.",
				"url":   longURL,
			},
		)
	}

	shortURL := shortener.ShortenURL()
	row := shortener.PersistURL(longURL, shortURL)
	if errorMessage := shortener.GetURLPersistenceError(row); errorMessage != "" {
		log.Print(errorMessage)
		return fiber.NewError(fiber.StatusInternalServerError, errorMessage)
	}

	return context.Render("default", fiber.Map{
		"baseURL":  "http://localhost:3001",
		"longURL":  longURL,
		"shortURL": shortURL,
		"success":  true,
	})
}

// shortURLRedirectHandler attempts to retrieve a long URL matching the short
// URL supplied in the request path. If the URL was shortened, then the user is
// redirected to the matching long URL. Otherwise, a 404 page is displayed.
func shortURLRedirectHandler(context *fiber.Ctx, shortener *urlshortener.URLShortener) error {
	longURL, err := shortener.GetLongURL(context.Params("url"))
	if err != nil {
		log.Println(err.Error())
		return fiber.NewError(fiber.StatusNotFound, "Record not found")
	}

	return context.Redirect(longURL)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s?sslmode=disable",
		os.Getenv("DB_USERNAME"),
		os.Getenv("DB_PASSWORD"),
		net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		os.Getenv("DB_NAME"),
	)
	database, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	shortener := urlshortener.NewURLShortener(database, os.Getenv("DB_TABLE_NAME"))

	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			err = ctx.Status(code).Render("404", fiber.Map{
				"error": err,
			})

			if err != nil {
				return ctx.
					Status(fiber.StatusInternalServerError).
					SendString("Internal Server Error")
			}

			return nil
		},
	})

	app.Get("/", defaultRouteHandler).Name("default")

	app.Post("/", func(context *fiber.Ctx) error {
		return shortenURLHandler(context, shortener)
	})

	app.Get("/:url", func(context *fiber.Ctx) error {
		return shortURLRedirectHandler(context, shortener)
	})

	app.Static("/", "./public")

	log.Fatal(app.Listen(":3000"))
}
