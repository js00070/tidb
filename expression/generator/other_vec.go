// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

// +build ignore

package main

import (
	"bytes"
	"go/format"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"

	. "github.com/pingcap/tidb/expression/generator/helper"
)

const header = `// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by go generate in expression/generator; DO NOT EDIT.

package expression
`

const newLine = "\n"

const builtinOtherImports = `import (
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/types/json"
	"github.com/pingcap/tidb/util/chunk"
)
`

var builtinInTmpl = template.Must(template.New("builtinInTmpl").Parse(`
{{ define "BufAllocator" }}
	buf0, err := b.bufAllocator.get(types.ET{{ .Input.ETName }}, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf0)
	if err := b.args[0].VecEval{{ .Input.TypeName }}(b.ctx, input, buf0); err != nil {
		return err
	}
	buf1, err := b.bufAllocator.get(types.ET{{ .Input.ETName }}, n)
	if err != nil {
		return err
	}
	defer b.bufAllocator.put(buf1)
{{ end }}
{{ define "SetHasNull" }}
	for i := 0; i < n; i++ {
		if result.IsNull(i) {
			result.SetNull(i, hasNull[i])
		}
	}
	return nil
{{ end }}
{{ define "Compare" }}
	{{ if eq .Input.TypeName "Int" -}}
		compareResult = 1
		switch {
			case (isUnsigned0 && isUnsigned), (!isUnsigned0 && !isUnsigned):
				if arg1 == arg0 {
					compareResult = 0
				}
			case !isUnsigned0 && isUnsigned:
				if arg0 >= 0 && arg1 == arg0 {
					compareResult = 0
				}
			case isUnsigned0 && !isUnsigned:
				if arg1 >= 0 && arg1 == arg0 {
					compareResult = 0
				}
		}
	{{- else if eq .Input.TypeName "Decimal" -}}
		compareResult = 1
		if arg0 == arg1 {
			compareResult = 0
		}
	{{- else if eq .Input.TypeName "Time" -}}
		compareResult = arg0.Compare(arg1) 
	{{- else if eq .Input.TypeName "Duration" -}}
		compareResult = types.CompareDuration(arg0, arg1)
	{{- else if eq .Input.TypeName "JSON" -}}
		compareResult = json.CompareBinary(arg0, arg1)
	{{- else -}}
		compareResult = types.Compare{{ .Input.TypeNameInColumn }}(arg0, arg1)
	{{- end -}}
{{ end }}

{{ range . }}
{{ $InputInt := (eq .Input.TypeName "Int") }}
{{ $InputString := (eq .Input.TypeName "String") }}
{{ $InputJson := (eq .Input.TypeName "JSON") }}
{{ $InputFixed := ( .Input.Fixed ) }}
func (b *{{.SigName}}) vecEvalInt(input *chunk.Chunk, result *chunk.Column) error {
	n := input.NumRows()
	{{- template "BufAllocator" . }}
	{{- if $InputFixed }}
		args0 := buf0.{{.Input.TypeNameInColumn}}s()
	{{- end }}
	result.ResizeInt64(n, true)
	r64s := result.Int64s()
	for i:=0; i<n; i++ {
		r64s[i] = 0
	}
	hasNull := make([]bool, n)
	{{- if $InputInt }}
		isUnsigned0 := mysql.HasUnsignedFlag(b.args[0].GetType().Flag)
	{{- end }}
	var compareResult int
	args := b.args

	{{- if $InputInt }}
	if b.hashSet != nil {
		args = b.nonConstArgs
		for i := 0; i < n; i++ {
			arg0 := args0[i]
			if isUnsigned, ok := b.hashSet[arg0]; ok {
				if (isUnsigned0 && isUnsigned) || (!isUnsigned0 && !isUnsigned) {
					r64s[i] = 1
				}
				if arg0 >= 0 {
					r64s[i] = 1
				}
			}
		}
	}
	{{- end }}
	{{- if $InputString }}
	if b.hashSet != nil {
		args = b.nonConstArgs
		for i := 0; i < n; i++ {
			if buf0.IsNull(i) {
				hasNull[i] = true
				continue
			}
			arg0 := buf0.GetString(i)
			if _, ok := b.hashSet[arg0]; ok {
				r64s[i] = 1
			}
		}
	}
	{{- end }}

	for j := 1; j < len(args); j++ {
		if err := args[j].VecEval{{ .Input.TypeName }}(b.ctx, input, buf1); err != nil {
			return err
		}
		{{- if $InputInt }}
			isUnsigned := mysql.HasUnsignedFlag(args[j].GetType().Flag)
		{{- end }}
		{{- if $InputFixed }}
			args1 := buf1.{{.Input.TypeNameInColumn}}s()
			buf1.MergeNulls(buf0)
		{{- end }}
		for i := 0; i < n; i++ {
			if r64s[i] != 0 {
				continue
			}
{{- /* if is null */}}
			if buf1.IsNull(i) {{- if not $InputFixed -}} || buf0.IsNull(i) {{- end -}} {
				hasNull[i] = true
				continue
			}
	
{{- /* get args */}}
			{{- if $InputFixed }}
				arg0 := args0[i]
				arg1 := args1[i]
			{{- else }}
				arg0 := buf0.Get{{ .Input.TypeName }}(i)
				arg1 := buf1.Get{{ .Input.TypeName }}(i)
			{{- end }}
	
{{- /* compare */}}
			{{- template "Compare" . }}
			if compareResult == 0 {
				result.SetNull(i, false)
				r64s[i] = 1
			}
		} // for i
	} // for j
	{{- template "SetHasNull" . -}}
}

func (b *{{.SigName}}) vectorized() bool {
	return true
}
{{ end }}{{/* range */}}
`))

