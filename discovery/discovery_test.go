package discovery

import (
	"net/url"
	"testing"

	"github.com/livepeer/go-livepeer/core"
	"github.com/livepeer/go-livepeer/eth"
	"github.com/livepeer/go-livepeer/server"
)

type stubOffchainOrchestrators struct {
	uri   []*url.URL
	bcast server.Broadcaster
}

func StubOffchainOrchestrators(addresses []string) *stubOffchainOrchestrators {
	var uris []*url.URL

	for _, addr := range addresses {
		uri, err := url.ParseRequestURI(addr)
		if err == nil {
			uris = append(uris, uri)
		}
	}
	node, _ := core.NewLivepeerNode(nil, "", nil)
	bcast := core.NewBroadcaster(node)

	return &stubOffchainOrchestrators{bcast: bcast, uri: uris}
}

func TestNewOrchestratorPool(t *testing.T) {
	node, _ := core.NewLivepeerNode(nil, "", nil)
	addresses := []string{"https://127.0.0.1:8936", "https://127.0.0.1:8937", "https://127.0.0.1:8938"}
	expectedOffchainOrch := StubOffchainOrchestrators(addresses)

	offchainOrch := NewOrchestratorPool(node, addresses)

	for i, uri := range offchainOrch.uri {
		if uri.String() != expectedOffchainOrch.uri[i].String() {
			t.Error("Uri(s) in NewOrchestratorPool do not match expected values")
		}
	}

	addresses[0] = "https://127.0.0.1:89"
	expectedOffchainOrch = StubOffchainOrchestrators(addresses)

	if offchainOrch.uri[0].String() == expectedOffchainOrch.uri[0].String() {
		t.Error("Uri string from NewOrchestratorPool not expected to match expectedOffchainOrch")
	}

	node.Eth = &eth.StubClient{}
	expectedRegisteredTranscoders, err := node.Eth.RegisteredTranscoders()
	if err != nil {
		t.Error("Unable to get expectedRegisteredTranscoders")
	}

	offchainOrchFromOnchainList := NewOnchainOrchestratorPool(node)
	for i, uri := range offchainOrchFromOnchainList.uri {
		if uri.String() != expectedRegisteredTranscoders[i].ServiceURI {
			t.Error("Uri(s) in NewOrchestratorPoolFromOnchainList do not match expected values")
		}
	}
}
