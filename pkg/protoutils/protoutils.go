package protoutils

import (
	"fmt"
	"reflect"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func listEqual(a, b protoreflect.List) bool {
	if a.Len() != b.Len() {
		return false
	}
	for i := 0; i < a.Len(); i++ {
		if a.Get(i).Interface() != b.Get(i).Interface() {
			return false
		}
	}
	return true
}

func mapEqual(a, b protoreflect.Map) bool {
	if a.Len() != b.Len() {
		return false
	}
	equal := true
	a.Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
		if bv := b.Get(k); bv.IsValid() {
			equal = v.Interface() == bv.Interface()
		} else {
			equal = false
		}
		return equal
	})
	return equal
}

// ApplyFieldMaskToNewMessage creates a new message containing only the fields specified in the FieldMask.
// If mask is nil then source is returned in its original unalterned state
func ApplyFieldMaskToNewMessage(source proto.Message, mask *fieldmaskpb.FieldMask) (proto.Message, error) {
	if mask == nil {
		return source, nil
	}

	newMessage := proto.Clone(source)
	newMessageProto := newMessage.ProtoReflect()
	sourceProto := source.ProtoReflect()

	// Clear all fields initially
	newMessageProto.Range(func(field protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		newMessageProto.Clear(field)
		return true
	})

	// Apply each field specified in the FieldMask
	for _, path := range mask.Paths {
		err := ApplyNestedField(newMessageProto, sourceProto, strings.Split(path, "."))
		if err != nil {
			return nil, err
		}
	}

	return newMessage, nil
}

// ApplyNestedField sets the value of a nested field in the target message based on the source message.
func ApplyNestedField(target, source protoreflect.Message, path []string) error {
	if len(path) == 0 {
		return nil
	}

	// Look up the field descriptor for the current level in the path
	field := source.Descriptor().Fields().ByName(protoreflect.Name(path[0]))
	if field == nil {
		return fmt.Errorf("field %q not found in target message", path[0])
	}

	// If we are at the final field in the path, set it directly
	if len(path) == 1 {
		if source.Has(field) {
			target.Set(field, source.Get(field))
		}
		return nil
	}

	// Recurse into the nested message
	if field.Message() == nil {
		return fmt.Errorf("field %q is not a message type", path[0])
	}

	// Ensure the target has an initialized message at this field
	if !target.Has(field) {
		target.Set(field, target.NewField(field))
	}

	// return ApplyNestedField(target.Mutable(field).Message(), source.Get(field).Message(), path[1:])
	r := ApplyNestedField(target.Mutable(field).Message(), source.Get(field).Message(), path[1:])
	return r
}

// GenerateFieldMask compares two protobuf messages and generates a FieldMask with changed fields.
func GenerateFieldMask(original, updated protoreflect.ProtoMessage) (*fieldmaskpb.FieldMask, error) {
	if original == nil || updated == nil {
		return nil, fmt.Errorf("both original and updated messages must be non-nil")
	}

	originalReflect := original.ProtoReflect()
	updatedReflect := updated.ProtoReflect()

	if originalReflect.Descriptor() != updatedReflect.Descriptor() {
		return nil, fmt.Errorf("messages must have the same descriptor")
	}

	paths := []string{}
	err := compareMessages(originalReflect, updatedReflect, "", &paths)
	if err != nil {
		return nil, err
	}

	return &fieldmaskpb.FieldMask{Paths: paths}, nil
}

// ClearProto resets fields on the proto message recursively
func ClearProto(in protoreflect.Message) {
	in.Range(func(fd protoreflect.FieldDescriptor, _ protoreflect.Value) bool {
		in.Clear(fd)
		return true
	})
	in.SetUnknown(nil)
}

func isWrapperField(fd protoreflect.FieldDescriptor) bool {
	if fd.Kind() != protoreflect.MessageKind {
		return false
	}
	switch string(fd.Message().FullName()) {
	case "google.protobuf.StringValue",
		"google.protobuf.Int32Value",
		"google.protobuf.Int64Value",
		"google.protobuf.UInt32Value",
		"google.protobuf.UInt64Value",
		"google.protobuf.BoolValue",
		"google.protobuf.FloatValue",
		"google.protobuf.DoubleValue",
		"google.protobuf.BytesValue":
		return true
	default:
		return false
	}
}

func unwrapWrapper(m protoreflect.Message) (any, bool) {
	// wrappers always have a single field named "value"
	fd := m.Descriptor().Fields().ByName("value")
	if fd == nil || !m.Has(fd) {
		return nil, false
	}
	return m.Get(fd).Interface(), true
}

