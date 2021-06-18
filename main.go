package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/cssivision/reverseproxy"
	"github.com/gin-gonic/gin"
)

var reactRouterPaths = []string{
	"/alerts",
	"/config",
	"/flags",
	"/graph",
	"/rules",
	"/service-discovery",
	"/status",
	"/targets",
	"/tsdb-status",
	"/starting",
}

type Service struct {
	backendUrl string
	staticDir  string
	title      string
	ginEngine  *gin.Engine
}

func NewService() *Service {
	r := gin.Default()
	s := &Service{
		ginEngine: r,
	}
	return s
}

func (s *Service) ParseEnv() {
	backendUrl := os.Getenv("BACKEND_URL")
	if backendUrl == "" {
		backendUrl = "http://localhost:8000"
	}
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./react/build"
	}
	title := os.Getenv("TITLE")
	if title == "" {
		title = "prometheus"
	}
	s.backendUrl = backendUrl
	s.staticDir = staticDir
	s.title = title
}

func (s *Service) RegisterRoutes() {
	for _, p := range reactRouterPaths {
		s.ginEngine.GET(p, func(c *gin.Context) {
			idx, err := ioutil.ReadFile(s.staticDir + "/index.html")
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			replacedIdx := bytes.ReplaceAll(idx, []byte("TITLE_PLACEHOLDER"), []byte(s.title))
			c.Data(http.StatusOK, "text/html", replacedIdx)
		})
	}

	s.ginEngine.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/graph") })
	s.ginEngine.StaticFS("/static", gin.Dir(s.staticDir+"/static", false))
	s.ginEngine.GET("/new", func(c *gin.Context) { c.Redirect(http.StatusFound, path.Join("/", "new")) })
	s.ginEngine.StaticFS("/new/static", gin.Dir(s.staticDir+"/static", false))
	s.ginEngine.GET("/new/:path", func(c *gin.Context) {
		p := c.Param("path")
		for _, rp := range reactRouterPaths {
			if "/"+p != rp {
				continue
			}
			idx, err := ioutil.ReadFile(s.staticDir + "/index.html")
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}
			replacedIdx := bytes.ReplaceAll(idx, []byte("TITLE_PLACEHOLDER"), []byte(s.title))
			c.Data(http.StatusOK, "text/html", replacedIdx)
			return
		}
	})
	s.ginEngine.NoRoute(func(c *gin.Context) { c.Status(http.StatusNotFound) })
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
	s := NewService()
	s.ParseEnv()
	s.RegisterRoutes()
	if err := http.ListenAndServe(":8080", s); err != nil {
		panic(err)
	}
}
