package golang

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"

	// "os/exec"
	"path/filepath"
	"runtime"

	exec "golang.org/x/sys/execabs"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/imports"

	"github.com/yomorun/cli/pkg/file"
	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/cli/serverless"
)

type GolangServerless struct {
	opts     *serverless.Options
	source   string
	target   string
	tempDir  string
	buildDir string
}

func (s *GolangServerless) Init(opts *serverless.Options) error {
	// now := time.Now()
	// msg := "Init: serverless function..."
	// initSpinning := log.Spinner(os.Stdout, msg)
	// defer initSpinning(log.Failure)

	s.opts = opts
	if !file.Exists(s.opts.Filename) {
		return fmt.Errorf("the file %s doesn't exist", s.opts.Filename)
	}
	// s.buildDir = file.Dir(s.opts.Filename)
	// generate source code
	source := file.GetBinContents(s.opts.Filename)
	if len(source) < 1 {
		return fmt.Errorf(`"%s" content is empty`, s.opts.Filename)
	}
	// append main function
	// // TODO: 替换变量
	ctx := Context{
		Name: s.opts.Name,
		Host: s.opts.Host,
		Port: s.opts.Port,
	}
	mainFunc, err := RenderTmpl(string(MainFuncTmpl), &ctx)
	if err != nil {
		return fmt.Errorf("Init: %s", err)
	}
	source = append(source, mainFunc...)
	// log.InfoStatusEvent(os.Stdout, "merge source elapse: %v", time.Since(now))
	// Create the AST by parsing src
	fset := token.NewFileSet()
	astf, err := parser.ParseFile(fset, "", source, 0)
	if err != nil {
		return err
	}
	// Add import packages
	astutil.AddNamedImport(fset, astf, "yomoclient", "github.com/yomorun/yomo/pkg/client")
	astutil.AddNamedImport(fset, astf, "stdlog", "log")
	// log.InfoStatusEvent(os.Stdout, "import elapse: %v", time.Since(now))
	// Generate the code
	code, err := GenerateCode(fset, astf)
	if err != nil {
		return fmt.Errorf("Init: generate code err %s", err)
	}
	// Create a temp folder.
	tempDir, err := ioutil.TempDir("", "yomo_")
	if err != nil {
		return err
	}
	s.tempDir = tempDir
	tempFile := filepath.Join(tempDir, "app.go")
	// Fix imports
	fixedSource, err := imports.Process(tempFile, code, nil)
	if err != nil {
		return fmt.Errorf("Init: imports %s", err)
	}
	// log.InfoStatusEvent(os.Stdout, "fix import elapse: %v", time.Since(now))
	if err := file.PutContents(tempFile, fixedSource); err != nil {
		return fmt.Errorf("Init: write file err %s", err)
	}
	// log.InfoStatusEvent(os.Stdout, "final write file elapse: %v", time.Since(now))
	// mod
	cmd := exec.Command("/bin/sh", "-c", "go mod init yomo")
	// cmd := exec.Command(fmt.Sprintf("go mod init %s", s.opts.Name))
	cmd.Dir = tempDir
	env := os.Environ()
	env = append(env, fmt.Sprintf("GO111MODULE=%s", "on"))
	cmd.Env = env
	// log.InfoStatusEvent(os.Stdout, "Init: cmd: %+v", cmd)
	// log.InfoStatusEvent(os.Stdout, "Init: cmd.Dir: %v", cmd.Dir)
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		// get error message from stdout.
		err = errors.New("Build: go mod init err " + string(out))
		return err
	}

	// TODO: 检查临时目录是否已存构建源码文件md5
	s.source = tempFile
	return nil
}

func (s *GolangServerless) Build(clean bool) error {
	// check if the file exists
	appPath := s.source
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return fmt.Errorf("the file %s doesn't exist", appPath)
	}

	// mod
	env := os.Environ()
	env = append(env, fmt.Sprintf("GO111MODULE=%s", "on"))
	// yomo
	cmd := exec.Command("/bin/sh", "-c", "go get -u github.com/yomorun/yomo")
	cmd.Dir = s.tempDir
	// cmd := exec.Command("go mod download")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		// get error message from stdout.
		err = errors.New("Build: go get yomo " + string(out))
		return err
	}
	// y3-codec
	cmd = exec.Command("/bin/sh", "-c", "go get -u github.com/yomorun/y3-codec-golang")
	cmd.Env = env
	cmd.Dir = s.tempDir
	// cmd := exec.Command("go mod download")
	out, err = cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		// get error message from stdout.
		err = errors.New("Build: go get y3-codec-golang" + string(out))
		return err
	}
	// deps
	cmd = exec.Command("/bin/sh", "-c", "go mod download")
	cmd.Env = env
	cmd.Dir = s.tempDir
	// cmd := exec.Command("go mod download")
	out, err = cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		// get error message from stdout.
		err = errors.New("Build: go mod download err " + string(out))
		return err
	}

	// build
	goos := runtime.GOOS
	dir, _ := filepath.Split(s.opts.Filename)
	sl, _ := filepath.Abs(dir + "sl.yomo")

	// clean build
	if clean {
		defer func() {
			file.Remove(s.tempDir)
		}()
	}
	s.target = sl
	if goos == "windows" {
		sl = dir + "sl.exe"
		cmd := exec.Command("go build -ldflags \"-s -w\" -o " + sl + " " + appPath)
		cmd.Env = env
		cmd.Dir = ""
		out, err := cmd.CombinedOutput()
		if err != nil && len(out) > 0 {
			// get error message from stdout.
			err = errors.New("Build: " + string(out))
		}
		return err
	}
	// *nix
	cmd = exec.Command("/bin/sh", "-c", "go build -ldflags \"-s -w\" -o "+sl+" "+appPath)
	cmd.Env = env
	cmd.Dir = s.tempDir
	// log.InfoStatusEvent(os.Stdout, "Build: cmd: %+v", cmd)
	// source := file.GetContents(s.source)
	// log.InfoStatusEvent(os.Stdout, "source: %s", source)
	out, err = cmd.CombinedOutput()
	if err != nil && len(out) > 0 {
		// get error message from stdout.
		err = errors.New("Build: failure " + string(out))
	}
	return err
}

func (s *GolangServerless) Run() error {
	log.InfoStatusEvent(os.Stdout, "Run: %s", s.target)
	cmd := exec.Command(s.target)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func GenerateCode(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func init() {
	serverless.Register(".go", &GolangServerless{})
}
