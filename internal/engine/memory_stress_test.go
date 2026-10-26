package engine

import (
 "context"
 "runtime"
 "testing"
)
// This catches gross retained-Go-heap regressions for repeated real PDF
// transformations. It is NOT a hard RSS limit for external executables.
func TestRepeatedBatchPDFRetainedHeap(t *testing.T) {
 input:=twoPageFixture(t)
 inputs:=make([]Input,5)
 for i:=range inputs { inputs[i]=input }
 runtime.GC()
 var before,after runtime.MemStats
 runtime.ReadMemStats(&before)
 for n:=0;n<16;n++{
  result,e:=RunContext(context.Background(),"pdf-rotate",inputs)
  if e!=nil{t.Fatalf("batch %d failed: %v",n,e)}
  if len(result.Data)==0||len(result.Data)>80<<20{t.Fatal("empty or oversized ZIP")}
 }
 runtime.GC()
 runtime.ReadMemStats(&after)
 retainedMiB:=float64(int64(after.HeapAlloc)-int64(before.HeapAlloc))/(1<<20)
 t.Logf("Retained Go heap after repeated batch PDF: %.2f MiB, baseline %.2f MiB, after %.2f MiB",
  retainedMiB,float64(before.HeapAlloc)/(1<<20),float64(after.HeapAlloc)/(1<<20))
 if retainedMiB>96 {t.Fatalf("retained Go heap increase > 96 MiB: %.2f",retainedMiB)}
}
