package main
import "testing"
func TestProcessGate(t *testing.T){
 gate:=newProcessGate(2)
 if !gate.TryEnter()||!gate.TryEnter(){t.Fatal("expected two available jobs")}
 if gate.TryEnter(){t.Fatal("third heavy job should be rejected")}
 gate.Leave()
 if !gate.TryEnter(){t.Fatal("slot not released")}
 gate.Leave();gate.Leave()
}
