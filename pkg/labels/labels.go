package labels

import (
	"fmt"
	"slices"
)

const DefaultLabelPrefix = "voiyd.io"

var DefaultContainerLabels = []string{LabelPrefix("uuid").String()}

type LabelPrefix string

func (d LabelPrefix) String() string {
	return fmt.Sprintf("%s/%s", DefaultLabelPrefix, string(d))
}

// Label is a map representing metadata of each resource.
type Label map[string]string

func (l *Label) Get(key string) string {
	if val, ok := (*l)[key]; ok {
		return val
	}
	return ""
}

func (l *Label) Set(key string, value string) {
	(*l)[key] = value
}

func (l *Label) Delete(key string) {
	delete(*l, key)
}

func (l *Label) AppendMap(m map[string]string) {
	for k, v := range m {
		l.Set(k, v)
	}
}

type Selector interface {
	Matches(labels Label) bool
}

type EqualitySelector struct {
	Key   string
	Value string
}

func (s EqualitySelector) Matches(labels Label) bool {
	return labels[s.Key] == s.Value
}

type SetSelector struct {
	Key    string
	Values []string
	In     bool // true for "in", false for "notin"
}

func (s SetSelector) Matches(labels Label) bool {
	value, exists := labels[s.Key]
	if !exists {
		return false
	}
	if slices.Contains(s.Values, value) {
		return s.In
	}
	return !s.In
}

type ExistsSelector struct {
	Key string
}

func (s ExistsSelector) Matches(labels Label) bool {
	_, exists := labels[s.Key]
	return exists
}

type CompositeSelector struct {
	Selectors []Selector
}

func (s CompositeSelector) Matches(labels Label) bool {
	for _, selector := range s.Selectors {
		if !selector.Matches(labels) {
			return false
		}
	}
	return true
}

func NewCompositeSelectorFromMap(labelMap map[string]string) CompositeSelector {
	selectors := make([]Selector, 0, len(labelMap))
	for key, value := range labelMap {
		selectors = append(selectors, EqualitySelector{Key: key, Value: value})
	}
	return CompositeSelector{Selectors: selectors}
}

func New() Label {
	return Label{}
}
