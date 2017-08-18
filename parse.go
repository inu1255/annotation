package annotation

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path"
	"reflect"
	"runtime"
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

// get func by name
func GetFuncByName(methodName, structPkg string) (funcDecl *ast.FuncDecl) {
	return getFunc(methodName, structPkg, nil)
}

// get func or struct method
func GetFunc(f interface{}) (funcDecl *ast.FuncDecl) {
	pc := reflect.ValueOf(f).Pointer()
	fn := runtime.FuncForPC(pc)
	// github.com/inu1255/annotation.Abc
	// github.com/inu1255/annotation.(*A).Show-fm
	methodName := fn.Name()
	dir, file := path.Split(methodName)
	ss := strings.Split(file, ".")
	// github.com/inu1255/annotation
	structPkg := path.Join(dir, ss[0])
	n := len(ss) - 1
	methodName = ss[n]
	structName := ""
	if n > 1 {
		structName = strings.Trim(ss[1], "(*)")
		methodName = ss[n][:strings.LastIndex(ss[n], "-")]
	}
	return getFunc2(methodName, structPkg, structName)
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
	if fn := getFunc2(methodName, structPkg, structName); fn != nil {
		return fn
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

func getFunc2(methodName, structPkg, structName string) (funcDecl *ast.FuncDecl) {
	var mpkg map[string]*ast.Package
	var err error
	fset := token.NewFileSet()
	if structPkg == "" {
		mpkg, err = parser.ParseDir(fset, ".", nil, parser.ParseComments)
	} else {
		gopath, goroot := GetGopathGoroot()
		dir := path.Join("vendor", structPkg)
		if _, err := os.Stat(dir); err != nil {
			dir = path.Join(gopath, "src", structPkg)
			if _, err := os.Stat(dir); err != nil {
				dir = path.Join(goroot, "src", structPkg)
			}
		}
		mpkg, err = parser.ParseDir(fset, dir, nil, parser.ParseComments)
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
	return
}

func GetGopathGoroot() (gopath, goroot string) {
	output, _ := exec.Command("go", "env").Output()
	s := string(output)
	lines := strings.Split(s, "\n")
	for _, s := range lines {
		ss := strings.Split(s, "=")
		if len(ss) > 1 {
			if ss[0] == "GOPATH" {
				gopath = strings.Trim(ss[1], "\"")
			} else if ss[0] == "GOROOT" {
				goroot = strings.Trim(ss[1], "\"")
			}
		}
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
