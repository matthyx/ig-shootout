package main

import (
	"context"
	"fmt"
	"os"
	"time"

	gadgetcontext "github.com/inspektor-gadget/inspektor-gadget/pkg/gadget-context"
	gadgetregistry "github.com/inspektor-gadget/inspektor-gadget/pkg/gadget-registry"
	_ "github.com/inspektor-gadget/inspektor-gadget/pkg/gadgets/trace/dns/tracer"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/gadgets/trace/dns/types"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/logger"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/params"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/runtime/local"
)

func do() error {
	runtime := local.New()
	if err := runtime.Init(nil); err != nil {
		return fmt.Errorf("runtime init: %w", err)
	}
	defer runtime.Close()

	gadgetDesc := gadgetregistry.Get("trace", "dns")
	parser := gadgetDesc.Parser()
	parser.SetEventCallback(func(ev any) {
		event := ev.(*types.Event)
		fmt.Printf("command %s (%d) resolved %s\n", event.Comm, event.Pid, event.DNSName)
	})

	gadgetCtx := gadgetcontext.NewBuiltIn(
		context.TODO(),
		"",
		runtime,
		runtime.ParamDescs().ToParams(),
		gadgetDesc,
		gadgetDesc.ParamDescs().ToParams(),
		[]string{},
		params.Collection{},
		parser,
		logger.DefaultLogger(),
		time.Duration(0),
	)

	if _, err := runtime.RunBuiltInGadget(gadgetCtx); err != nil {
		return fmt.Errorf("running gadget: %w", err)
	}

	return nil
}

func main() {
	if err := do(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
		os.Exit(1)
	}
}
