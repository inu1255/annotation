package annotation

import (
	"fmt"
	"reflect"
	"testing"
)

// abc
func Abc() {

}

type A struct {
	Name string
}

// show the name
func (this *A) Show() {
	fmt.Println(this.Name)
}

func Test_Parse(t *testing.T) {
	a := &A{"inu1255"}
	typ := reflect.TypeOf(a)
	funcDecl := GetStructMethod(typ, "Show")
	if funcDecl == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(funcDecl.Doc.Text())
	}

	funcDecl = GetFunc("Abc", "")
	if funcDecl == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(funcDecl.Doc.Text())
	}

	funcDecl = GetFunc("Rdb", "github.com/inu1255/gev2/models")
	if funcDecl == nil {
		fmt.Println("nil")
	} else {
		fmt.Println(funcDecl.Doc.Text())
	}
}
