package annotation

import (
	"encoding/json"
	"go/ast"
	"io/ioutil"
	"reflect"
)

type funcInfo struct {
	StructName string
	Name       string
	Doc        string
	Params     []string
}

type FuncInfoCache map[string]*funcInfo

func (this FuncInfoCache) Restore(filename string) bool {
	if data, err := ioutil.ReadFile(filename); err == nil {
		err = json.Unmarshal(data, this)
		return err == nil
	}
	return false
}

func (this FuncInfoCache) Save(filename string) bool {
	if data, err := json.Marshal(this); err == nil {
		err = ioutil.WriteFile(filename, data, 0644)
		return err == nil
	}
	return false
}

func (this FuncInfoCache) ReadFunc(f interface{}) *funcInfo {
	if f == nil {
		return nil
	}
	var funcDecl *ast.FuncDecl
	var methodName, structPkg, structName string
	switch method := f.(type) {
	case reflect.Method:
		typ := isStruct(method.Type.In(0))
		methodName, structPkg, structName = method.Name, typ.PkgPath(), typ.Name()
		funcDecl = GetFunc(method)
	default:
		methodName, structPkg, structName = GetFuncInfo(f)
		funcDecl = FindFunc(methodName, structPkg, structName)
	}
	key := structPkg + "." + structName + "." + methodName
	// fmt.Println(key, funcDecl)
	if funcDecl == nil {
		return this[key]
	}
	info := &funcInfo{
		StructName: structName,
		Name:       methodName,
		Doc:        funcDecl.Doc.Text(),
		Params:     Params(funcDecl),
	}
	this[key] = info
	return info
}
