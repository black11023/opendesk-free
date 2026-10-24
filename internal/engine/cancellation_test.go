package engine
import (
"context"
"errors"
"testing"
)
func TestCancelledInputFailsClosed(t *testing.T){
 ctx,cancel:=context.WithCancel(context.Background());cancel()
 for _,op:=range []string{"pdf-merge","images-pdf","organize","csv-xlsx"}{
  _,e:=RunContext(ctx,op,[]Input{{Name:"sample.txt",Data:[]byte("hello")}})
  if !errors.Is(e,context.Canceled){t.Fatalf("%s: expected context cancellation, got %v",op,e)}
 }
}
