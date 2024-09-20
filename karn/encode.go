package karn

import (
	"encoding/json"
	"fmt"
)

type DataEncoder interface {
	Encode(Map) ([]byte, error)
}

type DataDecoder interface {
	Decode([]byte, any) error
}

type JSONEncoder struct{}

func (JSONEncoder) Encode(data Map) ([]byte, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}
	return bytes, nil
}

type JSONDecoder struct{}

func (JSONDecoder) Decode(b []byte, v any) error {
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("failed to decode data: %w", err)
	}
	return nil
}
