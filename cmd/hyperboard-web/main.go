package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"

	"github.com/rs/zerolog"
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
	cfg   *Config
	api   *APIClient
	tmpls map[string]*template.Template
}

func run() error {
	cfg := loadConfig()

	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	tmpls, err := parseTemplates()
	if err != nil {
		return fmt.Errorf("failed to parse templates: %w", err)
	}

	app := &App{
		cfg:   cfg,
		api:   newAPIClient(cfg.APIURL, cfg.AdminPassword),
		tmpls: tmpls,
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
	protected.HandleFunc("/media/", app.handleMedia)
	protected.HandleFunc("/posts-partial", app.handleGallery)
	protected.HandleFunc("/posts/{id}", app.handlePost)
	protected.HandleFunc("/posts/{id}/note", app.handlePostNote)
	protected.HandleFunc("/posts/{id}/tags", app.handlePostTags)
	protected.HandleFunc("/posts/{id}/tags/{tag}", app.handlePostTags)
	protected.HandleFunc("/tag-suggestions", app.handleTagSuggestions)
	protected.HandleFunc("/upload", app.handleUpload)
	protected.HandleFunc("/tags", app.handleTags)
	protected.HandleFunc("POST /tags/{name}/convert-to-alias", app.handleTagConvertToAlias)
	protected.HandleFunc("/tags/{name}", app.handleTagEdit)
	protected.HandleFunc("/tag-categories", app.handleTagCategories)
	protected.HandleFunc("/tag-categories/{name}", app.handleTagCategoryEdit)
	protected.HandleFunc("/notes", app.handleNotes)
	protected.HandleFunc("/notes/{id}", app.handleNote)

	mux.Handle("/", app.sessionMiddleware(protected))

	log.Info().Str("port", cfg.Port).Msg("Starting web server")
	return http.ListenAndServe(":"+cfg.Port, mux)
}

// parseTemplates parses each page template together with the base layout
// so that each page gets its own "content" definition.
func parseTemplates() (map[string]*template.Template, error) {
	funcs := templateFuncs()
	base := "templates/base.html"

	// Page templates that use the base layout
	pages := []string{
		"templates/gallery.html",
		"templates/post.html",
		"templates/upload.html",
		"templates/tags.html",
		"templates/tag_edit.html",
		"templates/tag_categories.html",
		"templates/tag_category_edit.html",
		"templates/notes.html",
		"templates/note.html",
	}

	tmpls := make(map[string]*template.Template)

	for _, page := range pages {
		t, err := template.New("").Funcs(funcs).ParseFS(embeddedFiles, base, page)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", page, err)
		}
		// Register all defined templates by name so renderTemplate can find them
		for _, dt := range t.Templates() {
			if dt.Name() != "" {
				tmpls[dt.Name()] = t
			}
		}
	}

	// Standalone templates (no base layout)
	standalone := []string{"templates/login.html"}
	for _, s := range standalone {
		t, err := template.New("").Funcs(funcs).ParseFS(embeddedFiles, s)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", s, err)
		}
		for _, dt := range t.Templates() {
			if dt.Name() != "" {
				tmpls[dt.Name()] = t
			}
		}
	}

	return tmpls, nil
}

func (app *App) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, ok := app.tmpls[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
