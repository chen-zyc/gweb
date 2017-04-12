package gweb

import (
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
)

type RouterGroup struct {
	s              *Server
	basePath       string
	globalHandlers []Handler
}

func NewGroup(s *Server, basePath string, handlers ...Handler) *RouterGroup {
	return &RouterGroup{
		s:              s,
		basePath:       basePath,
		globalHandlers: handlers,
	}
}

func (g *RouterGroup) Group(relativePath string, handlers ...Handler) *RouterGroup {
	absolutePath := joinPaths(g.basePath, relativePath)
	handlers = g.combineHandlers(handlers...)
	return NewGroup(g.s, absolutePath, handlers...)
}

func (g *RouterGroup) Global(handlers ...Handler) {
	g.globalHandlers = append(g.globalHandlers, handlers...)
}

func (g *RouterGroup) Handle(method, path string, handlers ...Handler) {
	Assert(path[0] == '/', fmt.Sprintf("path must begin with '/' in path '%s'", path))
	Assert(method != "", fmt.Sprintf("HTTP method can not be empty in path '%s'", path))
	Assert(len(handlers) > 0, "there must be at least one handler")
	if matched, err := regexp.MatchString("^[A-Z]+$", method); !matched || err != nil {
		panic(fmt.Sprintf("http method '%s' is not valid", method))
	}

	router := g.s.trees[method]
	if router == nil {
		router = NewRouter()
		g.s.trees[method] = router
	}
	handlers = g.combineHandlers(handlers...) // + global handlers
	absolutePath := joinPaths(g.basePath, path)
	router.Add(absolutePath, handlers)
}

func (g *RouterGroup) GET(path string, handlers ...Handler) {
	g.Handle(MethodGet, path, handlers...)
}

func (g *RouterGroup) POST(path string, handlers ...Handler) {
	g.Handle(MethodPost, path, handlers...)
}

func (g *RouterGroup) PUT(path string, handlers ...Handler) {
	g.Handle(MethodPut, path, handlers...)
}

func (g *RouterGroup) DELETE(path string, handlers ...Handler) {
	g.Handle(MethodDelete, path, handlers...)
}

func (g *RouterGroup) PATCH(path string, handlers ...Handler) {
	g.Handle(MethodPatch, path, handlers...)
}

func (g *RouterGroup) OPTIONS(path string, handlers ...Handler) {
	g.Handle(MethodOptions, path, handlers...)
}

func (g *RouterGroup) HEAD(path string, handlers ...Handler) {
	g.Handle(MethodHead, path, handlers...)
}

func (g *RouterGroup) CONNECT(path string, handlers ...Handler) {
	g.Handle(MethodConnect, path, handlers...)
}

func (g *RouterGroup) TRACE(path string, handlers ...Handler) {
	g.Handle(MethodTrace, path, handlers...)
}

func (g *RouterGroup) HandleMethods(methods []string, path string, handlers ...Handler) {
	for _, method := range methods {
		g.Handle(strings.TrimSpace(method), path, handlers...)
	}
}

func (g *RouterGroup) Any(path string, handlers ...Handler) {
	allMethods := []string{
		MethodGet, MethodPost, MethodHead, MethodOptions, MethodPut,
		MethodDelete, MethodTrace, MethodConnect, MethodPatch,
	}
	g.HandleMethods(allMethods, path, handlers...)
}

func (g *RouterGroup) StaticFile(path, filePath string) {
	if strings.ContainsAny(path, ":*") {
		panic("URL parameters can not be used when serving a static file")
	}

	g.HandleMethods([]string{MethodGet, MethodHead}, path, func(c *Context) {
		c.File(filePath)
	})
}

func (g *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) {
	if strings.ContainsAny(relativePath, ":*") {
		panic("URL parameters can not be used when serving a static file")
	}
	absolutePath := joinPaths(g.basePath, relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))

	var handler Handler
	if _, ok := fs.(*OnlyFilesFS); ok {
		handler = func(c *Context) {
			c.resp.WriteHeader(http.StatusNotFound)
			fileServer.ServeHTTP(c.resp, c.req)
		}
	} else {
		handler = func(c *Context) {
			fileServer.ServeHTTP(c.resp, c.req)
		}
	}

	urlPath := path.Join(relativePath, "/*filepath")
	g.HandleMethods([]string{MethodGet, MethodHead}, urlPath, handler)
}

func (g *RouterGroup) StaticDir(relativePath, root string) {
	fs := &OnlyFilesFS{http.Dir(root)}
	g.StaticFS(relativePath, fs)
}

func (g *RouterGroup) combineHandlers(handlers ...Handler) Handlers {
	if len(g.globalHandlers) == 0 {
		return handlers
	}
	numHandlers := len(g.globalHandlers) + len(handlers)
	mergedHandlers := make([]Handler, numHandlers)
	copy(mergedHandlers, g.globalHandlers)
	copy(mergedHandlers[len(g.globalHandlers):], handlers)
	return mergedHandlers
}
