package v1_1

import (
	"fmt"
	"os"
	"path"
)

type TemplateIndexer struct {
	store []Template
}

const templatesDirPath = "./templates/"

func NewIndexer() *TemplateIndexer {
	return &TemplateIndexer{
		store: make([]Template, 0),
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

			i.add(template)
		}
	}

	return nil
}

func (i *TemplateIndexer) add(template Template) {
	i.store = append(i.store, template)
}

func (i *TemplateIndexer) List() []Template {
	return i.store
}

func (i *TemplateIndexer) LookupBySource(cadenceBase64 string) (*Template, error) {
	// TODO: Implement
	return nil, nil
}