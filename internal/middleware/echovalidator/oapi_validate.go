// Package echovalidator provides an OpenAPI request validation middleware for Echo v5.
package echovalidator

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/c-mierez/godec/internal/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
	"github.com/labstack/echo/v5"
	echoMiddleware "github.com/labstack/echo/v5/middleware"
)

// Context keys used by the echovalidator middleware.
type contextKey string

// Context keys used by the echovalidator middleware.
const (
	EchoContextKey contextKey = "oapi-codegen/echo-context"
	UserDataKey    contextKey = "oapi-codegen/user-data"
)

// ErrorHandler is called when request validation fails.
type ErrorHandler func(c *echo.Context, err error) error

// MultiErrorHandler handles multiple validation errors from OpenAPI spec validation.
type MultiErrorHandler func(openapi3.MultiError) error

// Options configures the OpenAPI request validator middleware.
type Options struct {
	ErrorHandler          ErrorHandler
	Options               openapi3filter.Options
	ParamDecoder          openapi3filter.ContentParameterDecoder
	UserData              any
	Skipper               echoMiddleware.Skipper
	MultiErrorHandler     MultiErrorHandler
	SilenceServersWarning bool
}

// OapiRequestValidatorWithOptions returns an Echo middleware that validates requests against the OpenAPI spec.
func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) echo.MiddlewareFunc {
	if swagger.Servers != nil && (options == nil || !options.SilenceServersWarning) {
		slog.Warn("OapiRequestValidatorWithOptions called with an OpenAPI spec that has Servers set. This can cause unexpected host validation failures.")
	}

	router, err := legacyrouter.NewRouter(swagger)
	if err != nil {
		panic(err)
	}

	skipper := getSkipperFromOptions(options)
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if skipper(c) {
				return next(c)
			}

			err := ValidateRequestFromContext(c, router, options)
			if err != nil {
				if options != nil && options.ErrorHandler != nil {
					return options.ErrorHandler(c, err)
				}
				return err
			}

			return next(c)
		}
	}
}

// ValidateRequestFromContext validates the current request against the OpenAPI router.
func ValidateRequestFromContext(ctx *echo.Context, router routers.Router, options *Options) error {
	req := ctx.Request()
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		var routeErr *routers.RouteError
		if errors.As(err, &routeErr) {
			return echo.NewHTTPError(http.StatusNotFound, routeErr.Reason)
		}
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error validating route: %s", err.Error()))
	}

	validationInput := &openapi3filter.RequestValidationInput{
		Request:    req,
		PathParams: pathParams,
		Route:      route,
	}

	requestContext := context.WithValue(req.Context(), EchoContextKey, ctx)
	if options != nil {
		validationInput.Options = &options.Options
		validationInput.ParamDecoder = options.ParamDecoder
		requestContext = context.WithValue(requestContext, UserDataKey, options.UserData)
	}

	err = openapi3filter.ValidateRequest(requestContext, validationInput)
	if err != nil {
		me := openapi3.MultiError{}
		if errors.As(err, &me) {
			return getMultiErrorHandlerFromOptions(options)(me)
		}

		var reqErr *openapi3filter.RequestError
		if errors.As(err, &reqErr) {
			errorLines := strings.Split(reqErr.Error(), "\n")
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: errorLines[0]}
		}

		var secErr *openapi3filter.SecurityRequirementsError
		if errors.As(err, &secErr) {
			var authErr *middleware.AuthError
			if errors.As(err, &authErr) {
				return authErr
			}
			for _, e := range secErr.Errors {
				var httpErr *echo.HTTPError
				if errors.As(e, &httpErr) {
					return httpErr
				}
			}
			return &echo.HTTPError{Code: http.StatusForbidden, Message: secErr.Error()}
		}

		return &echo.HTTPError{Code: http.StatusInternalServerError, Message: fmt.Sprintf("error validating request: %s", err)}
	}

	ctx.SetRequest(validationInput.Request)
	return nil
}

// GetEchoContext retrieves the Echo context from a standard context.Context.
func GetEchoContext(c context.Context) *echo.Context {
	iface := c.Value(EchoContextKey)
	if iface == nil {
		return nil
	}
	eCtx, ok := iface.(*echo.Context)
	if !ok {
		return nil
	}
	return eCtx
}

// GetUserData retrieves the user data set in Options from a standard context.Context.
func GetUserData(c context.Context) any {
	return c.Value(UserDataKey)
}

func getSkipperFromOptions(options *Options) echoMiddleware.Skipper {
	if options == nil || options.Skipper == nil {
		return echoMiddleware.DefaultSkipper
	}

	return options.Skipper
}

func getMultiErrorHandlerFromOptions(options *Options) MultiErrorHandler {
	if options == nil || options.MultiErrorHandler == nil {
		return defaultMultiErrorHandler
	}

	return options.MultiErrorHandler
}

func defaultMultiErrorHandler(me openapi3.MultiError) error {
	return &echo.HTTPError{Code: http.StatusBadRequest, Message: me.Error()}
}
