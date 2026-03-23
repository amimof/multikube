// Package events provides interfaces and types for working with events
package events

import (
	"context"
	"maps"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventv1 "github.com/amimof/multikube/api/event/v1"
	"github.com/amimof/multikube/pkg/labels"
)

const (
	CientConnect  = eventv1.Event_EVENT_CLIENT_CONNECT
	BackendCreate = eventv1.Event_EVENT_BACKEND_CREATE
	BackendDelete = eventv1.Event_EVENT_BACKEND_DELETE
	BackendUpdate = eventv1.Event_EVENT_BACKEND_UPDATE
	BackendPatch  = eventv1.Event_EVENT_BACKEND_PATCH
)

type Subscriber interface {
	Subscribe(context.Context, ...eventv1.Event) chan *eventv1.Envelope
	Unsubscribe(context.Context, eventv1.Event) error
}

type Publisher interface {
	Publish(context.Context, *eventv1.Envelope) error
}

type Forwarder interface {
	Forward(context.Context, *eventv1.Envelope) error
}

type Object protoreflect.ProtoMessage

func NewRequest(evType eventv1.Event, obj Object, labels ...map[string]string) *eventv1.PublishRequest {
	return &eventv1.PublishRequest{
		Event: NewEvent(evType, obj, labels...),
	}
}

func NewEvent(evType eventv1.Event, obj Object, eventLabels ...map[string]string) *eventv1.Envelope {
	// Merge the maps
	l := labels.New()
	for _, label := range eventLabels {
		maps.Copy(l, label)
	}
	o, _ := anypb.New(obj)

	return &eventv1.Envelope{
		Event:     evType,
		Timestamp: timestamppb.Now(),
		Object:    o,
	}
}
