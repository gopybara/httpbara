package httpbara

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"reflect"
	"regexp"
	"strings"
)

const (
	// MiddlewareTag is a struct tag key used to specify a single middleware name.
	MiddlewareTag = "middleware"

	// MiddlewaresTag is a struct tag key used to specify a comma-separated list of middleware names.
	MiddlewaresTag = "middlewares"

	// GroupTag is a struct tag key used to specify the group path prefix for routes.
	GroupTag = "group"

	// RouteTag is a struct tag key used to define the route's HTTP method and path.
	RouteTag = "route"
)

// Handler processes a given handler struct to extract and configure routes, groups, and middlewares.
// It uses reflection to parse struct tags and associates them with the actual Gin handler functions.
//
// The workflow is as follows:
// 1. Recursively scan the provided struct (including embedded and nested structs) for fields of type `Route`, `Group`, and `Middleware`.
// 2. Extract corresponding tags (e.g. `route:"POST /checkout/apply"`) to determine routes, their HTTP methods, paths, and middleware associations.
// 3. Match field names with struct methods (signature `func(*gin.Context)`) to create fully configured routes and middleware.
// 4. Store all parsed routes, groups, and middleware for later registration in a Gin router.
//
// **Example:**
//
// Consider a scenario in an online store application. We might have a versioned API group (`/api/v3`) and routes for products, cart, and checkout processes.
//
// For example, define a group for version 3 of our API:
//
// ```go
//
//	type IV3Group struct {
//	    V3 Group `group:"/api/v3"`
//	}
//
//	type V3GroupImpl struct {
//	    IV3Group
//	}
//
// ```
//
// Now, inside this versioned group, we could define product routes:
//
// ```go
//
//	type IProductRoutes struct {
//	    ListProducts Route `route:"GET /products" middlewares:"auth,logging" group:"v3"`
//	    GetProduct   Route `route:"GET /products/:id" group:"v3"`
//	}
//
//	type ProductRoutesImpl struct {
//	    IProductRoutes
//	}
//
//	func (p *ProductRoutesImpl) ListProducts(ctx *gin.Context) {
//	    // Handler logic for listing products
//	    // e.g., ctx.JSON(http.StatusOK, productList)
//	}
//
//	func (p *ProductRoutesImpl) GetProduct(ctx *gin.Context) {
//	    // Handler logic for getting a single product by ID
//	    // e.g., product := findProductByID(ctx.Param("id"))
//	    // ctx.JSON(http.StatusOK, product)
//	}
//
// ```
//
// We can also define middleware and checkout routes that use these middlewares:
//
// ```go
//
//	type ICheckoutRouter struct {
//	    ApplyCart Route `route:"POST /checkout/apply" middlewares:"log,cors,analytics" group:"v3"`
//
//	    LogMiddleware       Middleware `middleware:"log"`
//	    CorsMiddleware      Middleware `middleware:"cors"`
//	    AnalyticsMiddleware Middleware `middleware:"analytics"`
//	}
//
//	type CheckoutRouterImpl struct {
//	    ICheckoutRouter
//	}
//
//	func (c *CheckoutRouterImpl) ApplyCart(ctx *gin.Context) {
//	    // Handler logic for applying cart items to the checkout process
//	    // e.g., ctx.JSON(http.StatusOK, checkoutResponse)
//	}
//
// // Middleware example methods:
//
//	func (c *CheckoutRouterImpl) LogMiddleware(ctx *gin.Context) {
//	    // Logging middleware logic
//	    ctx.Next()
//	}
//
//	func (c *CheckoutRouterImpl) CorsMiddleware(ctx *gin.Context) {
//	    // CORS middleware logic (set headers, etc.)
//	    ctx.Next()
//	}
//
//	func (c *CheckoutRouterImpl) AnalyticsMiddleware(ctx *gin.Context) {
//	    // Analytics middleware logic (track requests)
//	    ctx.Next()
//	}
//
// ```
//
// By creating a `Handler` with these structures, you can automatically discover and set up all routes, groups, and middleware:
//
// ```go
// handler := NewHandler(&CheckoutRouterImpl{})
// // handler now includes all defined routes and associated middleware within the V3 group.
// ```
type Handler struct {
	routes       []*Route
	casualRoutes []*casualRoute

	groups      []*Group
	middlewares []*Middleware
}

