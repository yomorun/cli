package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	y3 "github.com/yomorun/y3-codec-golang"
	"github.com/yomorun/yomo/pkg/rx"

	"github.com/yomorun/yomo/pkg/client"
)

var store = func(_ context.Context, i interface{}) (interface{}, error) {
	value := i.(string)
	fmt.Printf("save `%v` to FaunaDB\n", value)
	return value, nil
}

var callback = func(v []byte) (interface{}, error) {
	return y3.ToUTF8String(v)
}

// Handler will handle data in Rx way
func Handler(rxstream rx.RxStream) rx.RxStream {
	stream := rxstream.
		Subscribe(0x11).
		OnObserve(callback).
		AuditTime(100).
		Map(store).
		Encode(0x12)
	return stream
}

func main() {
	cli, err := client.NewOutputConnector("MockDB").Connect("localhost", getPort())
	if err != nil {
		log.Print("❌ Connect to yomo-server failure: ", err)
		return
	}

	defer cli.Close()
	cli.Run(Handler)
}

func getPort() int {
	port := 9000
	if os.Getenv("PORT") != "" && os.Getenv("PORT") != "9000" {
		port, _ = strconv.Atoi(os.Getenv("PORT"))
	}
	
	return port
}