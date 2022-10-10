package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

func uniqid(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000

	return fmt.Sprintf("%s%08x%05x", prefix, sec, usec)
}

// Links:
// * https://www.php2golang.com/method/function.uniqid.html
// * https://pkg.go.dev/crypto/sha1#pkg-functions
func shortenURL() string {
	var (
		randomChars   = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0987654321")
		randIntLength = 27
		stringLength  = 32
	)

	str := make([]rune, stringLength)

	for char := range str {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(randIntLength)))
		if err != nil {
			panic(err)
		}

		str[char] = randomChars[nBig.Int64()]
	}

	hash := sha256.Sum256([]byte(uniqid(string(str))))
	encodedString := base64.StdEncoding.EncodeToString(hash[:])

	return encodedString[0:9]
}

func defaultRouteHandler(context *fiber.Ctx) error {
	return context.Render("default", fiber.Map{})
}

func shortenURLHandler(context *fiber.Ctx, database *sql.DB) error {
	longURL := context.FormValue("url")
	_, err := url.ParseRequestURI(longURL)

	if err != nil {
		return context.Render(
			"default",
			fiber.Map{
				"error": "Sorry, but that is not a valid URL.",
				"url":   longURL,
			},
		)
	}

	shortenedURL := shortenURL()
	row := database.QueryRow(
		fmt.Sprintf(
			"INSERT INTO %s (long, short) VALUES(%s, %s)",
			pq.QuoteIdentifier(os.Getenv("DB_TABLE_NAME")),
			pq.QuoteLiteral(longURL),
			pq.QuoteLiteral(shortenedURL),
		),
	)

	if row.Err() != nil {
		var re = regexp.MustCompile(`duplicate key value`)

		log.Printf(
			"Issue shortening or persisting URL. Original URL: [%s]. Shortened URL: [%s] because: %s",
			longURL,
			shortenedURL,
			row.Err().Error(),
		)

		var err string

		if len(re.FindStringIndex(row.Err().Error())) > 0 {
			err = fmt.Sprintf("%s has already been shortened.", longURL)
		} else {
			err = fmt.Sprintf("There was an issue shortening '%s'.", longURL)
		}

		return fiber.NewError(fiber.StatusInternalServerError, err)
	}

	return context.RedirectToRoute("default", nil, fiber.StatusFound)
}

// shortURLRedirectHandler attempts to retrieve a long URL matching the short
// URL supplied in the request path. If the URL was shortened, then the user is
// redirected to the matching long URL. Otherwise, a 404 page is displayed.
func shortURLRedirectHandler(context *fiber.Ctx, database *sql.DB) error {
	var longURL string

	url := context.Params("url")
	err := database.QueryRow(
		fmt.Sprintf(
			"SELECT long FROM %s WHERE short = %s",
			pq.QuoteIdentifier(os.Getenv("DB_TABLE_NAME")),
			pq.QuoteLiteral(url),
		),
	).Scan(&longURL)

	if err != nil {
		log.Println(err.Error())

		return fiber.NewError(fiber.StatusNotFound, "Record not found")
	}

	return context.Redirect(longURL)
}

// Need to do the following
// . Set up Go templates so that it will render like the PHP version
// . Set up database interaction or some other data store
// . Set up 404 handling and exception.
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

	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			// Send custom error page
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
		return shortenURLHandler(context, database)
	})

	app.Get("/:url", func(context *fiber.Ctx) error {
		return shortURLRedirectHandler(context, database)
	})

	// Add a route for all static files
	app.Static("/", "./public")

	// Launch the application
	log.Fatal(app.Listen(":3000"))
}
