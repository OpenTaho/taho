package main

import "testing"

func TestHandleNoArgs(t *testing.T) {
	taho := NewTahoWithMockProxy()
	taho.HandleArgs()
}

func TestHandleUnknownArg(t *testing.T) {
	taho := NewTahoWithMockProxy()
	taho.proxy.args = []string{"test", "unknown"}
	taho.proxy.expected = append(taho.proxy.expected, []string{"Fatalf", "Unable to handle argumet \"%s\"", "unknown"})
	taho.HandleArgs()
}
