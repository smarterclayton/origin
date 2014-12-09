package server

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/openshift/origin/pkg/api/latest"
	appgen "github.com/openshift/origin/pkg/generate/app"
)

type Config struct {
	BindAddress  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewConfig(bindAddress string) *Config {
	if len(bindAddress) == 0 {
		bindAddress = ":80"
	}
	return &Config{
		BindAddress:  bindAddress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func Serve(cfg *Config) {
	m := http.NewServeMux()
	m.HandleFunc("/generate", handleGenerate)

	s := &http.Server{
		Addr:         cfg.BindAddress,
		Handler:      m,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	log.Printf("Starting generator server on bind address: %s\n", cfg.BindAddress)
	log.Fatal(s.ListenAndServe())
}

func handleGenerate(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		log.Printf("Error occurred parsing request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	images := []string{}
	imagesStr := req.Form.Get("images")
	if len(imagesStr) > 0 {
		images = strings.Split(imagesStr, ",")
	}
	g := appgen.Generator{
		Source:       req.Form.Get("source"),
		Name:         req.Form.Get("name"),
		BuilderImage: req.Form.Get("builderImage"),
		Images:       images,
	}
	cfg, err := g.Generate()
	if err != nil {
		log.Printf("Error occurred generating artifacts: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := latest.Codec.Encode(cfg)
	if err != nil {
		log.Printf("Error occurred serializing configuration: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
