package v1_1

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/onflow/cadence/runtime/parser"
	"github.com/turbolent/prettier"
	"regexp"
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
	Contract string            `json:"contract"`
	Networks []contractNetwork `json:"networks"`
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

// CadenceAstHashes returns all Cadence AST hashes that map to the semantically equivalent source code
func (l *Template) CadenceAstHashes() ([][]byte, error) {
	supportedNetworkNames := []string{
		"emulator",
		"testnet",
		"mainnet",
	}
	var astHashes [][]byte

	// Also include plan source code, without patched imports
	astHash, err := cadenceAstHash([]byte(l.Data.Cadence.Body))
	if err != nil {
		return nil, err
	}
	astHashes = append(astHashes, astHash)

	for _, networkName := range supportedNetworkNames {
		sourceCode := l.sourceCodeForNetwork(networkName)
		if sourceCode == "" {
			// Network not supported for this template
			continue
		}

		astHash, err := cadenceAstHash([]byte(sourceCode))
		if err != nil {
			return nil, err
		}
		astHashes = append(astHashes, astHash)
	}

	return astHashes, nil
}

// sourceCodeForNetwork returns source code for a specific network name.
// Returns an empty string if no source code can be produced for a given network.
func (l *Template) sourceCodeForNetwork(networkName string) string {
	sourceCode := l.Data.Cadence.Body

	for _, deps := range l.Data.Dependencies {
		for _, dep := range deps.Contracts {
			var network *contractNetwork
			for _, net := range dep.Networks {
				if net.Network == networkName {
					network = &net
				}
			}

			if network == nil {
				// Can't build source code for this network
				return ""
			}

			importPattern := regexp.MustCompile(fmt.Sprintf(`import +"%s"`, dep.Contract))
			replacementImport := fmt.Sprintf("import %s from %s", dep.Contract, prefixedAddress(network.Address))
			sourceCode = string(importPattern.ReplaceAll([]byte(sourceCode), []byte(replacementImport)))
		}
	}

	return sourceCode
}

func (l *Template) GetMessage(key, tag string) string {
	for _, msg := range l.Data.Messages {
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

func (l *Template) SetMessage(key, tag, translation string) {
	for _, msg := range l.Data.Messages {
		if msg.Key == key {
			for _, msgI18n := range msg.I18n {
				if msgI18n.Tag == tag {
					msgI18n.Translation = translation
					return
				}
			}
		}
	}

	l.Data.Messages = append(l.Data.Messages, message{
		Key: key,
		I18n: []i18n{
			{
				Tag:         tag,
				Translation: translation,
			},
		},
	})
}

func prefixedAddress(address string) string {
	if strings.HasPrefix(address, "0x") {
		return address
	} else {
		return "0x" + address
	}
}

func cadenceAstHash(source []byte) ([]byte, error) {
	program, err := parser.ParseProgram(nil, source, parser.Config{})
	if err != nil {
		return nil, err
	}

	var prettySource strings.Builder
	prettier.Prettier(&prettySource, program.Doc(), 80, "    ")

	s := sha256.New()
	s.Write([]byte(prettySource.String()))
	hash := s.Sum(nil)

	return hash, nil
}
