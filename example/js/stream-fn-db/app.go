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
	// sfn := yomo.NewStreamFunction("{{.Name}}", yomo.WithZipperAddr("{{.Host}}:{{.Port}}"))
	sfn := yomo.NewStreamFunction("MockDB", yomo.WithZipperAddr("127.0.0.1:9000"))
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
	prg, err := complieJS(vm, string(source))
	if err != nil {
		log.Printf("complie js err: %v\n", err)
		return
	}

	// set observe DataIDs
	// sfn.SetObserveDataID(DataID()...)
	sfn.SetObserveDataID(getObserveDataID(vm, prg)...)

	// set handler
	// sfn.SetHandler(Handler)
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
		id, payload, err := JsHandler(vm, prg, data)
		if err != nil {
			log.Printf("app err: %v", err)
			return 0x0, nil
		}
		// var noise float32
		// err = json.Unmarshal(payload, &noise)
		// if err != nil {
		// 	log.Printf(">> [sink] unmarshal data failed, err=%v", err)
		// } else {
		// 	log.Printf(">> [sink](Compile) save `%v` to FaunaDB\n", noise)
		// }

		// // result
		// log.Printf(">> [sink](Compile) return id `%v`\n", id)
		return id, payload
	}
}

// Handler will handle the raw data.
// func Handler(data []byte) (byte, []byte) {
// 	source, err := os.ReadFile("app.js")
// 	if err != nil {
// 		return 0x0, nil
// 	}

// 	id, payload, err := JsHandler(string(source), data)
// 	if err != nil {
// 		return 0x0, nil
// 	}

// 	var noise float32
// 	err = json.Unmarshal(payload, &noise)
// 	if err != nil {
// 		log.Printf(">> [sink] unmarshal data failed, err=%v", err)
// 	} else {
// 		log.Printf(">> [sink] save `%v` to FaunaDB\n", noise)
// 	}

// 	// result
// 	log.Printf(">> [sink] return id `%v`\n", id)
// 	return id, nil
// }

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

// func DataID() (dataIDs []byte) {
// 	// return []byte{0x34}
// 	source, err := os.ReadFile("app.js")
// 	if err != nil {
// 		return
// 	}
// 	// vm
// 	vm := goja.New()
// 	vm.Set("log", log.Default())
// 	// TODO: 考虑使用 Complie 复用结果,这个方法仅执行一次
// 	_, err = vm.RunString(string(source))
// 	if err != nil {
// 		return
// 	}
// 	// observe data id
// 	var dataIDFn func() []byte
// 	// var dataIDs []byte
// 	err = vm.ExportTo(vm.Get("dataID"), &dataIDFn)
// 	if err != nil {
// 		return
// 	}
// 	dataIDs = dataIDFn()
// 	log.Printf("<< JS DataID result: dataIDs=%v\n", dataIDs)
// 	return
// }

// func JsHandler(source string, data []byte) (id byte, payload []byte, err error) {
func JsHandler(vm *goja.Runtime, prg *goja.Program, data []byte) (id byte, payload []byte, err error) {
	var handlerFn func(goja.ArrayBuffer) map[string]interface{}
	// var handlerFn func(goja.ArrayBuffer) HandlerResult
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
	// log.Printf("source.Hex: %x\n", buf)
	result := handlerFn(buf)
	// result data id
	if v, ok := result["id"].(int64); ok {
		id = byte(v)
	}
	// result payload
	if v, ok := result["data"].(goja.ArrayBuffer); ok {
		payload = v.Bytes()
	}
	// log.Printf("<< JS Handler result: id=%v, payload=%s\n", id, payload)

	return
}

type HandlerResult struct {
	ID      byte   `json:"id"`
	Payload []byte `json:"payload"`
}