// AsHandler creates a new Handler by analyzing the provided `handlerStruct`.
// It recursively extracts all methods and fields, builds routes, groups, and middleware,
// and returns a fully initialized Handler instance.
//
// **Example:**
// Given `CheckoutRouterImpl` and `ProductRoutesImpl` as in the examples above, you could do:
//
// ```go
// handler := AsHandler(&ProductRoutesImpl{})
// // The handler now holds routes like GET /api/v3/products (with auth, logging middleware),
// // and GET /api/v3/products/:id, ready to be registered in your Gin engine.
// ```
func AsHandler(handlerStruct interface{}) (*Handler, error) {
	handler := &Handler{}

	ginHandlers, casualHandlers := handler.getAllGinHandlers(reflect.ValueOf(handlerStruct))
	flatFields := handler.getAllReflectionFieldsRecursive(reflect.ValueOf(handlerStruct))

	err := handler.searchForGroups(flatFields)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to search for groups: %w",
			err,
		)
	}

	handler.searchForMiddlewares(flatFields, ginHandlers)

	err = handler.searchForRoutes(flatFields, ginHandlers, casualHandlers)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to search for routes: %w",
			err,
		)
	}

	return handler, nil
}

// searchForRoutes looks for fields of type `Route` in the given fields. It parses their tags to determine
// the HTTP method, path, associated middlewares, and optionally their group. It then creates `Route` objects.
//
// Each route field must correspond to a handler method on the same struct implementing `func(*gin.Context)`.
//
// **Example:**
// ```go
// ListProducts Route `route:"GET /products" middlewares:"auth,logging" group:"v3"`
// ```
// This defines a GET route at `/api/v3/products` (because of group "v3"), with middleware "auth" and "logging".
func (h *Handler) searchForRoutes(flatFields []reflect.StructField, foundHandlers map[string]gin.HandlerFunc, foundCasualHandlers map[string]*casualHandler) error {
	typeOfRoute := reflect.TypeOf(Route{})
	var err error
	routes := make([]*Route, 0)
	casualRoutes := make([]*casualRoute, 0)

	for _, fieldType := range flatFields {
		if fieldType.Type != typeOfRoute {
			continue
		}

		if foundHandlers[fieldType.Name] != nil {
			route := &Route{
				handler:     foundHandlers[fieldType.Name],
				middlewares: h.parseMiddlewaresTag(fieldType.Tag.Get(MiddlewaresTag)),
				group:       fieldType.Tag.Get(GroupTag),
			}

			route.method, route.path, err = h.parseRouteTag(fieldType.Tag.Get(RouteTag))
			if err != nil {
				return fmt.Errorf("failed to parse route tag: %w", err)
			}

			routes = append(routes, route)
		} else if foundCasualHandlers[fieldType.Name] != nil {
			route := &casualRoute{
				handler:     foundCasualHandlers[fieldType.Name],
				middlewares: h.parseMiddlewaresTag(fieldType.Tag.Get(MiddlewaresTag)),
				group:       fieldType.Tag.Get(GroupTag),
			}

			route.method, route.path, err = h.parseRouteTag(fieldType.Tag.Get(RouteTag))
			if err != nil {
				return fmt.Errorf("failed to parse route tag: %w", err)
			}

			casualRoutes = append(casualRoutes, route)
		}
	}

	h.routes = routes
	h.casualRoutes = casualRoutes

	return nil
}

// parseRouteTag parses a route tag which should be in the format: "METHOD /path".
// For example: "POST /checkout/apply".
// It returns the extracted HTTP method and path, or an error if the format is invalid.
func (h *Handler) parseRouteTag(tag string) (method string, path string, err error) {
	re := regexp.MustCompile(`(?i)^([A-Z]{3,10}) (.*)$`)
	matches := re.FindStringSubmatch(tag)
	if len(matches) != 3 {
		return "", "", errors.New("invalid route tag")
	}

	return matches[1], matches[2], nil
}

