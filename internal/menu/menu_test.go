package menu

import (
	"ssm/internal/config"
	"testing"
)

func TestHostItem_FilterValue(t *testing.T) {
	item := HostItem{host: config.Host{Alias: "test-host", HostName: "example.com"}}
	if item.FilterValue() != "test-host" {
		t.Errorf("Expected 'test-host', got '%s'", item.FilterValue())
	}
}

func TestHostItem_Title(t *testing.T) {
	item := HostItem{host: config.Host{Alias: "test-host", HostName: "example.com"}}
	if item.Title() != "test-host" {
		t.Errorf("Expected 'test-host', got '%s'", item.Title())
	}
}

func TestHostItem_Description(t *testing.T) {
	item := HostItem{host: config.Host{Alias: "test-host", HostName: "example.com"}}
	if item.Description() != "example.com" {
		t.Errorf("Expected 'example.com', got '%s'", item.Description())
	}
}

func TestNewModel(t *testing.T) {
	hosts := []config.Host{
		{Alias: "host1", HostName: "server1.com"},
		{Alias: "host2", HostName: "server2.com"},
		{Alias: "host3", HostName: "server3.com"},
	}
	model := NewModel(hosts)

	// Check that the model has the correct number of items (just hosts, no exit)
	expectedItems := len(hosts)
	if len(model.list.Items()) != expectedItems {
		t.Errorf("Expected %d items, got %d", expectedItems, len(model.list.Items()))
	}

	// Check that the first item is correct
	if item, ok := model.list.Items()[0].(HostItem); ok {
		if item.Title() != "host1" {
			t.Errorf("Expected first item to be 'host1', got '%s'", item.Title())
		}
		if item.Description() != "server1.com" {
			t.Errorf("Expected description to be 'server1.com', got '%s'", item.Description())
		}
	} else {
		t.Error("First item is not a HostItem")
	}

	// Check that the last item is correct
	if item, ok := model.list.Items()[len(hosts)-1].(HostItem); ok {
		if item.Title() != "host3" {
			t.Errorf("Expected last item to be 'host3', got '%s'", item.Title())
		}
	} else {
		t.Error("Last item is not a HostItem")
	}

	// Check that filtering is enabled
	if !model.list.FilteringEnabled() {
		t.Error("Expected filtering to be enabled")
	}
}
