package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"dagger.io/dagger"
	"github.com/pkg/errors"
)

const (
	pythonImage    = "python:3.11"
	postgresImage  = "postgres:latest"
	cacheVolumeKey = "go-cache"
	cacheMountPath = "/cache"
)

var (
	golangImage      = fmt.Sprintf("golang:%s", strings.TrimPrefix(runtime.Version(), "go"))
	cacheGoBuildPath = filepath.Join(cacheMountPath, "go-build")
	cacheGoModPath   = filepath.Join(cacheMountPath, "go-mod")
)

type Config struct {
	Source  string                `json:"source"`
	Targets map[string]TestTarget `json:"targets"`
}

type TestTarget struct {
	Repository   string   `json:"repository"`
	Tag          string   `json:"tag"`
	Patch        string   `json:"patch"`
	Requirements []string `json:"requirements"`
	Tests        []string `json:"tests"`
}

func main() {
	cacheFrom := flag.String("cache-from", "", "Load the Go build cache from the given directory")
	cacheTo := flag.String("cache-to", "", "Save the Go build cache to the given directory")
	targetList := flag.String("targets", "*", "Comma-separated list of test targets to run")
	configPath := flag.String("config", filepath.Join(getLocalDir(), "config.json"), "Path to config file")
	flag.Parse()

	configDir, err := filepath.Abs(filepath.Dir(*configPath))
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	// nolint:errcheck
	defer f.Close()

	var config Config
	if err := json.NewDecoder(f).Decode(&config); err != nil {
		log.Fatal(err)
	}

	client, err := dagger.Connect(context.Background(), dagger.WithLogOutput(os.Stderr))
	if err != nil {
		log.Fatalf("Error connecting to Dagger: %s\n", err)
	}
	// nolint:errcheck
	defer client.Close()

	if *cacheFrom != "" {
		_, err := client.Container().
			From("alpine").
			WithDirectory("/import", client.Host().Directory(*cacheFrom)).
			WithMountedCache(cacheMountPath, client.CacheVolume(cacheVolumeKey)).
			WithExec([]string{"rm", "-rf", cacheGoBuildPath}).
			WithExec([]string{"cp", "-r", "/import", cacheGoBuildPath}).
			Sync(context.Background())
		if err != nil {
			log.Fatalf("Error loading cache: %s\n", err)
		}
	}

	targets := strings.Split(*targetList, ",")
	if *targetList == "*" {
		targets = []string{}
		for name := range config.Targets {
			targets = append(targets, name)
		}
	}

	failed_targets := map[string]error{}
	for _, target := range targets {
		td, ok := config.Targets[target]
		if !ok {
			failed_targets[target] = errors.New("unknown target")
			continue
		}

		var srcDir *dagger.Directory
		if td.Repository != "" {
			srcDir = getPatchedSourceDirectory(client, td.Repository, td.Tag, filepath.Join(configDir, td.Patch))
		} else {
			srcDir = getLocalTestsSourceDirectory(client, filepath.Join(configDir, config.Source))
		}

		if _, err := getTestContainer(
			client,
			srcDir,
			getPythonVirtualEnvDirectory(client, td.Requirements),
			getBinary(client, filepath.Join(configDir, config.Source)),
			getDatabaseService(client),
		).
			Pipeline("tests").
			WithExec(append([]string{
				"pytest",
				"-v",
				"--color=yes",
			}, td.Tests...)).
			Sync(context.Background()); err != nil {
			failed_targets[target] = err
		}
	}

	if len(failed_targets) > 0 {
		// This is a workaround to allow the Dagger engine to flush the logs
		time.Sleep(5 * time.Second)

		ft := []string{}
		for target, err := range failed_targets {
			log.Printf("Test target %s failed: %s\n", target, err)
			ft = append(ft, target)
		}
		log.Fatalf("Test targets failed: %s\n", strings.Join(ft, ", "))
	}

	if *cacheTo != "" {
		if err := os.RemoveAll(*cacheTo); err != nil {
			log.Fatalln(err)
		}

		_, err := client.Container().
			From("alpine").
			WithMountedCache(cacheMountPath, client.CacheVolume(cacheVolumeKey)).
			WithExec([]string{"cp", "-r", cacheGoBuildPath, "/export"}).
			Directory("/export").
			Export(context.Background(), *cacheTo)
		if err != nil {
			log.Fatalf("Error saving cache: %s\n", err)
		}
	}
}

func getLocalDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

func getPythonVirtualEnvDirectory(client *dagger.Client, requirements []string) *dagger.Directory {
	return client.Container().
		From(pythonImage).
		WithExec([]string{
			"python", "-mvenv", "/venv",
		}).
		WithExec(append([]string{
			"/venv/bin/pip", "install", "--no-cache-dir",
		}, requirements...)).
		Directory("/venv")
}

func getPatchedSourceDirectory(client *dagger.Client, repo, tag, patchPath string) *dagger.Directory {
	return client.Container().
		From(pythonImage).
		WithDirectory("/src", client.Git(repo).Tag(tag).Tree()).
		WithFile("/tmp/patch", client.Host().File(patchPath)).
		WithWorkdir("/src").
		WithExec([]string{
			"git", "apply", "/tmp/patch",
		}).
		Directory("/src")
}

func getLocalTestsSourceDirectory(client *dagger.Client, sourcePath string) *dagger.Directory {
	return client.Container().
		From(pythonImage).
		WithDirectory("/src", client.Host().Directory(sourcePath, dagger.HostDirectoryOpts{
			Include: []string{"python"},
		})).
		Directory("/src")
}

func getDatabaseService(client *dagger.Client) *dagger.Service {
	return client.Container().
		From(postgresImage).
		WithEnvVariable("POSTGRES_PASSWORD", "postgres").
		WithEnvVariable("LC_COLLATE", "POSIX").
		WithExec([]string{
			"postgres",
			"-c", "log_min_error_statement=panic",
			"-c", "log_min_messages=fatal",
		}).
		WithExposedPort(5432).
		AsService()
}

func getBinary(client *dagger.Client, sourcePath string) *dagger.File {
	return client.Container().
		From(golangImage).
		WithEnvVariable("GOCACHE", cacheGoBuildPath).
		WithEnvVariable("GOMODCACHE", cacheGoModPath).
		WithMountedCache(cacheMountPath, client.CacheVolume(cacheVolumeKey)).
		WithDirectory("/src", client.Host().Directory(sourcePath, dagger.HostDirectoryOpts{
			Include: []string{"main.go", "pkg", "Makefile", ".go-build-tags", "go.mod", "go.sum", "python"},
		})).
		WithWorkdir("/src").
		WithExec([]string{
			"make", "build",
		}).
		File("/src/fml")
}

func getTestContainer(
	client *dagger.Client,
	sourceDirectory,
	venvDirectory *dagger.Directory,
	fmlBinary *dagger.File,
	dbService *dagger.Service,
) *dagger.Container {
	return client.Container().
		From(pythonImage).
		WithDirectory("/src", sourceDirectory).
		WithDirectory("/venv", venvDirectory).
		WithFile("/usr/local/bin/fml", fmlBinary).
		WithServiceBinding("postgres", dbService).
		WithWorkdir("/src").
		WithEnvVariable("PATH", "/venv/bin:$PATH", dagger.ContainerWithEnvVariableOpts{Expand: true})
}
