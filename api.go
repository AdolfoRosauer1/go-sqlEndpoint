package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"log"
	"net/http"
	"os"
)

// Media model for psql schema table. Includes tags for JSON deserialization
type Media struct {
	MediaId          int         `json:"mediaid"`
	Type             bool        `json:"type"`
	Name             string      `json:"name"`
	OriginalLanguage string      `json:"originallanguage"`
	Adult            bool        `json:"adult"`
	ReleaseDate      pgtype.Date `json:"releasedate"`
	Overview         string      `json:"overview"`
	BackdropPath     string      `json:"backdroppath"`
	PosterPath       string      `json:"posterpath"`
	TrailerLink      string      `json:"trailerlink"`
	TMDBRating       float32     `json:"tmdbrating"`
	Status           string      `json:"status"`
}

// global db connection
var db *pgx.Conn

// global err status for the db connection
var err error

func main() {
	db, err = pgx.Connect(context.Background(), os.Getenv("MOOVIE_DB"))
	if err != nil {
		log.Fatalf("unable to connect to database: %v", err)
	}
	defer func(db *pgx.Conn, ctx context.Context) {
		err = db.Close(ctx)
		if err != nil {
			log.Fatalf("unable to close database: %v", err)
		}
	}(db, context.Background())

	router := gin.Default()

	router.GET("/movie", getMovies)
	router.GET("/movie/:id", getMovie)

	log.Fatal(router.Run(":6969"))
}

func getMovies(c *gin.Context) {
	rows, qerr := db.Query(context.Background(), `SELECT * FROM media LIMIT 50`)
	if qerr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error reading rows"})
			return
		}

		item := make(map[string]interface{})
		for i, col := range rows.FieldDescriptions() {
			item[string(col.Name)] = values[i]
		}

		items = append(items, item)
	}

	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating rows"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func getMovie(c *gin.Context) {
	toReturn := Media{}

	qerr := db.QueryRow(context.Background(), `select * from media where mediaid=$1`, c.Param(`id`)).Scan(
		&toReturn.MediaId, &toReturn.Type, &toReturn.Name, &toReturn.OriginalLanguage, &toReturn.Adult,
		&toReturn.ReleaseDate, &toReturn.Overview, &toReturn.BackdropPath,
		&toReturn.PosterPath, &toReturn.TrailerLink, &toReturn.TMDBRating, &toReturn.Status)
	if qerr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}

	c.JSON(http.StatusOK, toReturn)
}
