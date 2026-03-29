package events

import (
	"context"
	"runtime"
	"strings"

	"github.com/amimof/multikube/pkg/logger"
	"google.golang.org/protobuf/proto"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	eventsv1 "github.com/amimof/multikube/api/event/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
)

type Handler interface {
	Handle(context.Context, *eventsv1.Envelope) error
}

type (
	HandlerFunc        func(context.Context, *eventsv1.Envelope) error
	BackendHandlerFunc func(context.Context, *backendv1.Backend) error
	RouteHandlerFunc   func(context.Context, *routev1.Route) error
	PolicyHandlerFunc  func(context.Context, *policyv1.Policy) error
)

// getCallerInfo gets the file, line, and function name of the caller
func getCallerInfo(skip int) (string, int, string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown_file", 0, "unknown_func"
	}

	// Get function name
	funcName := runtime.FuncForPC(pc).Name()
	funcName = trimFunctionName(funcName)

	// Trim the file path to only the base name
	fileParts := strings.Split(file, "/")
	file = fileParts[len(fileParts)-1]

	return file, line, funcName
}

func trimFunctionName(funcName string) string {
	funcParts := strings.Split(funcName, "/")
	return funcParts[len(funcParts)-1]
}

func HandleErrors(log logger.Logger, h HandlerFunc) HandlerFunc {
	return func(ctx context.Context, ev *eventsv1.Envelope) error {
		err := h(ctx, ev)
		if err != nil {
			_, _, funcName := getCallerInfo(2)
			log.Error("handler returned error", "error", err, "event", ev.GetEvent().String(), "handler", funcName)
			return err
		}
		return nil
	}
}

func Handle(h HandlerFunc) HandlerFunc {
	return func(ctx context.Context, ev *eventsv1.Envelope) error {
		return h(ctx, ev)
	}
}

func HandleNew[T any, PT interface {
	*T
	proto.Message
}](h func(context.Context, PT) error) HandlerFunc {
	return func(ctx context.Context, ev *eventsv1.Envelope) error {
		var obj T
		if err := ev.GetObject().UnmarshalTo(PT(&obj)); err != nil {
			return err
		}
		return h(ctx, PT(&obj))
	}
}

func NewHandler[T any, PT interface {
	*T
	proto.Message
}](
	h func(context.Context, PT) error,
) func(context.Context, proto.Message) error {
	return func(ctx context.Context, m proto.Message) error {
		return h(ctx, m.(PT)) // single type assertion, in one place
	}
}

func HandleBackends(h ...BackendHandlerFunc) HandlerFunc {
	return func(ctx context.Context, ev *eventsv1.Envelope) error {
		for _, ih := range h {
			var backend backendv1.Backend
			err := ev.GetObject().UnmarshalTo(&backend)
			if err != nil {
				return err
			}
			if err := ih(ctx, &backend); err != nil {
				return err
			}
		}
		return nil
	}
}

func HandleRoutes(h ...RouteHandlerFunc) HandlerFunc {
	return func(ctx context.Context, ev *eventsv1.Envelope) error {
		for _, ih := range h {
			var route routev1.Route
			err := ev.GetObject().UnmarshalTo(&route)
			if err != nil {
				return err
			}
			if err := ih(ctx, &route); err != nil {
				return err
			}
		}
		return nil
	}
}

func HandlePolicies(h ...PolicyHandlerFunc) HandlerFunc {
	return func(ctx context.Context, ev *eventsv1.Envelope) error {
		for _, ih := range h {
			var policy policyv1.Policy
			err := ev.GetObject().UnmarshalTo(&policy)
			if err != nil {
				return err
			}
			if err := ih(ctx, &policy); err != nil {
				return err
			}
		}
		return nil
	}
}
