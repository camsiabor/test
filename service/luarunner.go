package service

import (
	"fmt"
	"github.com/camsiabor/golua/lua"
	"github.com/camsiabor/golua/luar"
	"github.com/camsiabor/qcom/qref"

	"github.com/camsiabor/qcom/util"

	"path/filepath"
	"runtime"
	"strings"
)

func GetVal(L *lua.State, idx int) (interface{}, error) {
	if L.IsNoneOrNil(idx) {
		return nil, nil
	}
	var ltype = int(L.Type(idx))
	switch ltype {
	case int(lua.LUA_TNUMBER):
		return L.ToNumber(idx), nil
	case int(lua.LUA_TSTRING):
		return L.ToString(idx), nil
	case int(lua.LUA_TBOOLEAN):
		return L.ToBoolean(idx), nil
	}
	var r interface{}
	var err = luar.LuaToGo(L, idx, &r)
	return r, err
}

func FormatStack(stacks []lua.LuaStackEntry) []lua.LuaStackEntry {
	var count = len(stacks)
	var clones = make([]lua.LuaStackEntry, count)
	for i := 0; i < count; i++ {
		var stack = stacks[i]
		var clone = lua.LuaStackEntry{
			Name: stack.Name,
		}
		var linenum = stack.CurrentLine
		if linenum >= 0 {
			var lines = strings.Split(stack.Source, "\n")
			if linenum < len(lines) {
				clone.ShortSource = lines[linenum-1]
			} else {
				clone.ShortSource = stack.ShortSource
			}
		}
		clone.Source = ""
		clone.CurrentLine = linenum
		clones[i] = clone
	}
	return clones
}

func FormatStackToString(stacks []lua.LuaStackEntry, prefix string, suffix string) string {
	var str = ""
	var count = len(stacks)

	for i := 0; i < count; i++ {
		var stack = stacks[i]
		var source = stack.ShortSource
		var funcname = stack.Name
		var linenum = stack.CurrentLine
		if linenum >= 0 {
			var lines = strings.Split(stack.Source, "\n")
			if linenum < len(lines) {
				source = lines[linenum-1]
			}
		}
		var one = fmt.Sprintf("%s%s %s %d\n%s", prefix, source, funcname, linenum, suffix)
		str = str + one
	}
	return str
}

func FormatStackToMap(stacks []lua.LuaStackEntry) []map[string]interface{} {
	var count = len(stacks)
	var clones = make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		var stack = stacks[i]
		var clone = make(map[string]interface{})
		var linenum = stack.CurrentLine
		if linenum >= 0 {
			var lines = strings.Split(stack.Source, "\n")
			if linenum < len(lines) {
				clone["linesrc"] = lines[linenum-1]
			} else {
				clone["linesrc"] = stack.ShortSource
			}
		}
		clone["line"] = linenum
		clone["func"] = stack.Name
		clones[i] = clone
	}
	return clones
}

func GetLuaPath(luaPath, luaCPath string) (luaPathAbs, luaCPathAbs, luaPathFull, luaCPathFull string) {

	var err error
	var luaVersion = lua.GetVersionNumber()
	var luaVersionWithoutDot = lua.GetVersionNumberWithoutDot()

	if luaPath, err = filepath.Abs(luaPath); err != nil {
		panic(err)
	}
	if luaCPath, err = filepath.Abs(luaCPath); err != nil {
		panic(err)
	}

	var luaLibSuffix = "so"
	if runtime.GOOS == "windows" {
		luaPath = strings.Replace(luaPath, "\\", "/", -1)
		luaCPath = strings.Replace(luaCPath, "\\", "/", -1)
		luaLibSuffix = "dll"
	}

	if luaPath[:len(luaPath)-1] != "/" {
		luaPath = luaPath + "/"
	}

	if luaCPath[:len(luaCPath)-1] != "/" {
		luaCPath = luaCPath + "/"
	}

	luaPathFull = luaPath + "?.lua;" +
		luaPath + "?init.lua;" +
		luaPath + "?;" +
		luaPath + "lib/?.lua;" +
		luaPath + "lib/?init.lua;" +
		luaPath + "lib/?;"

	luaCPathFull = luaCPath + "lib/?." + luaLibSuffix + ";" +
		luaCPath + "lib/?" + luaVersion + "." + luaLibSuffix + ";" +
		luaCPath + "lib/?" + luaVersionWithoutDot + "." + luaLibSuffix + ";" +
		luaCPath + "lib/load.all." + luaLibSuffix + ";" +
		luaCPath + "lib/?;"

	return luaPath, luaCPath, luaPathFull, luaCPathFull
}

func InitState(luaPath, luaCPath string) (L *lua.State, err error) {

	defer func() {
		var pan = recover()
		if pan != nil {
			if L != nil {
				L.Close()
			}
			var ok bool
			err, ok = pan.(error)
			if !ok {
				err = fmt.Errorf("lua init state error %v", pan)
			}
		}
	}()

	L = luar.Init()

	L.OpenBase()
	L.OpenLibs()
	L.OpenTable()
	L.OpenString()
	L.OpenPackage()
	L.OpenOS()
	L.OpenMath()
	L.OpenDebug()
	L.OpenBit32()
	L.OpenDebug()

	err = luar.SetLuaPath(L, luaPath, luaCPath)

	return L, err
}

func DefaultErrHandler(L *lua.State, pan interface{}) {
	if pan == nil {
		return
	}
	var ok bool
	L.Notice, ok = pan.(*lua.Interrupt)
	if !ok {
		var stackinfo = qref.StackInfo(5)
		var stackstr = util.AsStr(stackinfo["stack"], "")
		stackstr = strings.Replace(stackstr, "\t", "  ", -1)
		stackinfo["stack"] = strings.Split(stackstr, "\n")
		L.SetData("err_stack", stackinfo)
	}
}

func RunFile(L *lua.State, filename string, errhandler lua.LuaGoErrHandler) (rets []interface{}, err error) {

	var fpath = filename
	L.GetGlobal(luar.LUA_PATH_ABS)
	if L.IsString(-1) {
		var abs = L.ToString(-1)
		fpath = abs + filename
	}
	L.Pop(-1)

	var top_before = L.GetTop()
	if err = L.LoadFileEx(fpath); err != nil {
		return rets, err
	}

	if errhandler == nil {
		errhandler = DefaultErrHandler
	}

	err = L.CallHandle(0, lua.LUA_MULTRET, errhandler)
	var top_after = L.GetTop()
	var return_num = top_after - top_before
	if err == nil {
		if return_num > 0 {
			rets = make([]interface{}, return_num)
			for i := 0; i < return_num; i++ {
				rets[i], err = GetVal(L, i+1)
				if err != nil {
					break
				}
			}
		}
	}
	top_after = L.GetTop()
	if top_after-top_before > 0 {
		for i := 0; i < return_num; i++ {
			L.Pop(-1)
		}
	}

	if err != nil {
		var qrace = L.GetData("qrace")
		if qrace != nil {
			rets = make([]interface{}, 1)
			rets[0] = qrace
		}
	}

	return rets, err
}
