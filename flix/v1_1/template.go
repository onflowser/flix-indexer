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
	FType    string `json:"f_type"`
	FVersion string `json:"f_version"`
	Id       string `json:"id"`
	Data     data   `json:"data"`
}

type message struct {
	Key  string `json:"key"`
	I18n []i18n `json:"i18n"`
}

type i18n struct {
	Tag         string `json:"tag"`
	Translation string `json:"translation"`
}

type data struct {
	Type         string         `json:"type"`
	Interface    string         `json:"interface"`
	Cadence      cadence        `json:"cadence"`
	Parameters   []parameter    `json:"parameters"`
	Dependencies []dependencies `json:"dependencies"`
	Messages     []message      `json:"messages"`
}

type dependencies struct {
	Contracts []contractDependency `json:"contracts"`
}

type contractDependency struct {
	Contracts string            `json:"contract"`
	Networks  []contractNetwork `json:"networks"`
}

type contractNetwork struct {
	Network                  string        `json:"network"`
	Address                  string        `json:"address"`
	DependencyPinBlockHeight int           `json:"dependency_pin_block_height"`
	DependencyPin            dependencyPin `json:"dependency_pin"`
}

type dependencyPin struct {
	Pin                string          `json:"pin"`
	PinSelf            string          `json:"pin_self"`
	PinContractName    string          `json:"pin_contract_name"`
	PinContractAddress string          `json:"pin_contract_address"`
	Imports            []dependencyPin `json:"imports"`
}

type parameter struct {
	Label    string    `json:"label"`
	Index    int       `json:"index"`
	Type     string    `json:"type"`
	Messages []message `json:"messages"`
}

type cadence struct {
	Body        string       `json:"body"`
	NetworkPins []networkPin `json:"network_pins"`
}

type networkPin struct {
	Network string `json:"network"`
	PinSelf string `json:"pin_self"`
}

func NewFromJson(rawJson []byte) (*Template, error) {
	var parsed Template
	err := json.Unmarshal(rawJson, &parsed)
	if err != nil {
		return nil, err
	}

	if parsed.FVersion != "1.1.0" {
		return nil, fmt.Errorf("unsupported f_version '%s'", parsed.FVersion)
	}

	return &parsed, nil
}

func (t *Template) CadenceAstHash() ([]byte, error) {
	astHash, err := cadenceAstHash([]byte(t.Data.Cadence.Body))
	if err != nil {
		return nil, err
	}
	return astHash, nil
}

func (t *Template) MatchesSource(source []byte) (bool, error) {
	astHash1, err := cadenceAstHash(source)
	if err != nil {
		return false, err
	}

	astHash2, err := t.CadenceAstHash()
	if err != nil {
		return false, err
	}

	return bytes.Equal(astHash1, astHash2), nil
}

func (t *Template) GetMessage(key, tag string) string {
	for _, msg := range t.Data.Messages {
		if msg.Key == key {
			for _, msgI18n := range msg.I18n {
				if msgI18n.Tag == tag {
					return msgI18n.Translation
				}
			}
		}
	}
	return ""
}

func (t *Template) SetMessage(key, tag, translation string) {
	for _, msg := range t.Data.Messages {
		if msg.Key == key {
			for _, msgI18n := range msg.I18n {
				if msgI18n.Tag == tag {
					msgI18n.Translation = translation
					return
				}
			}
		}
	}

	t.Data.Messages = append(t.Data.Messages, message{
		Key: key,
		I18n: []i18n{
			{
				Tag:         tag,
				Translation: translation,
			},
		},
	})
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
