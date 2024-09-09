package v1_1

import (
	"fmt"
	"os"
	"path"
)

type TemplateIndexer struct {
	// lookupByAstHash stores a lookup from Cadence AST hash to template
	// Note: this map is not garbage collectible.
	// If we ever need to free memory for unused templates,
	//we'd have to use a different approach.
	lookupByAstHash map[string]*Template
	// templates stores all available (deduplicated) templates
	templates []*Template
}

const templatesDirPath = "./templates/"

func NewIndexer() *TemplateIndexer {
	return &TemplateIndexer{
		lookupByAstHash: make(map[string]*Template),
	}
}

func (i *TemplateIndexer) SeedFromFs() error {
	templatesDir, err := os.Open(templatesDirPath)

	if err != nil {
		return err
	}

	repoFolders, err := templatesDir.Readdirnames(0)

	if err != nil {
		return err
	}

	for _, repoName := range repoFolders {
		repoPath := path.Join(templatesDirPath, repoName)
		templates, err := os.ReadDir(repoPath)

		if err != nil {
			return err
		}

		for _, entry := range templates {
			if entry.IsDir() {
				return fmt.Errorf("expected file, found dir: %s", entry.Name())
			}

			templatePath := path.Join(repoPath, entry.Name())
			rawTemplateContent, err := os.ReadFile(templatePath)
			if err != nil {
				return err
			}

			template, err := NewFromJson(rawTemplateContent)
			if err != nil {
				return err
			}

			err = i.add(template)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (i *TemplateIndexer) add(template *Template) error {
	astHashes, err := template.CadenceAstHashes()
	if err != nil {
		return err
	}

	for _, astHash := range astHashes {
		i.lookupByAstHash[string(astHash)] = template
	}

	i.templates = append(i.templates, template)

	return nil
}

func (i *TemplateIndexer) List() []*Template {
	return i.templates
}

func (i *TemplateIndexer) GetByID(id string) *Template {
	for _, template := range i.lookupByAstHash {
		if template.Id == id {
			return template
		}
	}
	return nil
}

func (i *TemplateIndexer) GetBySource(cadenceSource []byte) (*Template, error) {
	astHash, err := cadenceAstHash(cadenceSource)
	if err != nil {
		return nil, err
	}
	return i.lookupByAstHash[string(astHash)], nil
}
