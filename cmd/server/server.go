package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/example/placeholder-webapp/assets"
	"github.com/example/placeholder-webapp/controllers"
	"github.com/example/placeholder-webapp/middlewares"
	"github.com/example/placeholder-webapp/models"
	"github.com/example/placeholder-webapp/templates"
	"github.com/example/placeholder-webapp/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

type config struct {
	Auth struct {
		Header string
	}

	Files struct {
		BasePath     string
		GeneralDir   string
		SafetyDir    string
		SchedulesDir string
	}

	Server struct {
		Address string
		CSP     string

		HSTS struct {
			MaxAge            int
			IncludeSubdomains bool
			Preload           bool
		}
	}
}

func loadEnvConfig() (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		log.Println("no .env file found --> using environment variables")
	}

	// Auth
	cfg.Auth.Header = os.Getenv("AUTH_HEADER")

	// Files
	cfg.Files.BasePath = os.Getenv("FILES_BASE_PATH")
	cfg.Files.GeneralDir = os.Getenv("FILES_GENERAL_DIR")
	cfg.Files.SafetyDir = os.Getenv("FILES_SAFETY_DIR")
	cfg.Files.SchedulesDir = os.Getenv("FILES_SCHEDULES_DIR")

	// Server
	cfg.Server.Address = os.Getenv("SERVER_ADDRESS")
	cfg.Server.CSP = os.Getenv("SERVER_CSP")

	// Server HSTS
	maxAge, err := strconv.Atoi(os.Getenv("SERVER_HSTS_MAX_AGE"))
	if err != nil {
		maxAge = 31536000
	}
	cfg.Server.HSTS.MaxAge = maxAge
	cfg.Server.HSTS.IncludeSubdomains, _ = strconv.ParseBool(os.Getenv("SERVER_HSTS_INCLUDE_SUBDOMAINS"))
	cfg.Server.HSTS.Preload, _ = strconv.ParseBool(os.Getenv("SERVER_HSTS_PRELOAD"))

	return cfg, nil
}

func run(cfg config) error {
	// Setup services
	fileService := &models.FileService{
		BasePath:     cfg.Files.BasePath,
		GeneralDir:   cfg.Files.GeneralDir,
		SafetyDir:    cfg.Files.SafetyDir,
		SchedulesDir: cfg.Files.SchedulesDir,
	}

	err := models.LoadTranslation("en", "en.json")
	if err != nil {
		panic(err)
	}

	err = models.LoadTranslation("fr", "fr.json")
	if err != nil {
		panic(err)
	}

	err = models.LoadTranslation("nl", "nl.json")
	if err != nil {
		panic(err)
	}

	// Setup controllers
	filesC := controllers.Files{
		FilesService: fileService,
	}
	filesC.Templates.List = views.Must(views.ParseFS(templates.FS, "tailwind.html", "files/list.html"))

	// Setup router and routes
	r := chi.NewRouter()

	r.Use(middlewares.EasyAuth(cfg.Auth.Header))
	// r.Use(middlewares.HSTS(cfg.Server.HSTS.MaxAge, cfg.Server.HSTS.IncludeSubdomains, cfg.Server.HSTS.Preload)) // is done by azure
	r.Use(middlewares.CSP(cfg.Server.CSP))
	r.Use(middlewares.NoSniff)
	r.Use(middleware.NoCache)

	r.Route("/", func(r chi.Router) {
		r.Get("/", filesC.List)
		r.Get("/general/{filename}", filesC.DownloadGeneral)
		r.Get("/safety/{filename}", filesC.DownloadSafety)
		r.Get("/schedule/{filename}", filesC.DownloadSchedule)
	})

	assetsHandler := http.FileServer(http.FS(assets.FS))
	r.Get("/assets/*", http.StripPrefix("/assets", assetsHandler).ServeHTTP)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Page not found", http.StatusNotFound)
	})

	log.Printf("starting the server on %s...\n", cfg.Server.Address)
	return http.ListenAndServe(cfg.Server.Address, r)
}

func main() {
	log.SetFlags(log.Lshortfile)

	cfg, err := loadEnvConfig()
	if err != nil {
		panic(err)
	}

	err = run(cfg)
	if err != nil {
		panic(err)
	}
}
