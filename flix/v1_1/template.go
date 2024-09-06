package v1_1

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/turbolent/prettier"
	"strings"
)

type Template struct {
	astHash []byte
	raw     []byte
	parsed  parsedTemplate
}

type parsedTemplate struct {
	FType    string     `json:"f_type"`
	FVersion string     `json:"f_version"`
	Id       string     `json:"id"`
	Data     parsedData `json:"data"`
}

type parsedData struct {
	Type    string        `json:"type"`
	Cadence parsedCadence `json:"cadence"`
}

type parsedCadence struct {
	Body string `json:"body"`
}

func NewFromJson(rawJson []byte) (Template, error) {
	var parsed parsedTemplate
	err := json.Unmarshal(rawJson, &parsed)
	if err != nil {
		return Template{}, err
	}

	if parsed.FVersion != "1.1.0" {
		return Template{}, fmt.Errorf("unsupported f_version '%s'", parsed.FVersion)
	}

	astHash, err := cadenceAstHash([]byte(parsed.Data.Cadence.Body))
	if err != nil {
		return Template{}, err
	}

	return Template{
		astHash: astHash,
		raw:     rawJson,
		parsed:  parsed,
	}, nil
}

func (t Template) MarshalJSON() ([]byte, error) {
	return t.raw, nil
}

func (t Template) MatchesSource(source []byte) (bool, error) {
	astHash, err := cadenceAstHash(source)
	if err != nil {
		return false, err
	}

	return bytes.Equal(t.astHash, astHash), nil
}

func cadenceAstHash(source []byte) ([]byte, error) {
	program, err := parser.ParseProgram(nil, source, parser.Config{})
	if err != nil {
		return nil, err
	}

	program.Doc()

	var prettySource strings.Builder
	prettier.Prettier(&prettySource, program.Doc(), 80, "    ")

	s := sha256.New()
	s.Write([]byte(prettySource.String()))
	hash := s.Sum(nil)

	return hash, nil
}
