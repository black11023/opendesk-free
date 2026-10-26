package main

// processGate caps heavy request concurrency. No queue of large uploads is held.
type processGate struct{slots chan struct{}}
func newProcessGate(max int)*processGate{return &processGate{slots:make(chan struct{},max)}}
func(g *processGate)TryEnter()bool{
 select{case g.slots<-struct{}{}:return true;default:return false}
}
func(g *processGate)Leave(){<-g.slots}

// Only these operations spawn optional third-party document processors.
func isExternalOperation(op string)bool{
 switch op {case "office-pdf","office-pdf-batch","ocr-pdf","ocr-images":return true}
 return false
}