var testFile = template.Must(template.New("").Parse(`// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by go generate in expression/generator; DO NOT EDIT.

package expression

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/pingcap/check"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/tidb/types"
	"github.com/pingcap/tidb/types/json"
)

type inGener struct {
	defaultGener
}

func (g inGener) gen() interface{} {
	if rand.Float64() < g.nullRation {
		return nil
	}
	randNum := rand.Int63n(10)
	switch g.eType {
	case types.ETInt:
		if rand.Float64() < 0.5 {
			return -randNum
		}
		return randNum
	case types.ETReal:
		if rand.Float64() < 0.5 {
			return -float64(randNum)
		}
		return float64(randNum)
	case types.ETDecimal:
		d := new(types.MyDecimal)
		f := float64(randNum * 100000)
		if err := d.FromFloat64(f); err != nil {
			panic(err)
		}
		return d
	case types.ETDatetime, types.ETTimestamp:
		gt := types.FromDate(2019, 11, 2, 22, 00, int(randNum), rand.Intn(1000000))
		t := types.Time{Time: gt, Type: convertETType(g.eType)}
		return t
	case types.ETDuration:
		return types.Duration{ Duration: time.Duration(randNum) }
	case types.ETJson:
		j := new(json.BinaryJSON)
		jsonStr := fmt.Sprintf("{\"key\":%v}", randNum)
		if err := j.UnmarshalJSON([]byte(jsonStr)); err != nil {
			panic(err)
		}
		return *j
	case types.ETString:
		return fmt.Sprint(randNum)
	}
	return randNum
}

{{/* Add more test cases here if we have more functions in this file */}}
var vecBuiltin{{ .Category }}GeneratedCases = map[string][]vecExprBenchCase {
{{- range $.Functions }}
	ast.{{ .FuncName }}: {
	{{- range .Sigs }} 
		// {{ .SigName }}
		{
			retEvalType: types.ET{{ .Output.ETName }}, 
			childrenTypes: []types.EvalType{
				types.ET{{ .Input.ETName }}, 
				types.ET{{ .Input.ETName }},
				types.ET{{ .Input.ETName }},
				types.ET{{ .Input.ETName }},
			},
			geners: []dataGenerator{
				inGener{defaultGener{eType: types.ET{{.Input.ETName}}, nullRation: 0.2}},
				inGener{defaultGener{eType: types.ET{{.Input.ETName}}, nullRation: 0.2}},
				inGener{defaultGener{eType: types.ET{{.Input.ETName}}, nullRation: 0.2}},
				inGener{defaultGener{eType: types.ET{{.Input.ETName}}, nullRation: 0.2}},
			},
		}, 
	{{- end }}
{{- end }}
	}, 
}

func (s *testEvaluatorSuite) TestVectorizedBuiltin{{.Category}}EvalOneVecGenerated(c *C) {
	testVectorizedEvalOneVec(c, vecBuiltin{{.Category}}GeneratedCases)
}

func (s *testEvaluatorSuite) TestVectorizedBuiltin{{.Category}}FuncGenerated(c *C) {
	testVectorizedBuiltinFunc(c, vecBuiltin{{.Category}}GeneratedCases)
}

func BenchmarkVectorizedBuiltin{{.Category}}EvalOneVecGenerated(b *testing.B) {
	benchmarkVectorizedEvalOneVec(b, vecBuiltin{{.Category}}GeneratedCases)
}

func BenchmarkVectorizedBuiltin{{.Category}}FuncGenerated(b *testing.B) {
	benchmarkVectorizedBuiltinFunc(b, vecBuiltin{{.Category}}GeneratedCases)
}
`))

