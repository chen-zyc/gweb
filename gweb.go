package gweb

import (
	"net/http"
	"strings"
	"sync"
)

const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodTrace   = "TRACE"
	MethodConnect = "CONNECT"
	MethodPatch   = "PATCH"
)

type Handler func(c *Context)

type Handlers []Handler

type Server struct {
	*RouterGroup
	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router automatically replies to OPTIONS requests.
	// Custom OPTIONS handlers take priority over automatic replies.
	HandleOPTIONS bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// Configurable Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	// The "Allow" header with allowed request methods is set before the handler
	// is called.
	MethodNotAllowed Handler

	// Configurable Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	PanicHandler func(ctx *Context, err interface{})

	trees          map[string]Router
	ctxPool        sync.Pool
	globalHandlers Handlers
}

var _ http.Handler = (*Server)(nil)

func NewServer() *Server {
	s := &Server{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleOPTIONS:          true,
		HandleMethodNotAllowed: true,
		trees: make(map[string]Router, 9),
	}
	s.RouterGroup = NewGroup(s, "/")
	s.ctxPool.New = func() interface{} { return &Context{} }
	return s
}

func (s *Server) Run(address string, opts ...Option) error {
	for _, opt := range opts {
		opt(s)
	}

	err := http.ListenAndServe(address, s)
	return err
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := s.getContext()
	ctx.reset(req, w)

	if s.PanicHandler != nil {
		defer func() {
			if err := recover(); err != nil {
				s.PanicHandler(ctx, err)
			}
		}()
	}
	s.handleRequest(ctx)
	s.putContext(ctx)
}

func (s *Server) handleRequest(ctx *Context) {
	req := ctx.req
	method, path := req.Method, req.URL.Path

	if router := s.trees[method]; router != nil {
		handlers, params, tsr := router.Find(path)
		if handlers != nil {
			ctx.params = params
			ctx.handlers = handlers
			ctx.Next()
			return
		}
		if method != MethodConnect && path != "/" {
			code := http.StatusMovedPermanently // Permanent redirect, request with GET method
			if method != MethodGet {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = http.StatusTemporaryRedirect
			}
			if tsr && s.RedirectTrailingSlash {
				if pathLen := len(path); pathLen > 1 && path[pathLen-1] == '/' {
					req.URL.Path = path[:pathLen-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(ctx.resp, req, req.URL.String(), code)
				return
			}
			// try to fix the request path
			if s.RedirectFixedPath {
				fixedPath, found := router.FindCaseInsensitivePath(CleanPath(path), s.RedirectTrailingSlash)
				if found {
					req.URL.Path = string(fixedPath)
					http.Redirect(ctx.resp, req, req.URL.String(), code)
					return
				}
			}
		}
	}

	if method == MethodOptions {
		if s.HandleOPTIONS {
			if allow := s.allowed(path, method); allow != "" {
				ctx.resp.Header().Set("Allow", allow)
				return
			}
		}
	} else {
		// handle 405
		if s.HandleMethodNotAllowed {
			if allow := s.allowed(path, method); allow != "" {
				ctx.resp.Header().Set("Allow", allow)
				s.methodNotAllowed(ctx)
				return
			}
		}
	}

	// handle 404
	s.notFound(ctx)
}

func (s *Server) getContext() *Context { return s.ctxPool.Get().(*Context) }

func (s *Server) putContext(c *Context) { s.ctxPool.Put(c) }

func (s *Server) allowed(path, method string) (allow string) {
	allowSlice := make([]string, 0, len(s.trees)+1)
	if path == "*" { // server-wide
		for m := range s.trees {
			if m == MethodOptions {
				continue
			}
			allowSlice = append(allowSlice, m)
		}
	} else { // specific path
		for m := range s.trees {
			// Skip the requested method - we already tried this one
			if method == m || m == MethodOptions {
				continue
			}
			handler, _, _ := s.trees[m].Find(path)
			if handler != nil {
				allowSlice = append(allowSlice, m)
			}
		}
	}
	if len(allowSlice) > 0 {
		allowSlice = append(allowSlice, MethodOptions)
		allow = strings.Join(allowSlice, ", ")
	}
	return
}

func (s *Server) methodNotAllowed(ctx *Context) {
	if s.MethodNotAllowed == nil {
		s.MethodNotAllowed = func(c *Context) {
			http.Error(c.resp, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
	s.MethodNotAllowed(ctx)
}

func (s *Server) notFound(ctx *Context) {
	if s.NotFound == nil {
		s.NotFound = func(c *Context) {
			http.Error(c.resp, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		}
	}
	s.NotFound(ctx)
}

func Assert(guard bool, errMsg string) {
	if !guard {
		panic(errMsg)
	}
}
