package echovalidator

import (
	"context"
	"errors"
	"fmt"
	"log"
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

const (
	EchoContextKey = "oapi-codegen/echo-context"
	UserDataKey    = "oapi-codegen/user-data"
)

type ErrorHandler func(c *echo.Context, err error) error
type MultiErrorHandler func(openapi3.MultiError) error

type Options struct {
	ErrorHandler          ErrorHandler
	Options               openapi3filter.Options
	ParamDecoder          openapi3filter.ContentParameterDecoder
	UserData              any
	Skipper               echoMiddleware.Skipper
	MultiErrorHandler     MultiErrorHandler
	SilenceServersWarning bool
}

func OapiRequestValidatorWithOptions(swagger *openapi3.T, options *Options) echo.MiddlewareFunc {
	if swagger.Servers != nil && (options == nil || !options.SilenceServersWarning) {
		log.Println("WARN: OapiRequestValidatorWithOptions called with an OpenAPI spec that has Servers set. This can cause unexpected host validation failures.")
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

func ValidateRequestFromContext(ctx *echo.Context, router routers.Router, options *Options) error {
	req := ctx.Request()
	route, pathParams, err := router.FindRoute(req)
	if err != nil {
		switch e := err.(type) {
		case *routers.RouteError:
			return echo.NewHTTPError(http.StatusNotFound, e.Reason)
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("error validating route: %s", err.Error()))
		}
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

		switch e := err.(type) {
		case *openapi3filter.RequestError:
			errorLines := strings.Split(e.Error(), "\n")
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: errorLines[0]}
		case *openapi3filter.SecurityRequirementsError:
			var authErr *middleware.AuthError
			if errors.As(e, &authErr) {
				return authErr
			}

			for _, secErr := range e.Errors {
				httpErr, ok := secErr.(*echo.HTTPError)
				if ok {
					return httpErr
				}
			}

			return &echo.HTTPError{Code: http.StatusForbidden, Message: e.Error()}
		default:
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: fmt.Sprintf("error validating request: %s", err)}
		}
	}

	ctx.SetRequest(validationInput.Request)
	return nil
}

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
