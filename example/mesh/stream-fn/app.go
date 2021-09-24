package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/yomorun/cli/rx"
	"github.com/yomorun/yomo"
)

// NoiseDataKey represents the Tag of a Y3 encoded data packet
const NoiseDataKey = 0x10

// NoiseData represents the structure of data
type NoiseData struct {
	Noise float32 `json:"noise"`
	Time  int64   `json:"time"`
	From  string  `json:"from"`
}

var region = os.Getenv("REGION")

var printer = func(_ context.Context, i interface{}) (interface{}, error) {
	value := i.(*NoiseData)
	rightNow := time.Now().UnixNano() / int64(time.Millisecond)
	fmt.Println(fmt.Sprintf("%s %d > value: %f ⚡️=%dms", value.From, value.Time, value.Noise, rightNow-value.Time))
	value.Noise = value.Noise / 10
	return value, nil
}

// Handler will handle data in Rx way
func Handler(rxstream rx.Stream) rx.Stream {
	log.Println("Handler is running...")
	stream := rxstream.
		Unmarshal(json.Unmarshal, func() interface{} { return &NoiseData{} }).
		Debounce(50).
		Map(printer).
		Marshal(json.Marshal).
		PipeBackToZipper(0x14)

	return stream
}

func main() {
	addr := fmt.Sprintf("%s:%d", "localhost", getPort())
	sfn := yomo.NewStreamFunction("Noise", yomo.WithZipperAddr(addr))
	defer sfn.Close()

	// set observe DataIDs
	sfn.SetObserveDataID(DataID()...)

	// create a Rx runtime.
	rt := rx.NewRuntime(sfn)

	// set handler
	sfn.SetHandler(rt.RawByteHandler)

	// start
	err := sfn.Connect()
	if err != nil {
		log.Print("❌ Connect to YoMo-Zipper failure: ", err)
		return
	}

	// pipe rx stream and rx handler.
	rt.Pipe(Handler)

	select {}
}

func DataID() []byte {
	return []byte{0x10}
}

func getPort() int {
	port := 9000
	if os.Getenv("PORT") != "" && os.Getenv("PORT") != "9000" {
		port, _ = strconv.Atoi(os.Getenv("PORT"))
	}

	return port
}
