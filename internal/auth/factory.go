package auth

import (
	"errors"
	"reflect"
)

// ErrUnknownServiceType is returned when trying to create a middleware for an unknown service type
var ErrUnknownServiceType = errors.New("unknown auth service type")

// MiddlewareFactory is a function that creates a middleware for a service
type MiddlewareFactory func(Service) Middleware

var middlewareFactories = make(map[string]MiddlewareFactory)

// RegisterMiddlewareFactory registers a factory function for creating middleware for a specific service type
func RegisterMiddlewareFactory(serviceType reflect.Type, factory MiddlewareFactory) {
	middlewareFactories[serviceType.String()] = factory
}

// NewMiddleware creates an appropriate middleware for the given service
// It returns ErrUnknownServiceType if it doesn't know how to create middleware for the service
func NewMiddleware(service Service) (Middleware, error) {
	serviceType := reflect.TypeOf(service)

	// Try to find a factory for this service type
	if factory, ok := middlewareFactories[serviceType.String()]; ok {
		return factory(service), nil
	}

	return nil, ErrUnknownServiceType
}
