package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v10"
	"github.com/go-chi/render"
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

type User struct {
	Id    string
	Email string
}

var allUsers = map[string]*User{
	"fece": {Id: "fece", Email: "bill@deadbug.com"},
	"d00f": {Id: "d00f", Email: "hhill@stricklandpropance.com"},
}

type UserResponse struct {
	*User
	Elapsed int64 `json:"elapsed"`
}

func (rd *UserResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	rd.Elapsed = 10
	return nil
}

func main() {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("problem parsing config: %+v", err)
	}
	logger := setupLogger(context.Background(), filepath.Join(cfg.LogDir, "server.log"))

	r := chi.NewRouter()
	r.Use(middleware.RequestID)                 // add an id to context
	r.Use(middleware.RealIP)                    // do the True-Client-IP, X-Real-IP or the X-Forwarded-For dance
	r.Use(middleware.Logger)                    // log requests
	r.Use(middleware.Recoverer)                 // panic recovery with http 500
	r.Use(middleware.Timeout(60 * time.Second)) // request timeout
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Golang Chi microservice template"))
	})

	r.Route("/users", func(r chi.Router) {
		r.With(paginate).Get("/", ListUsers)

		// Subrouters:
		r.Route("/{userID}", func(r chi.Router) {
			r.Use(UserCtx)
			r.Get("/", GetUser)
		})
	})

	addrStr := fmt.Sprintf(":%d", cfg.Port)
	logger.Fatal().Err(http.ListenAndServe(addrStr, r))
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	if err := render.RenderList(w, r, NewUserListResponse(allUsers)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)
	if err := render.Render(w, r, NewUserResponse(user)); err != nil {
		render.Render(w, r, ErrRender(err))
		return
	}
}

// UserCtx convenience middleware for user specific endpoints
func UserCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userID")
		user, err := retrieveUser(userID)
		if err != nil {
			http.Error(w, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// retrieveUser mock user record retrieval
func retrieveUser(userId string) (*User, error) {
	if u, ok := allUsers[userId]; ok {
		return u, nil
	}
	return nil, fmt.Errorf("no user with id: %s", userId)
}

func NewUserListResponse(users map[string]*User) []render.Renderer {
	list := []render.Renderer{}
	for _, user := range users {
		list = append(list, NewUserResponse(user))
	}
	return list
}
func NewUserResponse(user *User) *UserResponse {
	resp := &UserResponse{User: user}
	if resp.User == nil {
		if user, _ := retrieveUser(resp.Id); user != nil {
			resp.User = user
		}
	}
	return resp
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

func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// just a stub.. some ideas are to look at URL query params for something like
		// the page number, or the limit, and send a query cursor down the chain
		next.ServeHTTP(w, r)
	})
}

type ErrResponse struct {
	Err            error  `json:"-"`               // low-level runtime error
	HTTPStatusCode int    `json:"-"`               // http response status code
	StatusText     string `json:"status"`          // user-level status message
	ErrorText      string `json:"error,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

func ErrRender(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: 422,
		StatusText:     "Error rendering response.",
		ErrorText:      err.Error(),
	}
}
