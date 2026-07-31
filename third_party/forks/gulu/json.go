// Gulu - Golang common utilities for everyone.
// Copyright (c) 2019-present, Scribli


package gulu

import (
	"encoding/json"
)

func (*GuluJSON) UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (*GuluJSON) MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (*GuluJSON) MarshalIndentJSON(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}
