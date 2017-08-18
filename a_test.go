package annotation

import (
	"fmt"
	"reflect"
)

// show abc
func Abc() {

}

type A struct {
	Name string
}

// show A
func (this *A) Show() {
	fmt.Println(this.Name)
}

type B struct {
	A
}

// show B
func (this *B) Show() {
}

func ExampleParse() {
	a := &A{}
	b := &B{}
	typ := reflect.TypeOf(a)
	funcDecl := GetStructMethod(typ, "Show")
	if funcDecl != nil {
		fmt.Println(funcDecl.Doc.Text())
	}

	funcDecl = GetFuncByName("Abc", "")
	if funcDecl != nil {
		fmt.Println(funcDecl.Doc.Text())
	}

	funcDecl = GetFunc(b.Show)
	if funcDecl != nil {
		fmt.Println(funcDecl.Doc.Text())
	}
	// output:
	// show A
	//
	// show abc
	//
	// show B
}
