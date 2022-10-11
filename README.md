# Go URL Shortener

This is a small, simplistic example of creating a URL shortener in Go.

## Requirements

To use this application, you will need the following:

- [Docker Engine](https://docs.docker.com/engine/install/) and [Docker Compose](https://docs.docker.com/compose/install/) **or** PHP 8.1 with the [PDO](https://www.php.net/manual/en/pdo.installation.php) and [PDO_PGSQL](https://www.php.net/manual/en/ref.pdo-pgsql.php) extensions and [PostgreSQL](https://www.postgresql.org/) 14 or above.
- Your favourite IDE or code editor
- [Go](https://go.dev/dl/) 1.19 or above

## Usage

First, clone the project to a directory on your local machine by running the following command:

```bash
git clone git@github.com:settermjd/go-url-shortener.git go-url-shortener
```

Then, start the application, using Docker Compose by running the following command in the top-level directory of the project.

```bash
docker compose up -d --build
```

**Note:** The first time that you run the command, if you don’t have one or more images in your local Docker cache, then they have to be downloaded.
This shouldn't take too long, allowing for the speed of your internet connection.

**New to Docker Compose and want a hand getting started?**
Then grab a copy of my free book: [Deploy with Docker Compose](https://deploywithdockercompose.com/).

## Have Questions?

If you have any questions or queries, either create [an issue](https://github.com/settermjd/go-url-shortener/issues/new/choose) or a PR, or email me: matthew[at]matthewsetter.com.