type sig struct {
	SigName       string
	Input, Output TypeContext
}

var inSigsTmpl = []sig{
	{SigName: "builtinInIntSig", Input: TypeInt, Output: TypeInt},
	{SigName: "builtinInStringSig", Input: TypeString, Output: TypeInt},
	{SigName: "builtinInDecimalSig", Input: TypeDecimal, Output: TypeInt},
	{SigName: "builtinInRealSig", Input: TypeReal, Output: TypeInt},
	{SigName: "builtinInTimeSig", Input: TypeDatetime, Output: TypeInt},
	{SigName: "builtinInDurationSig", Input: TypeDuration, Output: TypeInt},
	{SigName: "builtinInJSONSig", Input: TypeJSON, Output: TypeInt},
}

type function struct {
	FuncName string
	Sigs     []sig
}

var tmplVal = struct {
	Category  string
	Functions []function
}{
	Category: "Other",
	Functions: []function{
		{FuncName: "In", Sigs: inSigsTmpl},
	},
}

func generateDotGo(fileName string) error {
	w := new(bytes.Buffer)
	w.WriteString(header)
	w.WriteString(newLine)
	w.WriteString(builtinOtherImports)
	err := builtinInTmpl.Execute(w, inSigsTmpl)
	if err != nil {
		return err
	}
	data, err := format.Source(w.Bytes())
	if err != nil {
		log.Println("[Warn]", fileName+": gofmt failed", err)
		data = w.Bytes() // write original data for debugging
	}
	return ioutil.WriteFile(fileName, data, 0644)
}

func generateTestDotGo(fileName string) error {
	w := new(bytes.Buffer)
	err := testFile.Execute(w, tmplVal)
	if err != nil {
		return err
	}
	data, err := format.Source(w.Bytes())
	if err != nil {
		log.Println("[Warn]", fileName+": gofmt failed", err)
		data = w.Bytes() // write original data for debugging
	}
	return ioutil.WriteFile(fileName, data, 0644)
}

// generateOneFile generate one xxx.go file and the associated xxx_test.go file.
func generateOneFile(fileNamePrefix string) (err error) {
	err = generateDotGo(fileNamePrefix + ".go")
	if err != nil {
		return
	}
	err = generateTestDotGo(fileNamePrefix + "_test.go")
	return
}

func main() {
	var err error
	outputDir := "."
	err = generateOneFile(filepath.Join(outputDir, "builtin_other_vec_generated"))
	if err != nil {
		log.Fatalln("generateOneFile", err)
	}
}
