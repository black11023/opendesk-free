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

func TestExternalRendererAdmission(t *testing.T){
 external:=newProcessGate(1)
 if !external.TryEnter(){t.Fatal("external renderer should acquire single slot")}
 if external.TryEnter(){t.Fatal("second renderer should be rejected, not queued")}
 external.Leave()
 if !external.TryEnter(){t.Fatal("slot not released after failure or completion")}
 external.Leave()
 for _,op:=range []string{"office-pdf","office-pdf-batch","ocr-pdf","ocr-images"}{
  if !isExternalOperation(op){t.Fatal(op+" was not guarded")}
 }
 for _,op:=range []string{"pdf-audit","pdf-merge","organize","unknown"}{
  if isExternalOperation(op){t.Fatal(op+" must not consume external slot")}
 }
}
