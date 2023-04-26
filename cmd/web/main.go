package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"snippetbox/internal/models"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

type app struct {
	errLogger     *log.Logger
	infoLogger    *log.Logger
	snippets      *models.SnippetModel
	templaceCache map[string]*template.Template
}

var config struct {
	addr string
	dsn  string
}

func main() {
	flag.StringVar(&config.addr, "addr", ":5000", "HTTP network address")
	flag.StringVar(&config.dsn, "dsn", "root:password@/snippetbox?parseTime=true", "MySQL connect name")

	flag.Parse()

	infoLogger := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLogger := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	db, err := openDB(config.dsn)

	if err != nil {
		errLogger.Fatal(err)
	}

	defer db.Close()

	templateCache, err := newTemplateCache()

	if err != nil {
		errLogger.Fatal(err)
	}

	app := app{
		errLogger:     errLogger,
		infoLogger:    infoLogger,
		snippets:      &models.SnippetModel{DB: db},
		templaceCache: templateCache,
	}

	// use custom logger for server
	srv := &http.Server{
		Addr:     config.addr,
		ErrorLog: errLogger,
		Handler:  app.routes(),
	}

	infoLogger.Printf("Started listening on: %s", config.addr)
	err = srv.ListenAndServe()
	errLogger.Fatal(err)
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}