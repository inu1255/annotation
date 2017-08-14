package annotation

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"reflect"
	"strings"
)

// get struct method
func GetText(m reflect.Method) string {
	funcDecl := GetStructMethodByMethod(m)
	if funcDecl == nil {
		return ""
	}
	return funcDecl.Doc.Text()
}

func GetStructMethodByMethod(m reflect.Method) (funcDecl *ast.FuncDecl) {
	typ := m.Type.In(0).Elem()
	methodName := m.Name
	return GetStructMethod(typ, methodName)
}

// get struct method
func GetStructMethod(typ reflect.Type, methodName string) (funcDecl *ast.FuncDecl) {
	return getFunc(methodName, "", typ)
}

// get func
func GetFunc(methodName, structPkg string) (funcDecl *ast.FuncDecl) {
	return getFunc(methodName, structPkg, nil)
}

// find func or struct method
// find func by methodName,structPkg
// find struct method by methodName,typ
func getFunc(methodName, structPkg string, typ reflect.Type) (funcDecl *ast.FuncDecl) {
	structName := ""
	typ = IsStruct(typ)
	if typ != nil {
		structName = typ.Name()
		structPkg = typ.PkgPath()
	}
	var mpkg map[string]*ast.Package
	var err error
	fset := token.NewFileSet()
	separator := "/"
	if os.IsPathSeparator('\\') {
		separator = "\\"
	}
	if strings.Contains(structPkg, separator) {
		mpkg, err = parser.ParseDir(fset, path.Join(os.Getenv("GOPATH"), "src", structPkg), nil, parser.ParseComments)
	} else {
		mpkg, err = parser.ParseDir(fset, ".", nil, parser.ParseComments)
	}
	if err != nil {
		return nil
	}
	for _, pkg := range mpkg {
		for _, f := range pkg.Files {
			for _, decl := range f.Decls {
				if fun, ok := decl.(*ast.FuncDecl); ok {
					if fun.Name.Name == methodName {
						if structName == "" { // func Foo()
							return fun
						} else if fun.Recv != nil && len(fun.Recv.List) > 0 { // func (*Bar)Foo()
							if starExpr, ok := fun.Recv.List[0].Type.(*ast.StarExpr); ok {
								if index, ok := starExpr.X.(*ast.Ident); ok {
									if index.Name == structName {
										return fun
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if structName != "" {
		// find parent's func
		WalkFields(typ, func(field reflect.StructField) bool {
			if field.Anonymous {
				WalkMethods(field.Type, func(method reflect.Method) bool {
					if method.Name == methodName {
						funcDecl = GetStructMethodByMethod(method)
						return true
					}
					return false
				})
			}
			return false
		})
	}
	return
}

func IsStruct(typ reflect.Type) reflect.Type {
	if typ == nil {
		return nil
	}
	switch typ.Kind() {
	case reflect.Interface, reflect.Ptr:
		return IsStruct(typ.Elem())
	case reflect.Struct:
		return typ
	}
	return nil
}

func WalkFields(typ reflect.Type, call func(reflect.StructField) bool) {
	switch typ.Kind() {
	case reflect.Interface, reflect.Ptr:
		WalkFields(typ.Elem(), call)
	case reflect.Struct:
		numField := typ.NumField()
		for i := 0; i < numField; i++ {
			field := typ.Field(i)
			// Log.Println("field", field.Name, field.PkgPath, field.Anonymous, field.Type)
			if call(field) {
				break
			}
		}
	}
}

func WalkMethods(typ reflect.Type, call func(reflect.Method) bool) {
	switch typ.Kind() {
	case reflect.Interface, reflect.Ptr:
		numMethod := typ.NumMethod()
		for i := 0; i < numMethod; i++ {
			method := typ.Method(i)
			// typ := method.Type.In(0).Elem()
			// Log.Println("method", typ.PkgPath(), typ.Name(), method.Name)
			if call(method) {
				break
			}
		}
	case reflect.Struct:
		WalkMethods(reflect.PtrTo(typ), call)
	}
}