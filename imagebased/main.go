// Copyright 2024 The Inspektor Gadget authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/inspektor-gadget/inspektor-gadget/pkg/datasource"
	igjson "github.com/inspektor-gadget/inspektor-gadget/pkg/datasource/formatters/json"
	gadgetcontext "github.com/inspektor-gadget/inspektor-gadget/pkg/gadget-context"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/operators"
	_ "github.com/inspektor-gadget/inspektor-gadget/pkg/operators/ebpf"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/operators/kubemanager"
	ocihandler "github.com/inspektor-gadget/inspektor-gadget/pkg/operators/oci-handler"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/operators/simple"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/operators/socketenricher"
	"github.com/inspektor-gadget/inspektor-gadget/pkg/runtime/local"
)

func do() error {
	const opPriority = 50000
	myOperator := simple.New("myHandler", simple.OnInit(func(gadgetCtx operators.GadgetContext) error {
		for _, d := range gadgetCtx.GetDataSources() {
			jsonFormatter, _ := igjson.New(d,
				// Show all fields
				igjson.WithShowAll(true),

				// Print json in a pretty format
				igjson.WithPretty(true, "  "),
			)

			d.Subscribe(func(source datasource.DataSource, data datasource.Data) error {
				jsonOutput := jsonFormatter.Marshal(data)
				fmt.Printf("%s\n", jsonOutput)
				return nil
			}, opPriority)
		}
		return nil
	}))

	// There are some gadgets, like trace_dns, trace_sni, trace_network that
	// require the kube manager operator to work when running in Kubernetes.

	// Create the kube manager operator for Kubernetes environments.
	// It automatically detects containers and enriches events with K8s metadata.
	kubeManagerOp := kubemanager.KubeManagerOperator
	kubeManagerParams := kubeManagerOp.GlobalParamDescs().ToParams()
	if err := kubeManagerOp.Init(kubeManagerParams); err != nil {
		return fmt.Errorf("init kube manager: %w", err)
	}
	defer kubeManagerOp.Close()

	// The socket enricher operator is used to provide information about the
	// process performing the DNS query.
	socketEnricherOp := &socketenricher.SocketEnricher{}
	socketEnricherParams := socketEnricherOp.GlobalParamDescs().ToParams()
	if err := socketEnricherOp.Init(socketEnricherParams); err != nil {
		return fmt.Errorf("init socket enricher: %w", err)
	}
	defer socketEnricherOp.Close()

	gadgetCtx := gadgetcontext.New(
		context.Background(),
		"ghcr.io/inspektor-gadget/gadget/trace_dns:v0.46.0",
		gadgetcontext.WithDataOperators(
			ocihandler.OciHandler,
			kubeManagerOp,
			socketEnricherOp,
			myOperator,
		),
	)

	runtime := local.New()
	if err := runtime.Init(nil); err != nil {
		return fmt.Errorf("runtime init: %w", err)
	}
	defer runtime.Close()

	if err := runtime.RunGadget(gadgetCtx, nil, nil); err != nil {
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