// searchForMiddlewares finds fields of type `Middleware`, parses their tags,
// and constructs `Middleware` objects. The `middleware` tag defines a single middleware name,
// while the `middlewares` tag can define multiple middleware names that this middleware will apply.
//
// Handlers for middlewares are methods with the same name as the field, but following `func(*gin.Context)` signature.
//
// **Example:**
// ```go
// LogMiddleware Middleware `middleware:"log"`
// CorsMiddleware Middleware `middleware:"cors"`
// AnalyticsMiddleware Middleware `middleware:"analytics"`
// ```
//
// Each middleware can be referenced by routes through the `middlewares:"..."` tag.
func (h *Handler) searchForMiddlewares(flatFields []reflect.StructField, foundHandlers map[string]gin.HandlerFunc) {
	typeOfMiddleware := reflect.TypeOf(Middleware{})
	middlewares := make([]*Middleware, 0)

	for _, fieldType := range flatFields {
		if fieldType.Type != typeOfMiddleware {
			continue
		}

		if foundHandlers[fieldType.Name] != nil {
			middlewareName := fieldType.Tag.Get(MiddlewareTag)
			if middlewareName == "" {
				middlewareName = fieldType.Name
			}

			m := &Middleware{
				handler:     foundHandlers[fieldType.Name],
				middleware:  strings.ToLower(middlewareName),
				middlewares: h.parseMiddlewaresTag(fieldType.Tag.Get(MiddlewaresTag)),
			}

			middlewares = append(middlewares, m)
		}
	}

	h.middlewares = middlewares
}

// getAllGinHandlers scans the given reflected value (struct) for methods
// that match the signature `func(*gin.Context)` and returns them in a map keyed by method name.
// These methods can be route handlers or middleware handlers.
func (h *Handler) getAllGinHandlers(rv reflect.Value) (map[string]gin.HandlerFunc, map[string]*casualHandler) {
	rt := rv.Type()
	handlers := make(map[string]gin.HandlerFunc)
	casualHandlers := make(map[string]*casualHandler)

	for i := 0; i < rt.NumMethod(); i++ {
		method := rt.Method(i)

		if isSimpleGinHandler(method.Type) {
			handlers[method.Name] = rv.Method(i).Interface().(func(*gin.Context))
		} else if isCasualHandler(method.Type) {
			casualHandlers[method.Name] = &casualHandler{
				rv: &rv,
				rm: &method,
			}
		}
	}

	return handlers, casualHandlers
}

// searchForGroups finds fields of type `Group`, parses the `group` tag to identify the path prefix,
// and constructs `Group` objects. A group can also have middleware specified via the `middlewares` tag.
//
// Groups help organize routes under a common path prefix and apply shared middleware.
//
// **Example:**
// ```go
//
//	type IV3Group struct {
//	    V3 Group `group:"/api/v3"`
//	}
//
//	type V3GroupImpl struct {
//	    IV3Group
//	}
//
// ```
//
// This creates a group named "v3" with a path prefix "/api/v3". Routes referencing `group:"v3"` will be placed under `/api/v3`.
func (h *Handler) searchForGroups(flatFields []reflect.StructField) error {
	typeOfGroup := reflect.TypeOf(Group{})
	groups := make([]*Group, 0)

	for _, field := range flatFields {
		if field.Type != typeOfGroup {
			continue
		}

		groupTagValue := field.Tag.Get(GroupTag)
		if groupTagValue != "" {
			group, err := h.parseGroupTag(&parseGroupTagRequest{
				tagValue: groupTagValue,
				handler:  field.Name,
				field:    field.Name,
			})
			if err != nil {
				return err
			}

			middlewaresTagValue := field.Tag.Get(MiddlewaresTag)
			if middlewaresTagValue != "" {
				group.middlewares = h.parseMiddlewaresTag(middlewaresTagValue)
			}

			groups = append(groups, group)
		}
	}

	h.groups = groups
	return nil
}

// getAllReflectionFieldsRecursive recursively extracts all fields (including those from embedded and nested structs)
// from the given reflected value.
func (h *Handler) getAllReflectionFieldsRecursive(rv reflect.Value) []reflect.StructField {
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	rt := rv.Type()
	fields := make([]reflect.StructField, 0)

	for i := 0; i < rv.NumField(); i++ {
		if rt.Field(i).Type.Kind() == reflect.Struct {
			fields = append(fields, h.getAllReflectionFieldsRecursive(rv.Field(i))...)
		}
		fields = append(fields, rt.Field(i))
	}

	return fields
}

