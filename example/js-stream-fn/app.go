package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/dop251/goja"
)

// Handler will handle the raw data.
func Handler(data []byte) (byte, []byte) {
	source, err := os.ReadFile("app.js")
	if err != nil {
		return 0x0, nil
	}
	err = JsHandler(string(source), data, func(id byte, payload []byte) {
		var noise float32
		err := json.Unmarshal(payload, &noise)
		if err != nil {
			log.Printf(">> [sink] unmarshal data failed, err=%v", err)
		} else {
			log.Printf(">> [sink] save `%v` to FaunaDB\n", noise)
		}

	})
	// error
	if err != nil {
		return 0x0, nil
	}
	// result
	return 0x35, nil
}

func DataID() []byte {
	return []byte{0x34}
}

func JsHandler(source string, data []byte, callback func(byte, []byte)) error {
	// vm
	vm := goja.New()
	vm.Set("log", log.Default())
	vm.Set("yomo", Yomo{})
	_, err := vm.RunString(source)
	if err != nil {
		return err
	}
	// observe data id
	// TODO: 需要移出去
	var dataIDFn func() []byte
	var dataIDs []byte
	err = vm.ExportTo(vm.Get("dataID"), &dataIDFn)
	if err != nil {
		return err
	}
	dataIDs = dataIDFn()
	log.Printf("<< JS DataID result: dataIDs=%v\n", dataIDs)
	// handler
	var handlerFn func(goja.ArrayBuffer) map[string]interface{}
	err = vm.ExportTo(vm.Get("handler"), &handlerFn)
	if err != nil {
		return err
	}
	// wrapped data
	buf := vm.NewArrayBuffer(data)
	log.Printf("source.Hex: %x\n", buf)
	result := handlerFn(buf)
	// result data id
	var id byte = 0x0
	if v, ok := result["id"].(byte); ok {
		id = v
	}
	// result payload
	var payload []byte
	if v, ok := result["data"].(goja.ArrayBuffer); ok {
		payload = v.Bytes()
	}
	log.Printf("<< JS Handler result: id=%v, payload=%s\n", id, payload)
	callback(id, payload)

	return nil
}

type Yomo struct{}

func (y Yomo) BytesToString(buf []byte) string {
	return string(buf)
}
func (y Yomo) ArrayBufferToString(buf goja.ArrayBuffer) string {
	return string(buf.Bytes())
}
