package js

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"strings"

	"os/exec"
	"path/filepath"
	"runtime"

	"golang.org/x/tools/imports"

	"github.com/spf13/viper"
	"github.com/yomorun/cli/pkg/file"
	"github.com/yomorun/cli/pkg/log"
	"github.com/yomorun/cli/serverless"
)

// JsServerless defines golang implementation of Serverless interface.
type JsServerless struct {
	opts    *serverless.Options
	source  string
	target  string
	tempDir string
}

// Init initializes the serverless
func (s *JsServerless) Init(opts *serverless.Options) error {
	s.opts = opts
	if !file.Exists(s.opts.Filename) {
		return fmt.Errorf("the file %s doesn't exist", s.opts.Filename)
	}

	// generate source code
	source := file.GetBinContents(s.opts.Filename)
	if len(source) < 1 {
		return fmt.Errorf(`"%s" content is empty`, s.opts.Filename)
	}

	// append main function
	credential := viper.GetString("credential")
	if len(credential) > 0 {
		log.InfoStatusEvent(os.Stdout, "Credential=%s", credential)
	}
	ctx := Context{
		Name: s.opts.Name,
		Host: s.opts.Host,
		Port: s.opts.Port,
		Credential: credential,
	}

	mainFuncTmpl := string(MainFuncRawBytesTmpl)
	mainFunc, err := RenderTmpl(mainFuncTmpl, &ctx)
	if err != nil {
		return fmt.Errorf("Init: render template %s", err)
	}
	// merge .js to .go source
	buffer := bytes.NewBuffer(mainFunc)
	buffer.WriteString("\nconst source = `")
	buffer.Write(source)
	buffer.WriteString("`")
	// Create the AST by parsing src
	fset := token.NewFileSet()
	astf, err := parser.ParseFile(fset, "", buffer, parser.AllErrors)
	if err != nil {
		return fmt.Errorf("Init: parse source file err %s", err)
	}
	// Add import packages
	// astutil.AddNamedImport(fset, astf, "", "github.com/yomorun/yomo")
	// astutil.AddNamedImport(fset, astf, "stdlog", "log")
	// log.InfoStatusEvent(os.Stdout, "import elapse: %v", time.Since(now))
	// Generate the code
	code, err := generateCode(fset, astf)
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
	name := strings.ReplaceAll(opts.Name, " ", "_")
	cmd := exec.Command("go", "mod", "init", name)
	cmd.Dir = tempDir
	env := os.Environ()
	env = append(env, fmt.Sprintf("GO111MODULE=%s", "on"))
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Init: go mod init err %s", out)
		return err
	}

	s.source = tempFile
	return nil
}

// Build compiles the serverless to executable
func (s *JsServerless) Build(clean bool) error {
	// check if the file exists
	appPath := s.source
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return fmt.Errorf("the file %s doesn't exist", appPath)
	}
	// env
	env := os.Environ()
	env = append(env, fmt.Sprintf("GO111MODULE=%s", "on"))
	// use custom go.mod
	if s.opts.ModFile != "" {
		mfile, _ := filepath.Abs(s.opts.ModFile)
		if !file.Exists(mfile) {
			return fmt.Errorf("the mod file %s doesn't exist", mfile)
		}
		// go.mod
		log.WarningStatusEvent(os.Stdout, "Use custom go.mod: %s", mfile)
		tempMod := filepath.Join(s.tempDir, "go.mod")
		file.Copy(mfile, tempMod)
		// source := file.GetContents(tempMod)
		// log.InfoStatusEvent(os.Stdout, "go.mod: %s", source)
		// mod download
		cmd := exec.Command("go", "mod", "tidy")
		cmd.Env = env
		cmd.Dir = s.tempDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("Build: go mod tidy err %s", out)
			return err
		}
	} else {
		// Upgrade modules that provide packages imported by packages in the main module
		cmd := exec.Command("go", "get", "-d", "-u", "./...")
		cmd.Dir = s.tempDir
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		if err != nil {
			err = fmt.Errorf("Build: go get err %s", out)
			return err
		}
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
	// fmt.Printf("goos=%s\n", goos)
	if goos == "windows" {
		sl, _ = filepath.Abs(dir + "sl.exe")
		s.target = sl
	}
	// go build
	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", sl, appPath)
	cmd.Env = env
	cmd.Dir = s.tempDir
	// log.InfoStatusEvent(os.Stdout, "Build: cmd: %+v", cmd)
	// source := file.GetContents(s.source)
	// log.InfoStatusEvent(os.Stdout, "source: %s", source)
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = fmt.Errorf("Build: failure %s", out)
		return err
	}
	return nil
}

// Run compiles and runs the serverless
func (s *JsServerless) Run(verbose bool) error {
	log.InfoStatusEvent(os.Stdout, "Run: %s", s.target)
	cmd := exec.Command(s.target)
	if verbose {
		cmd.Env = []string{"YOMO_LOG_LEVEL=debug"}
	}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func (s *JsServerless) Executable() bool {
	return false
}

func generateCode(fset *token.FileSet, file *ast.File) ([]byte, error) {
	var output []byte
	buffer := bytes.NewBuffer(output)
	if err := printer.Fprint(buffer, fset, file); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func init() {
	serverless.Register(&JsServerless{}, ".js")
}
