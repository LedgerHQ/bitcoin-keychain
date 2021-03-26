// +build mage

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/magefile/mage/sh"
)

const (
	entryPoint = "cmd/server.go"
	ldFlags    = "-X $PACKAGE/version/version.commitHash=$COMMIT_HASH " +
		"-X $PACKAGE/version/version.buildDate=$BUILD_DATE"
	protoPlugins  = "plugins=grpc"
	protoDir      = "pb"
	protoKeychainFileName = "keychain/service.proto"
	protoGrpcClientFileName = "bitcoin/service.proto"
	protoArtifactModule = "github.com/ledgerhq/bitcoin-keychain/pb"
)

// Allow user to override executables on UNIX-like systems.
var (
	goexe  = "go"     // GOEXE=xxx mage build
	protoc = "protoc" // PROTOC=xxx mage proto
	buf    = "buf"    // BUF=xxx mage protolist
)

func init() {
	if exe := os.Getenv("GOEXE"); exe != "" {
		goexe = exe
	}

	if exe := os.Getenv("PROTOC"); exe != "" {
		protoc = exe
	}

	if exe := os.Getenv("BUF"); exe != "" {
		buf = exe
	}

	// We want to use Go 1.11 modules even if the source lives inside GOPATH.
	// The default is "auto".
	os.Setenv("GO111MODULE", "on")
}

func Proto() error {
	runner := func(proto string, dir string) error {
		return sh.Run(protoc,
			fmt.Sprintf("--go_out=%s:%s", protoPlugins, dir),  // protoc flags
			fmt.Sprintf("%s/%s", dir, proto),          // input .proto
			fmt.Sprintf("--go_opt=module=%s", protoArtifactModule), // module output
		)
	}

	if err := runner(protoKeychainFileName, protoDir); err != nil {
		return err
	}

	return runner(protoGrpcClientFileName, protoDir)
}

func Buf() error {
	// Verify if the proto files can be compiled.
	if err := sh.Run(buf, "image", "build", "-o /dev/null"); err != nil {
		return err
	}

	// Run Buf lint checks on the protobuf file.
	if err := sh.Run(buf, "check", "lint"); err != nil {
		return err
	}

	return nil
}

// Build binary
func Build() error {
	if err := Proto(); err != nil {
		return err
	}

	return sh.RunWith(flagEnv(), goexe, "build", "-ldflags", ldFlags,
		entryPoint)
}

// Run tests
func Test() error {
	return sh.Run(goexe, "test", "./...")
}

// Run integration tests
func Integration() error {
	return sh.Run(goexe, "test", "--tags=integration", "./...")
}

// Run tests with race detector
func TestRace() error {
	return sh.Run(goexe, "test", "-race", "./...")
}

// Run tests with race-detector and code-coverage.
// Useful on CI, but can be run locally too.
func TestRaceCover() error {
	return sh.Run(
		goexe, "test", "-race", "-coverprofile=coverage.txt",
		"-covermode=atomic", "./...")
}

// Run basic golangci-lint check.
func Lint() error {
	linterArgs := []string{
		"run",
		"-E=golint",
		"-E=interfacer",
		"-E=unconvert",
		"-E=dupl",
		"-E=goconst",
		"-E=gofmt",
		"-E=goimports",
		"-E=maligned",
		"-E=depguard",
		"-E=misspell",
		"-E=whitespace",
		"-E=gocritic",
	}

	if err := sh.Run("golangci-lint", linterArgs...); err != nil {
		return err
	}

	return nil
}

func flagEnv() map[string]string {
	hash, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return map[string]string{
		"PACKAGE":     entryPoint,
		"COMMIT_HASH": hash,
		"BUILD_DATE":  time.Now().Format("2006-01-02T15:04:05Z0700"),
	}
}
