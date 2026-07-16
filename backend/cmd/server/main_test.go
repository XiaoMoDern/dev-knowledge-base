package main

import "testing"

func TestNewServerUsesLocalDevelopmentAddress(t *testing.T) {
	server := newServer()

	if server.Addr != "127.0.0.1:8181" {
		t.Fatalf("server address = %q, want %q", server.Addr, "127.0.0.1:8181")
	}

	if server.Handler == nil {
		t.Fatal("server handler must not be nil")
	}
}
