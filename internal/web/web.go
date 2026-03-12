package web

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dharmab/hyperboard/internal/middleware/logging"
	"github.com/dharmab/hyperboard/internal/middleware/security"
	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//go:embed templates static
var embeddedFiles embed.FS

// newResourceName is a sentinel path value indicating a new resource is being created.
const newResourceName = "_new"

// configPath is the path to the configuration file, set via CLI flag.
var configPath string

// NewCommand returns the cobra command for the hyperboard-web server.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hyperboard-web",
		Short: "Hyperboard web frontend",
		RunE:  func(cmd *cobra.Command, args []string) error { return run() },
	}
	cmd.Flags().StringVar(&configPath, "config", "", "Path to configuration file")
	bindConfig(cmd)
	cobra.OnInitialize(initConfig)
	return cmd
}

// initConfig reads the config file if configPath is set.
func initConfig() {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			log.Fatal().Err(err).Str("config", configPath).Msg("Failed to read config file")
		}
	}
}

// app holds the web application state including configuration, API client, and templates.
type app struct {
	cfg   *config
	api   *client.ClientWithResponses
	media *mediaClient
	tmpls map[string]*template.Template
}

// run initializes configuration, templates, and starts the web server.
func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

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

	api, err := newAPIClient(cfg.APIURL, cfg.AdminPassword)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	a := &app{
		cfg:   cfg,
		api:   api,
		media: newMediaClient(cfg.APIURL, cfg.AdminPassword),
		tmpls: tmpls,
	}

	staticFS, _ := fs.Sub(embeddedFiles, "static")

	mux := http.NewServeMux()

	apiProxy, err := newAPIProxy(cfg.APIURL)
	if err != nil {
		return fmt.Errorf("failed to create API proxy: %w", err)
	}
	mux.Handle("/api/", apiProxy)

	mux.HandleFunc("/login", maxBody(maxFormBody, a.handleLogin))
	mux.HandleFunc("/logout", a.handleLogout)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	protected := http.NewServeMux()
	a.registerRoutes(protected)
	mux.Handle("/", a.sessionMiddleware(protected))

	httpServer := &http.Server{
		Handler:           logging.RequestLoggingMiddleware(security.SecurityHeadersMiddleware(mux)),
		Addr:              ":" + cfg.Port,
		ReadHeaderTimeout: 30 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("port", cfg.Port).Msg("Starting web server")
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("failed to serve: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.Info().Msg("Shutting down web server...")
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(timeoutCtx); err != nil {
			return fmt.Errorf("failed to shut down web server: %w", err)
		}
		log.Info().Msg("Web server stopped")
		return nil
	}
}

// parseTemplates parses each page template together with the base layout
// so that each page gets its own "content" definition.
func parseTemplates() (map[string]*template.Template, error) {
	funcs := templateFuncs()
	base := "templates/base.html"

	pages := []string{
		"templates/posts.html",
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
		for _, dt := range t.Templates() {
			if dt.Name() != "" {
				tmpls[dt.Name()] = t
			}
		}
	}

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

func newAPIProxy(apiURL string) (http.Handler, error) {
	target, err := url.Parse(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return proxy, nil
}

func (a *app) renderTemplate(w http.ResponseWriter, r *http.Request, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t, ok := a.tmpls[name]
	if !ok {
		http.Error(w, "Template not found", http.StatusInternalServerError)
		return
	}
	if err := t.ExecuteTemplate(w, name, data); err != nil {
		log.Error().Err(err).Str("template", name).Msg("template execution failed")
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
