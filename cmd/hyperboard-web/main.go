package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed templates static
var embeddedFiles embed.FS

var configPath string

var cmd = &cobra.Command{
	Use:   "hyperboard-web",
	Short: "Hyperboard web frontend",
	RunE:  func(cmd *cobra.Command, args []string) error { return run() },
}

func init() {
	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
	bindConfig(cmd)
}

func main() {
	cobra.OnInitialize(initConfig)
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start web server")
	}
}

func initConfig() {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Str("config", configPath).Msg("Failed to read config file")
		}
	}
}

type App struct {
	cfg  *Config
	api  *APIClient
	tmpl *template.Template
}

func run() error {
	cfg := loadConfig()

	tmpl, err := template.New("").Funcs(templateFuncs()).ParseFS(embeddedFiles, "templates/*.html")
	if err != nil {
		// No templates yet — use empty template set
		tmpl = template.New("").Funcs(templateFuncs())
	}

	app := &App{
		cfg:  cfg,
		api:  newAPIClient(cfg.APIURL, cfg.AdminPassword),
		tmpl: tmpl,
	}

	staticFS, _ := fs.Sub(embeddedFiles, "static")

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/login", app.handleLogin)
	mux.HandleFunc("/logout", app.handleLogout)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("/", app.handleGallery)
	protected.HandleFunc("/posts-partial", app.handleGallery)
	protected.HandleFunc("/posts/{id}", app.handlePost)
	protected.HandleFunc("/posts/{id}/note", app.handlePostNote)
	protected.HandleFunc("/posts/{id}/tags", app.handlePostTags)
	protected.HandleFunc("/posts/{id}/tags/{tag}", app.handlePostTags)
	protected.HandleFunc("/tag-suggestions", app.handleTagSuggestions)
	protected.HandleFunc("/upload", app.handleUpload)
	protected.HandleFunc("/tags", app.handleTags)
	protected.HandleFunc("/tags/{name}", app.handleTagEdit)
	protected.HandleFunc("/tag-categories", app.handleTagCategories)
	protected.HandleFunc("/tag-categories/{name}", app.handleTagCategoryEdit)
	protected.HandleFunc("/notes", app.handleNotes)
	protected.HandleFunc("/notes/{id}", app.handleNote)

	mux.Handle("/", app.sessionMiddleware(protected))

	log.Info().Str("port", cfg.Port).Msg("Starting web server")
	return http.ListenAndServe(":"+cfg.Port, mux)
}

func (app *App) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := app.tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
