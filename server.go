package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/rs/zerolog"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type config struct {
	Port   int    `env:"PORT" envDefault:"4000"`
	LogDir string `env:"LOGDIR,expand" envDefault:"${HOME}/tmp"`
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("problem parsing config: %+v", err)
	}
	logger := setupLogger(context.Background(), filepath.Join(cfg.LogDir, "server.log"))

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	addrStr := fmt.Sprintf(":%d", cfg.Port)
	logger.Fatal().Err(http.ListenAndServe(addrStr, r))
}

func setupLogger(ctx context.Context, logFilePath string) *zerolog.Logger {
	var outWriter = os.Stdout
	if logFilePath != "" && logFilePath != "stdout" {
		file, err := os.OpenFile(logFilePath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.Fatalln(err)
		}
		outWriter = file
	}
	cout := zerolog.ConsoleWriter{Out: outWriter, TimeFormat: time.RFC822}
	cout.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	// uncomment to remove timestamp from logs
	//out.FormatTimestamp = func(i interface{}) string {
	//	return ""
	//}
	baseLogger := zerolog.New(cout).With().Timestamp().Logger()
	logCtx := baseLogger.WithContext(ctx)
	l := zerolog.Ctx(logCtx)
	return l
}
