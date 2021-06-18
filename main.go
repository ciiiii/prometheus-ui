package main

import (
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cssivision/reverseproxy"
	"github.com/gin-gonic/gin"
)

type Service struct {
	backendUrl string
	ginEngine  *gin.Engine
}

func NewService(backendUrl, staticDir string) *Service {
	r := gin.Default()
	r.StaticFS("/", gin.Dir(staticDir, false))
	// r.Use(static.Serve("/*", static.LocalFile(staticDir, true)))

	s := &Service{
		backendUrl: backendUrl,
		ginEngine:  r,
	}
	return s
}

func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		u, _ := url.Parse(s.backendUrl)
		reverseproxy.NewReverseProxy(u).ServeHTTP(w, r)
		return
	}
	s.ginEngine.ServeHTTP(w, r)
}

func main() {
	backendUrl := os.Getenv("BACKEND_URL")
	if backendUrl == "" {
		backendUrl = "http://localhost:8000"
	}
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./react/build"
	}
	s := NewService(backendUrl, staticDir)
	if err := http.ListenAndServe(":8080", s); err != nil {
		panic(err)
	}
}
