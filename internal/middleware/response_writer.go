package middleware

import "net/http"

type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func NewResponseWriterInterceptor(w http.ResponseWriter) *responseWriterInterceptor {
	return &responseWriterInterceptor{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rwi *responseWriterInterceptor) WriteHeader(code int) {
	if !rwi.wroteHeader {
		rwi.statusCode = code
		rwi.wroteHeader = true
		rwi.ResponseWriter.WriteHeader(code)
	}
}

func (rwi *responseWriterInterceptor) Write(b []byte) (int, error) {
	if !rwi.wroteHeader {
		rwi.WriteHeader(http.StatusOK)
	}
	return rwi.ResponseWriter.Write(b)
}
