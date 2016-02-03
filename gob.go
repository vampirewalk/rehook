package main

import (
	"bytes"
	"encoding/gob"
)

func gobEncode(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(v)
	return buf.Bytes(), err
}

func gobDecode(p []byte, v interface{}) error {
	if len(p) == 0 {
		return nil
	}
	return gob.NewDecoder(bytes.NewBuffer(p)).Decode(v)
}
