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

// @path 1111111
// @path 2222222
// --> {"path":[1111111,2222222]}
func Doc2Map(doc string) map[string][]string {
	m := make(map[string][]string)
	for _, line := range strings.Split(doc, "\n") {
		if i := strings.Index(line, "@"); i >= 0 {
			line = line[i+1:]
			i = strings.IndexAny(line, " \t")
			var key, value string
			if i >= 0 {
				key = line[:i]
				value = line[i+1:]
			} else {
				key = line
			}
			if sa, ok := m[key]; ok {
				m[key] = append(sa, value)
			} else {
				m[key] = []string{value}
			}
		}
	}
	return m
}

// get func remment text
func Text(funcDecl *ast.FuncDecl) string {
	if funcDecl == nil {
		return ""
	}
	return funcDecl.Doc.Text()
}

// get func param list
func Params(funcDecl *ast.FuncDecl) []string {
	if funcDecl == nil {
		return make([]string, 0)
	}
	params := make([]string, 0, len(funcDecl.Type.Params.List))
	for _, param := range funcDecl.Type.Params.List {
		for _, item := range param.Names {
			params = append(params, item.Name)
		}
	}
	return params
}

// get func return list
func Outs(funcDecl *ast.FuncDecl) []string {
	if funcDecl.Type.Results == nil {
		return make([]string, 0)
	}
	params := make([]string, 0, len(funcDecl.Type.Results.List))
	for _, param := range funcDecl.Type.Results.List {
		for _, item := range param.Names {
			params = append(params, item.Name)
		}
	}
	return params
}

// get func or struct method
func GetFunc(f interface{}) (funcDecl *ast.FuncDecl) {
	switch v := f.(type) {
	case reflect.Method:
		return getFuncByMethod(v)
	default:
		return FindFunc(GetFuncInfo(f))
	}
}

func GetFuncInfo(f interface{}) (string, string, string) {
	switch v := f.(type) {
	case reflect.Method:
		typ := isStruct(v.Type.In(0))
		methodName, structPkg, structName := v.Name, typ.PkgPath(), typ.Name()
		return methodName, structPkg, structName
	default:
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
		return methodName, structPkg, structName
	}
}

func getFuncByMethod(m reflect.Method) (funcDecl *ast.FuncDecl) {
	typ := isStruct(m.Type.In(0))
	methodName := m.Name
	return getFunc(methodName, "", typ)
}

// find func or struct method
// find func by methodName,structPkg
// find struct method by methodName,typ
func getFunc(methodName, structPkg string, typ reflect.Type) (funcDecl *ast.FuncDecl) {
	structName := ""
	typ = isStruct(typ)
	if typ != nil {
		structName = typ.Name()
		structPkg = typ.PkgPath()
	}
	if fn := FindFunc(methodName, structPkg, structName); fn != nil {
		return fn
	}
	if structName != "" {
		// find parent's func
		walkFields(typ, func(field reflect.StructField) bool {
			if field.Anonymous {
				walkMethods(field.Type, func(method reflect.Method) bool {
					if method.Name == methodName {
						funcDecl = getFuncByMethod(method)
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

func FindFunc(methodName, structPkg, structName string) (funcDecl *ast.FuncDecl) {
	var mpkg map[string]*ast.Package
	var err error
	fset := token.NewFileSet()
	if structPkg == "" || structPkg == "main" {
		mpkg, err = parser.ParseDir(fset, ".", nil, parser.ParseComments)
	} else {
		gopath, goroot := getGopathGoroot()
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
								if ident, ok := starExpr.X.(*ast.Ident); ok {
									if ident.Name == structName {
										return fun
									}
								}
							} else if ident, ok := fun.Recv.List[0].Type.(*ast.Ident); ok {
								if ident.Name == structName {
									return fun
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

func getGopathGoroot() (gopath, goroot string) {
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

func isStruct(typ reflect.Type) reflect.Type {
	if typ == nil {
		return nil
	}
	switch typ.Kind() {
	case reflect.Interface, reflect.Ptr:
		return isStruct(typ.Elem())
	case reflect.Struct:
		return typ
	}
	return nil
}

func walkFields(typ reflect.Type, call func(reflect.StructField) bool) {
	switch typ.Kind() {
	case reflect.Interface, reflect.Ptr:
		walkFields(typ.Elem(), call)
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

func walkMethods(typ reflect.Type, call func(reflect.Method) bool) {
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
		walkMethods(reflect.PtrTo(typ), call)
	}
}