func compareMessages(orig, upd protoreflect.Message, prefix string, paths *[]string) error {
	fields := orig.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)

		// Get field values from both original and updated messages
		origValue := orig.Get(field)
		updValue := upd.Get(field)

		// Build the current field path
		currentPath := field.Name()
		if prefix != "" {
			currentPath = protoreflect.Name(fmt.Sprintf("%s.%s", prefix, field.Name()))
		}

		// Get presence + values
		oHas := orig.Has(field)
		uHas := upd.Has(field)

		// Handle field types
		switch {
		case field.IsList():
			// Compare lists
			if !listEqual(origValue.List(), updValue.List()) {
				*paths = append(*paths, string(currentPath))
			}
		case field.IsMap():
			// Compare maps
			if !mapEqual(origValue.Map(), updValue.Map()) {
				*paths = append(*paths, string(currentPath))
			}
		case field.Kind() == protoreflect.MessageKind && isWrapperField(field):

			// Treat wrappers like scalars; never emit ".value"
			if oHas != uHas {
				*paths = append(*paths, string(currentPath))
				continue
			}
			ov, _ := unwrapWrapper(origValue.Message())
			uv, _ := unwrapWrapper(updValue.Message())
			if !reflect.DeepEqual(ov, uv) {
				*paths = append(*paths, string(currentPath))
			}
		case field.Kind() == protoreflect.MessageKind:
			switch {
			case !oHas && !uHas: // Nothing present on either side
				continue
			case !oHas && uHas:
				collectPresentLeafPaths(updValue.Message(), string(currentPath), paths)
			default:
				// both present - recurse
				if err := compareMessages(origValue.Message(), updValue.Message(), string(currentPath), paths); err != nil {
					return err
				}
			}
		default:
			// scalar / enum
			if oHas != uHas {
				// presence changed (set or cleared)
				*paths = append(*paths, string(currentPath))
				continue
			}
			if uHas && (origValue.Interface() != updValue.Interface()) {
				*paths = append(*paths, string(currentPath))
			}
		}
	}

	return nil
}

// Walks a message and appends leaf paths for all PRESENT fields.
func collectPresentLeafPaths(m protoreflect.Message, prefix string, paths *[]string) {
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		p := prefix + "." + string(fd.Name())

		switch {
		case fd.IsList(), fd.IsMap():
			*paths = append(*paths, p)
		case fd.Kind() == protoreflect.MessageKind && isWrapperField(fd):
			*paths = append(*paths, p)
		case fd.Kind() == protoreflect.MessageKind:
			// descend only into PRESENT submessages
			collectPresentLeafPaths(v.Message(), p, paths)
		default:
			*paths = append(*paths, p)
		}
		return true
	})
}

type MergeFunc[T any] func(item T) string

func MergeSlices[T any](base, patch []T, keyFunc MergeFunc[T], mergeItem func(base, patch T) T) []T {
	baseMap := make(map[string]T, len(base))
	order := make([]string, 0, len(base))
	for _, item := range base {
		key := keyFunc(item)
		baseMap[key] = item
		order = append(order, key)
	}

	// Process patch items.
	for _, p := range patch {
		key := keyFunc(p)
		if _, exists := baseMap[key]; exists {
			// Update the existing item.
			baseMap[key] = mergeItem(baseMap[key], p)
		} else {
			// Add new items and record their key order.
			baseMap[key] = p
			order = append(order, key)
		}
	}

	// Build the merged slice preserving the order.
	merged := make([]T, 0, len(order))
	for _, key := range order {
		merged = append(merged, baseMap[key])
	}

	return merged
}

// ClearRepeatedFields will clear fields thare are of kind list
// FIX: Doesn't work on map fields
func ClearRepeatedFields(msg proto.Message) {
	m := msg.ProtoReflect()
	m.Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
		switch {
		case fd.IsList():
			m.Clear(fd)
		case fd.IsMap():
			break
		case fd.Kind() == protoreflect.MessageKind && m.Has(fd):
			ClearRepeatedFields(v.Message().Interface())
		}
		return true
	})
}

// StrategicMerge merges patch into base strategically as defined by provided merge funcs.
// Non-repated fields are cleared recursively before mergefuncs are applied to avoid dupliced list elements.
// Does not currently support removal of list-elements.
func StrategicMerge[T proto.Message](base, patch T, mergeFuncs ...func(b, p T)) T {
	tmp := proto.Clone(base).(T)
	patchClone := proto.Clone(patch).(T)

	ClearRepeatedFields(patchClone)

	proto.Merge(tmp, patchClone)

	for _, mergeFunc := range mergeFuncs {
		if !reflect.ValueOf(patch).IsNil() {
			mergeFunc(tmp, patch)
		}
	}

	return tmp
}

// EnsureMessageField initializes field on m if nil, using reflection
func EnsureMessageField(m proto.Message, fieldName string) proto.Message {
	if m == nil {
		return m
	}

	am := m.ProtoReflect()

	fd := am.Descriptor().Fields().ByName(protoreflect.Name(fieldName))
	if fd == nil {
		return m
	}
	if fd.Kind() != protoreflect.MessageKind {
		return m
	}

	am.Mutable(fd)
	return nil
}

// SpecEqual compares a and b's config field using reflection. Returns true if they match
func SpecEqual(a, b proto.Message) (bool, error) {
	if a == nil || b == nil {
		return a == b, nil
	}

	am := a.ProtoReflect()
	bm := b.ProtoReflect()

	// Same message type is expected in an update.
	if am.Descriptor().FullName() != bm.Descriptor().FullName() {
		return false, fmt.Errorf("type mismatch: %s vs %s",
			am.Descriptor().FullName(), bm.Descriptor().FullName())
	}

	fd := am.Descriptor().Fields().ByName("config")
	if fd == nil {
		// If you *require* a spec field, return error instead.
		// If not, you can treat the entire message as “spec”.
		return proto.Equal(a, b), nil
	}

	av := am.Get(fd)
	bv := bm.Get(fd)

	// Handle missing/zero values.
	if !am.Has(fd) && !bm.Has(fd) {
		return true, nil
	}
	if !am.Has(fd) || !bm.Has(fd) {
		return false, nil
	}

	// If spec is a message, compare messages.
	if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
		return proto.Equal(av.Message().Interface(), bv.Message().Interface()), nil
	}

	// Otherwise compare scalar/list/map via value equality.
	// (protoreflect.Value is comparable only in some cases; safer to normalize)
	return av.Interface() == bv.Interface(), nil
}