// parseMiddlewaresTag splits a comma-separated list of middleware names from a struct tag,
// trims spaces, converts them to lowercase, and returns them as a slice of strings.
func (h *Handler) parseMiddlewaresTag(tag string) []string {
	result := make([]string, 0)

	values := strings.Split(tag, ",")
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			result = append(result, strings.ToLower(v))
		}
	}

	return result
}

// parseGroupTagRequest holds data required to parse a group tag from a struct field.
type parseGroupTagRequest struct {
	tagValue string
	handler  string
	field    string
}

// parseGroupTag parses the group tag value, sets the group's name based on the field name (alphanumeric and underscore characters),
// converts it to lowercase, and sets the path prefix from the tag value.
//
// **Example:**
// Given a field defined as:
// ```go
//
//	type IV3Group struct {
//	    V3 Group `group:"/api/v3"`
//	}
//
//	type V3GroupImpl struct {
//	    IV3Group
//	}
//
// ```
// The resulting group has the name "v3" (derived from "V3") and the path "/api/v3".
func (h *Handler) parseGroupTag(req *parseGroupTagRequest) (*Group, error) {
	group := Group{}

	for _, char := range req.field {
		if (char >= 'A' && char <= 'Z') ||
			(char >= 'a' && char <= 'z') ||
			(char >= '0' && char <= '9') ||
			char == '_' {
			group.name += string(char)
		}
	}

	group.name = strings.ToLower(group.name)
	group.Path = req.tagValue

	return &group, nil
}

// Route defines an HTTP endpoint with a method, path, associated handler, and optional middlewares or group prefix.
//
// Fields:
// - `method`: The HTTP method (e.g., "GET", "POST").
// - `path`: The HTTP path (e.g., "/checkout/apply").
// - `handler`: The Gin handler function that processes the request.
// - `middlewares`: A list of middleware names applied before the handler.
// - `group`: The name of the group this route belongs to, if any.
//
// **Example:**
// ```go
//
//	type IProductRoutes struct {
//	    ListProducts Route `route:"GET /products" middlewares:"auth,logging" group:"v3"`
//	}
//
// ```
// This defines a GET route at `/api/v3/products` that applies "auth" and "logging" middleware.
type Route struct {
	middlewares []string
	group       string
	method      string
	path        string
	handler     gin.HandlerFunc
}

// Middleware defines a middleware associated with a handler function and possibly other nested middlewares.
//
// Fields:
// - `middleware`: The primary middleware name (from `middleware:"name"` tag or derived from the field name).
// - `middlewares`: A list of additional middleware names that this middleware applies internally.
// - `handler`: The Gin handler function for the middleware.
//
// **Example:**
// ```go
//
//	type ICheckoutRouter struct {
//	    ApplyCart Route `route:"POST /checkout/apply" middlewares:"log,cors,analytics" group:"v3"`
//
//	    LogMiddleware       Middleware `middleware:"log"`
//	    CorsMiddleware      Middleware `middleware:"cors"`
//	    AnalyticsMiddleware Middleware `middleware:"analytics"`
//	}
//
// ```
//
// Here, the `ApplyCart` route will be executed with the "log", "cors", and "analytics" middleware in the defined order.
type Middleware struct {
	handler     gin.HandlerFunc
	middleware  string
	middlewares []string
}

// Group defines a group of routes that share a common path prefix and possibly a set of middlewares.
//
// Fields:
// - `name`: The group's name, derived from the field name (e.g., "v3" from "V3").
// - `Path`: The prefix path for all routes in this group (e.g., "/api/v3").
// - `Middlewares`: A list of middleware names applied to all routes in the group.
//
// **Example:**
//
// ```go
//
//	type IV3Group struct {
//	    V3 Group `group:"/api/v3"`
//	}
//
//	type V3GroupImpl struct {
//	    IV3Group
//	}
//
// ```
//
// All routes referencing `group:"v3"` will be placed under "/api/v3".
type Group struct {
	name        string
	Path        string
	middlewares []string
}

func isSimpleGinHandler(t reflect.Type) bool {
	return t.NumIn() == 2 &&
		t.NumOut() == 0 &&
		t.In(1) == reflect.TypeOf((*gin.Context)(nil))
}
