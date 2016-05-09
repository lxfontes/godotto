package vmtest

import (
	"testing"

	"github.com/aybabtme/godotto"
	"github.com/aybabtme/godotto/internal/ottoutil"
	"github.com/aybabtme/godotto/pkg/extra/do/cloud"
	"github.com/aybabtme/godotto/pkg/extra/do/mockcloud"
	"github.com/robertkrimen/otto"
)

// A RunOption is applied on the otto VM before the test begins.
type RunOption func(vm *otto.Otto) error

// Run the JS source against godotto.
func Run(t testing.TB, cloud cloud.Client, src string, opts ...RunOption) {

	if cloud == nil {
		cloud = mockcloud.Client(nil)
	}

	vm := otto.New()

	pkg, err := godotto.Apply(vm, cloud)
	if err != nil {
		t.Fatal(err)
	}
	vm.Set("cloud", pkg)
	vm.Set("assert", func(call otto.FunctionCall) otto.Value {
		vm := call.Otto
		v, err := call.Argument(0).ToBoolean()
		if err != nil {
			ottoutil.Throw(vm, err.Error())
		}
		if v {
			return otto.UndefinedValue()
		}
		msg := "assertion failed!"
		if len(call.ArgumentList) > 1 {
			format, err := call.ArgumentList[1].ToString()
			if err != nil {
				ottoutil.Throw(vm, err.Error())
			}
			msg += "\n" + call.CallerLocation() + " | " + format
		}
		ottoutil.Throw(vm, msg)
		return otto.UndefinedValue()
	})
	script, err := vm.Compile("", src)
	if err != nil {
		t.Fatalf("invalid code: %v", err)
	}

	for _, opt := range opts {
		if err := opt(vm); err != nil {
			t.Fatalf("can't apply option: %v", err)
		}
	}

	if _, err := vm.Run(script); err != nil {
		oe := err.(*otto.Error)
		t.Fatalf(oe.String())
	}
}
