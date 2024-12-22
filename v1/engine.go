package httpbara

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// core is the main engine implementation that consolidates routes, groups, and middleware
// extracted by Handler instances. It sets up the Gin engine, applies all handlers, and runs the server.
// The 'core' type implements the Engine interface.
//
// Fields:
// - Params: A configuration object containing parameters for the engine (implementation details omitted).
// - flatGroups: A map of group names to Group objects. Each Group represents a set of related routes sharing a common prefix and middlewares.
// - flatMiddlewares: A map of middleware names to Middleware objects. Each middleware can also apply additional middleware.
// - flatRoutes: A slice of Route objects representing all routes extracted from Handler instances.
type core struct {
	Params

	flatGroups      map[string]*Group
	flatMiddlewares map[string]*Middleware
	flatRoutes      []*Route
}

// Engine defines the interface for an HTTP engine capable of registering routes, groups, and middleware
// and running the server. Implementations should integrate with a Gin engine.
//
// Methods:
// - flatHandlers([]*Handler): Process a collection of Handler objects to flatten their routes, groups, and middleware.
// - applyHandlers(): Apply all collected routes, groups, and middleware to the underlying Gin engine.
// - Run(addr string) chan error: Run the HTTP server at the specified address and return a channel for errors.
type Engine interface {
	flatHandlers(handlers []*Handler)
	applyHandlers()
	Run(addr string) error
}

// New creates a new Engine (core implementation) given a list of Handler objects
// and optional parameters. Handlers usually come from the logic of `NewHandler` (not shown here),
// which reflectively extracts routes, groups, and middleware from a user-defined struct.
//
// Parameters:
// - handlers: A slice of Handler objects containing routes, groups, and middleware definitions.
// - opts: A list of options (param functions) that configure the engine (e.g., setting a custom Gin instance or logger).
//
// Returns:
// - Engine: a configured engine ready to register routes and run.
// - error: If any configuration step fails, an error is returned.
//
// Example usage:
// ```go
// engine, err := New([]*Handler{handler1, handler2}, WithCustomLogger(myLogger))
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// err := engine.Run(":8080")
// // waiting for error
// ```
func New(handlers []*Handler, opts ...paramsCb) (Engine, error) {
	c := &core{
		flatGroups:      make(map[string]*Group),
		flatMiddlewares: make(map[string]*Middleware),
		flatRoutes:      make([]*Route, 0),
	}

	c.Params.shutdownTimeout = 30 * time.Second

	for _, opt := range opts {
		err := opt(&c.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Create a base Gin engine if none was provided
	if c.gin == nil {
		err := c.createBaseGin()
		if err != nil {
			return nil, fmt.Errorf("failed to create base gin engine: %w", err)
		}
	}

	// Set a default logger if none provided
	if c.log == nil {
		c.log = NewFmtLogger()
	}

	c.flatHandlers(handlers)
	c.applyHandlers()

	return c, nil
}

// flatHandlers processes the given list of Handler objects and flattens their groups, middlewares, and routes
// into core's internal maps and slices. This prepares all routing information to be later applied to the Gin engine.
//
// Parameters:
// - handlers: A slice of Handler objects, each containing discovered routes, groups, and middleware.
//
// After this method is called, `flatGroups`, `flatMiddlewares`, and `flatRoutes` will be populated.
func (c *core) flatHandlers(handlers []*Handler) {
	for _, handler := range handlers {
		c.flatRoutes = append(c.flatRoutes, handler.routes...)

		for _, group := range handler.groups {
			c.flatGroups[group.name] = group
		}

		for _, middleware := range handler.middlewares {
			c.flatMiddlewares[middleware.middleware] = middleware
		}
	}
}

// applyHandlers goes through all flattened routes and applies them to the Gin engine.
// It reconstructs the full path by combining group prefixes (if any) and sets up the middleware stack.
// Middleware can be defined at the group level and at the route level. If a route belongs to a group,
// the group's middleware is applied first, followed by the route's middleware.
//
// This method also logs warnings if a specified group or middleware cannot be found,
// and logs info messages about successful route registrations.
func (c *core) applyHandlers() {
	for _, route := range c.flatRoutes {
		path := route.path
		handleStack := make([]gin.HandlerFunc, 0)
		for _, mw := range c.rootMiddlewares {
			handleStack = append(handleStack, mw.handler)
		}

		// Apply group prefix and group-level middleware if route has a group
		if route.group != "" {
			if group, ok := c.flatGroups[route.group]; ok {
				path = strings.TrimSuffix(group.Path, "/") + "/" + strings.TrimPrefix(path, "/")

				for _, m := range group.middlewares {
					if mw, mwOk := c.flatMiddlewares[m]; mwOk {
						handleStack = append(handleStack, mw.handler)
					} else {
						c.log.Warn("skipping group middleware because there is no middleware with this name",
							"middlewareToSkip", m,
							"group", route.group,
						)
					}
				}
			} else {
				c.log.Warn("skipping group because group was not found",
					"path", route.path,
					"group", route.group,
				)
			}
		}

		var appliedMiddlewares []string
		for _, middleware := range route.middlewares {
			if mw, ok := c.flatMiddlewares[middleware]; ok {
				appliedMiddlewares = append(appliedMiddlewares, mw.middleware)

				// Some middleware can apply additional middleware
				for _, m := range mw.middlewares {
					if mw2, mw2ok := c.flatMiddlewares[m]; mw2ok {
						handleStack = append(handleStack, mw2.handler)
					} else {
						c.log.Warn("skipping middleware of middleware because there is no middleware with this name",
							"route", path,
							"middlewareToSkip", m,
							"parentMiddleware", mw.middleware,
						)
					}
				}

				handleStack = append(handleStack, mw.handler)
			} else {
				c.log.Warn("skipping route middleware because there is no middleware with this name",
					"route", path,
					"middlewareToSkip", middleware,
				)
			}
		}

		handleStack = append(handleStack, route.handler)

		c.gin.Handle(route.method, path, handleStack...)

		c.log.Info("route was registered",
			"method", route.method,
			"route", path,
			"middlewares", appliedMiddlewares,
		)
	}
}

// createBaseGin initializes a new default Gin engine with standard middleware (like Recovery).
// If a custom Gin instance was not provided via parameters, this method ensures there's at least
// a basic setup to work with.
//
// Returns:
// - error: If initialization fails for some reason (unlikely).
func (c *core) createBaseGin() error {
	c.gin = gin.New()
	c.gin.Use(gin.Recovery())

	return nil
}

// Run starts the HTTP server on the given address using the underlying Gin engine.
// It returns a channel of errors, allowing the caller to handle any runtime server errors asynchronously.
//
// Parameters:
// - addr: The address to listen on, e.g., ":8080" for port 8080.
//
// Returns:
// - chan error: A channel that will receive any error that occurs when running the server.
//
// Example:
// ```go
// engine, _ := New(handlers)
// errChan := engine.Run(":8080")
//
//	if err := <-errChan; err != nil {
//	    log.Fatal("server error:", err)
//	}
//
// ```
func (c *core) Run(addr string) error {
	errChan := make(chan error)
	srv := &http.Server{
		Addr:    addr,
		Handler: c.gin,
	}

	go func() {
		errChan <- func() error {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}

			return nil
		}()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit

	c.log.Info("shutting down server", "signal", sig)

	ctx, cancel := context.WithTimeout(context.Background(), c.shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	if c.taskTracker != nil {
		if err := c.taskTracker.Shutdown(ctx); err != nil {
			return fmt.Errorf("task tracker shutdown failed: %w", err)
		}
	}

	return <-errChan
}