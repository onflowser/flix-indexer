package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/language"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	flixv11 "flix-indexer/flix/v1_1"
	"github.com/go-git/go-git/v5"
	"golang.org/x/text/cases"
)

type Repository struct {
	GitUrl           string
	InteractionPaths []string
}

var cleanupTempFolders = false

var repos = []Repository{
	{
		GitUrl: "https://github.com/onflow/flow-nft",
		InteractionPaths: []string{
			"transactions/*",
		},
	},
	{
		GitUrl: "https://github.com/onflow/flow-evm-bridge",
		InteractionPaths: []string{
			"cadence/transactions/bridge/nft/*",
			"cadence/transactions/bridge/onboarding/*",
			"cadence/transactions/bridge/tokens/*",
			"cadence/transactions/evm/*",
			"cadence/transactions/flow-token/*",
		},
	},
	{
		GitUrl: "https://github.com/onflow/flow-core-contracts",
		InteractionPaths: []string{
			"transactions/accounts/*",
			"transactions/flowToken/*",
			"transactions/randomBeaconHistory/*",
		},
	},
	{
		GitUrl: "https://github.com/emerald-dao/float",
		InteractionPaths: []string{
			"src/flow/cadence/scripts/*",
			"src/flow/cadence/transactions/*",
		},
	},
}

func processCadenceFile(projectPath, cdcFilePath string) {
	projectName := filepath.Base(projectPath)
	flixFileName := strings.Replace(filepath.Base(cdcFilePath), ".cdc", ".template.json", 1)
	flixTemplate, err := flixGenerate(projectPath, cdcFilePath)
	if err != nil {
		log.Printf("Failed processing %s: %v\n", cdcFilePath, err)
		return
	}

	fmt.Printf("Succeed processing %s:\n", cdcFilePath)
	flixSubFolderPath := filepath.Join("templates", projectName)
	flixFilePath := filepath.Join(flixSubFolderPath, flixFileName)

	if _, err := os.Stat(flixSubFolderPath); os.IsNotExist(err) {
		err = os.MkdirAll(flixSubFolderPath, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating directory: %v\n", err)
		}
	}

	err = os.WriteFile(flixFilePath, []byte(flixTemplate), 0644)
	if err != nil {
		log.Fatalf("Error writing file: %v\n", err)
	}
}

// flixGenerate generates a Flix template from the given Cadence file
func flixGenerate(projectPath, sourceFilePath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(
		// This Flow CLI binary was manually built from source
		// and includes this change: https://github.com/onflow/flixkit-go/pull/78
		path.Join(cwd, "./flow-cli-next"),
		"flix",
		"generate",
		path.Join(cwd, sourceFilePath),
		"--exclude-networks=testing,previewnet",
	)
	cmd.Dir = projectPath

	rawTemplate, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run flow-cli: %v, rawTemplate: %s", err, rawTemplate)
	}

	template, err := flixv11.NewFromJson(rawTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template from json: %s", err)
	}

	existingTitle := template.GetMessage("title", "en-US")
	if existingTitle == "" {
		template.SetMessage("title", "en-US", humanizeFileName(sourceFilePath))
	}

	updatedRawTemplate, err := json.MarshalIndent(template, "", "\t")
	if err != nil {
		return "", fmt.Errorf("failed to marshal template to json: %s", err)
	}

	return string(updatedRawTemplate), nil
}

func humanizeFileName(filePath string) string {
	rawFileName := path.Base(filePath)
	rawFileName = strings.ReplaceAll(rawFileName, ".cdc", "")
	rawFileName = string(regexp.MustCompile("[_-]").ReplaceAll([]byte(rawFileName), []byte(" ")))
	rawFileName = cases.Title(language.English, cases.Compact).String(rawFileName)
	return rawFileName
}

func runEmulator(projectPath string) (*exec.Cmd, error) {
	cmd := exec.Command("flow", "emulator")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout

	err := cmd.Start()
	if err != nil {
		return cmd, err
	}

	return cmd, nil
}

func processRepositories(repos []Repository) error {
	for _, repo := range repos {
		repoName := filepath.Base(repo.GitUrl)
		repoDir := filepath.Join("temp", repoName)

		emulatorCmd, err := runEmulator(repoDir)
		if err != nil {
			return err
		}
		// Wait for emulator to start up
		time.Sleep(time.Second)

		if _, err := os.Stat(repoDir); os.IsNotExist(err) {
			fmt.Printf("Cloning %s into %s...\n", repo.GitUrl, repoDir)
			err := cloneRepository(repo.GitUrl, repoDir)
			if err != nil {
				log.Fatalf("Failed to clone repository: %v\n", err)
			}
		}

		for _, interactionPath := range repo.InteractionPaths {
			files, err := findFiles(repoDir, interactionPath)
			if err != nil {
				log.Fatalf("Failed to find files: %v\n", err)
			}

			for _, file := range files {
				if strings.HasSuffix(file, ".cdc") {
					processCadenceFile(repoDir, file)
				}
			}
		}

		if cleanupTempFolders {
			cleanupDirectory(repoDir)
		}

		err = emulatorCmd.Process.Kill()
		if err != nil {
			return err
		}
	}
	return nil
}

func cloneRepository(gitUrl, targetDir string) error {
	_, err := git.PlainClone(targetDir, false, &git.CloneOptions{
		URL:      gitUrl,
		Progress: os.Stdout,
	})
	return err
}

func findFiles(basePath, pattern string) ([]string, error) {
	files, err := filepath.Glob(filepath.Join(basePath, pattern))
	if err != nil {
		return nil, err
	}
	return files, nil
}

func cleanupDirectory(directory string) {
	err := os.RemoveAll(directory)
	if err != nil {
		log.Printf("Error cleaning up directory %s: %v\n", directory, err)
	} else {
		fmt.Printf("Cleaned up %s\n", directory)
	}
}

func main() {
	err := processRepositories(repos)
	if err != nil {
		panic(err)
	}
}
