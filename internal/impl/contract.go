// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.

// You should have received a copy of the GNU Lesser General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package impl

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"text/template"

	"github.com/getamis/grpc-contract/internal/util"
	"golang.org/x/tools/imports"
)

type Contract struct {
	Package    string
	Name       string
	Methods    Methods
	StructName string
	Sources    Sources
}

func NewContract(pack string, name string, sources Sources) Contract {
	c := Contract{
		Package: pack,
		Name:    name,
		Sources: make([]string, len(sources)),
	}
	c.StructName = strings.ToLower(string(c.Name[0])) + c.Name[1:len(c.Name)]
	for i, s := range sources {
		_, c.Sources[i] = path.Split(s)
	}
	return c
}

func (c *Contract) IsServerInterface(name string) bool {
	if name == c.Name+"Server" {
		return true
	}
	return false
}

var ContractTemplate string = `// Automatically generated by grpc-contract. DO NOT EDIT!
// sources: {{ range .Sources }}
//     {{ . }}
{{- end }}

package {{ .Package }};

type {{ .StructName }} struct {
	contract *{{ .Name }}
	transactOptsFn TransactOptsFn
}

func New{{ .Name }}Server(address common.Address, backend bind.ContractBackend, transactOptsFn TransactOptsFn) {{ .Name }}Server {
	contract, _ := New{{ .Name }}(address, backend)
	service := &{{ .StructName }}{
		contract:     contract,
		transactOptsFn: transactOptsFn,
	}
	if transactOptsFn == nil {
		service.transactOptsFn = DefaultTransactOptsFn
	}
	return service
}

{{ range .Methods }}
{{ . }}
{{ end }}
`

func (c *Contract) Write(filepath, filename string) {
	sort.Sort(c.Sources)
	sort.Sort(c.Methods)
	implTemplate, err := template.New("contract").Parse(ContractTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		os.Exit(-1)
	}
	result := new(bytes.Buffer)
	err = implTemplate.Execute(result, c)
	if err != nil {
		fmt.Printf("Failed to render template: %v\n", err)
		os.Exit(-1)
	}
	code, err := imports.Process(".", result.Bytes(), nil)
	if err != nil {
		fmt.Printf("Failed to process code: %v\n", err)
		os.Exit(-1)
	}
	util.WriteFile(string(code), filepath, filename)
}

type Sources []string

// Len is part of sort.Interface.
func (s Sources) Len() int {
	return len(s)
}

// Swap is part of sort.Interface.
func (s Sources) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less is part of sort.Interface.
func (s Sources) Less(i, j int) bool {
	return strings.Compare(s[i], s[j]) < 0
}
