package gweb

import "bytes"

type Option func(s *Server)

func RedirectTrailingSlashOption(redirectTrailingSlash bool) Option {
	return func(s *Server) {
		s.RedirectTrailingSlash = redirectTrailingSlash
	}
}

func RedirectFixedPathOption(redirectFixedPath bool) Option {
	return func(s *Server) {
		s.RedirectFixedPath = redirectFixedPath
	}
}

func HandleOptionsMethodOption(handleOPTIONS bool) Option {
	return func(s *Server) {
		s.HandleOPTIONS = handleOPTIONS
	}
}

func MethodNotAllowedOption(handleMethodNotAllowed bool, handler Handler) Option {
	return func(s *Server) {
		if handleMethodNotAllowed && handler != nil {
			s.HandleMethodNotAllowed = true
			s.MethodNotAllowed = handler
		} else {
			s.HandleMethodNotAllowed = false
			s.MethodNotAllowed = nil
		}
	}
}

func NotFoundOption(handler Handler) Option {
	return func(s *Server) {
		s.NotFound = handler
	}
}

func PanicHandlerOption(handler func(ctx *Context, err interface{})) Option {
	return func(s *Server) {
		s.PanicHandler = handler
	}
}

func LogoOption(printLogo bool, logo ...string) Option {
	buf := bytes.Buffer{}
	for _, s := range logo {
		buf.WriteString(s)
		buf.WriteByte('\n')
	}
	return func(s *Server) {
		s.PrintLogo = printLogo
		if !printLogo {
			s.Logo = ""
		} else {
			s.Logo = buf.String()
		}
	}
}

func NameOption(name string) Option {
	return func(s *Server) {
		s.name = name
	}
}
