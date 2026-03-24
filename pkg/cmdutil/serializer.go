package cmdutil

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/ghodss/yaml"
)

type Codec interface {
	Serializer
	Deserializer
}

type Serializer interface {
	Serialize(m proto.Message) ([]byte, error)
}

type Deserializer interface {
	Deserialize(b []byte, m proto.Message) error
}

type JSONSerializer struct{}

func (s *JSONSerializer) Serialize(m proto.Message) ([]byte, error) {
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: false,
		Indent:          "  ",
	}
	b, err := marshaler.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, err
}

type YamlSerializer struct{}

func (s *YamlSerializer) Serialize(m proto.Message) ([]byte, error) {
	jsonSerializer := &JSONSerializer{}
	jsonb, err := jsonSerializer.Serialize(m)
	if err != nil {
		return nil, err
	}

	var v any
	if err := json.Unmarshal(jsonb, &v); err != nil {
		return nil, err
	}

	// marshal to YAML
	yamlb, err := yaml.JSONToYAML(jsonb)
	if err != nil {
		return nil, err
	}

	return yamlb, nil
}

type TableSerializer struct{}

func (s *TableSerializer) Serialize(m proto.Message) ([]byte, error) {
	return nil, nil
}

type JSONDeserializer struct{}

func (d *JSONDeserializer) Deserialize(b []byte, m proto.Message) error {
	unmarshaller := protojson.UnmarshalOptions{
		DiscardUnknown: true,
	}
	return unmarshaller.Unmarshal(b, m)
}

type YamlDeserializer struct{}

func (d *YamlDeserializer) Deserialize(b []byte, m proto.Message) error {
	var v any
	if err := yaml.Unmarshal(b, &v); err != nil {
		return err
	}

	// generic Go value → JSON bytes
	jsonb, err := json.Marshal(v)
	if err != nil {
		return err
	}

	jsonDeserializer := &JSONDeserializer{}
	return jsonDeserializer.Deserialize(jsonb, m)
}

type JSONCodec struct {
	*JSONSerializer
	*JSONDeserializer
}

func NewJSONCodec() Codec {
	return &JSONCodec{
		&JSONSerializer{},
		&JSONDeserializer{},
	}
}

type YamlCodec struct {
	*YamlSerializer
	*YamlDeserializer
}

func NewYamlCodec() Codec {
	return &YamlCodec{
		&YamlSerializer{},
		&YamlDeserializer{},
	}
}
