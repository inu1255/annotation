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

	funcDecl := GetFunc(Abc)
	if funcDecl != nil {
		fmt.Println(funcDecl.Doc.Text())
	}

	funcDecl = GetFunc(a.Show)
	if funcDecl != nil {
		fmt.Println(funcDecl.Doc.Text())
	}

	funcDecl = GetFunc(b.Show)
	if funcDecl != nil {
		fmt.Println(funcDecl.Doc.Text())
	}

	m := reflect.TypeOf(b).Method(0)
	funcDecl = GetFunc(m)
	if funcDecl != nil {
		fmt.Println(funcDecl.Doc.Text())
	}

	// output:
	// show abc
	//
	// show A
	//
	// show B
	//
	// show B
}
