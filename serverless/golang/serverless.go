package golang

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/yomorun/cli/serverless"
)

type GolangServerless struct {
	source string
	target string
}

func (s *GolangServerless) Init(sourceFile string) {
	s.source = sourceFile
}

func (s *GolangServerless) Build(clean bool) error {
	// s.target = s.source + ".build"
	// TODO: 增加main代码，构建完成代码build
	// check if the file exists
	appPath := s.source
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return fmt.Errorf("the file %s doesn't exist", appPath)
	}

	// build
	version := runtime.GOOS
	dir, _ := filepath.Split(appPath)
	sl := dir + "sl.yomo"

	// clean build
	if clean {
		// .so file exists, remove it.
		if _, err := os.Stat(sl); !os.IsNotExist(err) {
			err = os.Remove(sl)
			if err != nil {
				return fmt.Errorf("clean build the file %s failed", appPath)
			}
		}
	}
	s.target = sl
	if version == "linux" {
		cmd := exec.Command("/bin/sh", "-c", "CGO_ENABLED=0 GOOS=linux go build -ldflags \"-s -w\" -o "+sl+" "+appPath)
		out, err := cmd.CombinedOutput()
		if err != nil && len(out) > 0 {
			// get error message from stdout.
			err = errors.New("\n" + string(out))
		}
		return err
	} else if version == "darwin" {
		cmd := exec.Command("/bin/sh", "-c", "go build -ldflags \"-s -w\" -o "+sl+" "+appPath)
		out, err := cmd.CombinedOutput()
		if err != nil && len(out) > 0 {
			// get error message from stdout.
			err = errors.New("\n" + string(out))
		}
		return err
	} else {
		cmd := exec.Command("go build -ldflags \"-s -w\" -o " + sl + " " + appPath)
		out, err := cmd.CombinedOutput()
		if err != nil && len(out) > 0 {
			// get error message from stdout.
			err = errors.New("\n" + string(out))
		}
		return err
	}

	return nil
}
func (s *GolangServerless) Run() error {
	log.Printf("s.traget: %s", s.target)
	// result, err := standalone.RunCmdAndWait(s.target)
	cmd := exec.Command(s.target)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func init() {
	serverless.Register(".go", &GolangServerless{})
}
