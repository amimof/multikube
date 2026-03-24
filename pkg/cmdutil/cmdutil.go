// Package cmdutil provides helper utilities and interfaces for working with command line tools
package cmdutil

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

type SyncWriter struct {
	mu sync.Mutex
	w  io.Writer
	// tw *tabwriter.Writer
}

func (sw *SyncWriter) Write(p []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(p)
}

func FormatPhase(phase string) string {
	p := strings.ToUpper(phase)
	switch p {
	case "UNKNOWN":
		return color.YellowString(strings.ToLower(p))
	case "RUNNING":
		return color.GreenString(strings.ToLower(p))
	case "PULLING":
		return color.GreenString(strings.ToLower(p))
	case "STOPPED":
		return color.RedString(strings.ToLower(p))
	case "ERROR":
		return color.HiRedString(strings.ToLower(p))
	case "SCHEDULED":
		return color.MagentaString(strings.ToLower(p))
	default:
		return color.HiBlackString(strings.ToLower(p))
	}
}

type (
	StopFunc  func()
	WatchFunc func(StopFunc) error
)

func Watch(ctx context.Context, id string, wf WatchFunc) error {
	s := false
	stop := func() {
		s = true
	}
	for {
		err := wf(stop)
		if err != nil {
			return err
		}
		if s {
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
	return nil
}

func FormatDuration(d time.Duration) string {
	// Convert the duration to whole days, hours, minutes, and seconds
	days := d / (24 * time.Hour)
	d -= days * 24 * time.Hour
	hours := d / time.Hour
	d -= hours * time.Hour
	minutes := d / time.Minute
	d -= minutes * time.Minute
	seconds := d / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm", minutes)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

type OutputFormat string

var (
	OutputFormatJSON  OutputFormat = "json"
	OutputFormatYAML  OutputFormat = "yaml"
	OutputFormatTable OutputFormat = "table"
)

func validateOutputFormat(s string) (OutputFormat, error) {
	allowedOutputs := []OutputFormat{
		OutputFormatJSON,
		OutputFormatYAML,
		OutputFormatTable,
	}

	o := strings.ToLower(s)
	if ok := slices.Contains(allowedOutputs, OutputFormat(o)); !ok {
		return "", fmt.Errorf("expected output to be one of %v", allowedOutputs)
	}

	return OutputFormat(o), nil
}

func SerializerFor(s string) (Serializer, error) {
	o, err := validateOutputFormat(s)
	if err != nil {
		return nil, err
	}

	switch o {
	case OutputFormatJSON:
		return &JSONSerializer{}, nil
	case OutputFormatYAML:
		return &YamlSerializer{}, nil
	case OutputFormatTable:
		return &TableSerializer{}, nil
	default:
		return &JSONSerializer{}, nil
	}
}

func DeserializerFor(s string) (Deserializer, error) {
	o, err := validateOutputFormat(s)
	if err != nil {
		return nil, err
	}

	switch o {
	case OutputFormatJSON:
		return &JSONDeserializer{}, nil
	case OutputFormatYAML:
		return &YamlDeserializer{}, nil
	default:
		return &JSONDeserializer{}, nil
	}
}

func CodecFor(s string) (Codec, error) {
	o, err := validateOutputFormat(s)
	if err != nil {
		return nil, err
	}

	switch o {
	case OutputFormatJSON:
		return NewJSONCodec(), nil
	case OutputFormatYAML:
		return NewYamlCodec(), nil
	default:
		return NewJSONCodec(), nil
	}
}

func dedupeStrSlice(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// ConvertKVStringsToMap converts a slice of "key=value" strings into a map.
func ConvertKVStringsToMap(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		k, v, _ := strings.Cut(value, "=")
		result[k] = v
	}
	return result
}

// ReadKVStringsMapFromLabel convers string slice into map
func ReadKVStringsMapFromLabel(labels []string) map[string]string {
	labelsDeduped := dedupeStrSlice(labels)
	return ConvertKVStringsToMap(labelsDeduped)
}
