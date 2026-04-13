package handler

import (
	"github.com/shrinishv/taskflow-shrinish/backend/internal/httputil"
)

var (
	JSON            = httputil.JSON
	ValidationError = httputil.ValidationError
	ErrorResponse   = httputil.ErrorResponse
	WriteError      = httputil.WriteError
)
