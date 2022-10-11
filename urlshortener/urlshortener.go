package urlshortener

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/lib/pq"
)

// uniqid returns a unique id string useful when generating random strings.
// It was lifted from https://www.php2golang.com/method/function.uniqid.html.
func uniqid(prefix string) string {
	now := time.Now()
	sec := now.Unix()
	usec := now.UnixNano() % 0x100000

	return fmt.Sprintf("%s%08x%05x", prefix, sec, usec)
}

// URLShortener is a coordinating struct for providing a URL shortening service.
type URLShortener struct {
	database *sql.DB
	table    string
}

// NewURLShortener returns an initialised instance of a URLShortener struct
func NewURLShortener(database *sql.DB, table string) *URLShortener {
	shortener := URLShortener{database: database, table: table}

	return &shortener
}

// GetLongURL retrieves a longer/original URL from the database, based
// on the shortURL provided, or an error if it could not be found.
func (shortener *URLShortener) GetLongURL(shortURL string) (string, error) {
	var longURL string

	err := shortener.database.QueryRow(
		fmt.Sprintf(
			"SELECT long FROM %s WHERE short = %s",
			pq.QuoteIdentifier(shortener.table),
			pq.QuoteLiteral(shortURL),
		),
	).Scan(&longURL)
	if err != nil {
		return "", err
	}

	return longURL, nil
}

// PersistURL persists a short and long URL combination to the database.
func (shortener *URLShortener) PersistURL(longURL, shortURL string) *sql.Row {
	return shortener.database.QueryRow(
		fmt.Sprintf(
			"INSERT INTO %s (long, short) VALUES(%s, %s)",
			pq.QuoteIdentifier(shortener.table),
			pq.QuoteLiteral(longURL),
			pq.QuoteLiteral(shortURL),
		),
	)
}

// GetURLPersistenceError returns an error string based on an sql.Row error, if
// one is available.
func (shortener *URLShortener) GetURLPersistenceError(row *sql.Row) string {
	if row.Err() != nil {
		re := regexp.MustCompile(`duplicate key value`)

		if len(re.FindStringIndex(row.Err().Error())) > 0 {
			return "The URL has already been shortened."
		}

		return "There was an issue shortening the URL."
	}

	return ""
}

// ShortenURL generates and returns a short URL string.
func (shortener *URLShortener) ShortenURL() string {
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
