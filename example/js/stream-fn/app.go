package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dop251/goja"
	"github.com/yomorun/yomo"
)

// Serverless main function
func main() {
	sfn := yomo.NewStreamFunction("Noise", yomo.WithZipperAddr("127.0.0.1:9000"))
	defer sfn.Close()
	// js source
	source, err := os.ReadFile("app.js")
	if err != nil {
		log.Printf("read file err: %v\n", err)
		return
	}
	// create a javascript vm
	vm := goja.New()
	vm.Set("log", log.Default())
	vm.Set("arrayBufferToString", func(buf goja.ArrayBuffer) string { return string(buf.Bytes()) })
	vm.Set("stringToArrayBuffer", func(v string) goja.ArrayBuffer {
		return vm.NewArrayBuffer([]byte(v))
	})
	prg, err := complieJS(vm, string(source))
	if err != nil {
		log.Printf("complie js err: %v\n", err)
		return
	}

	// set observe DataIDs
	sfn.SetObserveDataID(getObserveDataID(vm, prg)...)

	// set handler
	sfn.SetHandler(wrappedHandler(vm, prg))

	// start
	err = sfn.Connect()
	if err != nil {
		log.Printf("[flow] connect err=%v\n", err)
	}

	select {}
}

func wrappedHandler(vm *goja.Runtime, prg *goja.Program) func(data []byte) (byte, []byte) {
	return func(data []byte) (byte, []byte) {
		id, payload, err := jsHandler(vm, prg, data)
		if err != nil {
			log.Printf("app err: %v", err)
			return 0x0, nil
		}
		return id, payload
	}
}

// complieJS
func complieJS(vm *goja.Runtime, source string) (*goja.Program, error) {
	prg, err := goja.Compile("", source, false)
	if err != nil {
		return nil, err
	}
	_, err = vm.RunProgram(prg)
	if err != nil {
		return nil, err
	}
	return prg, nil
}

// getObserveDataID
func getObserveDataID(vm *goja.Runtime, prg *goja.Program) (dataIDs []byte) {
	var dataIDFn func() []byte
	jsFn := vm.Get("dataID")
	if jsFn == nil {
		log.Println("`dataID` function is not found")
		return
	}
	err := vm.ExportTo(jsFn, &dataIDFn)
	if err != nil {
		log.Println(err)
	}
	dataIDs = dataIDFn()
	return
}

func jsHandler(vm *goja.Runtime, prg *goja.Program, data []byte) (id byte, payload []byte, err error) {
	var handlerFn func(goja.ArrayBuffer) map[string]interface{}
	fn := vm.Get("handler")
	if fn == nil {
		err = fmt.Errorf("`handler` function is not found")
		return
	}
	err = vm.ExportTo(fn, &handlerFn)
	if err != nil {
		return
	}
	// wrapped data
	buf := vm.NewArrayBuffer(data)
	result := handlerFn(buf)
	// result data id
	if v, ok := result["id"].(int64); ok {
		id = byte(v)
	}
	// result payload
	if v, ok := result["data"].(goja.ArrayBuffer); ok {
		payload = v.Bytes()
	}

	return
}
