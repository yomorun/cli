package serverless

import (
	"io"

	"github.com/yomorun/yomo/pkg/rx"
)

// DispatcherWithFunc dispatches the input stream to downstreams.
func DispatcherWithFunc(flows []func() (io.ReadWriter, func()), reader io.Reader) rx.RxStream {
	stream := rx.FromReader(reader)

	for _, flow := range flows {
		stream = stream.MergeReadWriterWithFunc(flow)
	}

	return stream
}
