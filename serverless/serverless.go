package serverless

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]Serverless)
)

type Serverless interface {
	Init(sourceFile string)
	Build(clean bool) error
	Run() error
}

func Register(ext string, s Serverless) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if s == nil {
		panic("serverless: Register serverless is nil")
	}
	if _, dup := drivers[ext]; dup {
		panic("serverless: Register called twice for source " + ext)
	}
	drivers[ext] = s
}

func Resolve(source string) (Serverless, error) {
	// isSource := false
	ext := filepath.Ext(source)
	// if ext != "" && ext != ".exe" {
	// 	isSource = true
	// }

	driversMu.RLock()
	s, ok := drivers[ext]
	driversMu.RUnlock()
	log.Printf("sourceFile ext: %s", ext)
	if ok {
		s.Init(source)
		return s, nil
	}
	return nil, fmt.Errorf(`serverless: unsupport "%s" source (forgotten import?)`, ext)
}
