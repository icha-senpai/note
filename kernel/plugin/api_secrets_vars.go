// Scribli - Refactor your thinking
// Copyright (c) 2020-present, Scribli
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package plugin

import (
	"fmt"

	"github.com/icha-senpai/note/kernel/model"
	"github.com/icha-senpai/note/third_party/forks/github/dop251/goja"
	"github.com/icha-senpai/note/third_party/forks/github/samber/lo"
)

// injectSecretsVars adds scribli.secrets and scribli.vars to the plugin JS sandbox.

func injectSecretsVars(p *KernelPlugin, rt *goja.Runtime, scribli *goja.Object) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("injectSecretsVars: %v", r)
		}
	}()

	secrets := rt.NewObject()
	lo.Must0(secrets.Set("resolve", rt.ToValue(makeResolver(func(tpl string) string {
		return model.Conf.Secrets.Resolve(tpl)
	}))))
	lo.Must0(scribli.Set("secrets", secrets))

	vars := rt.NewObject()
	lo.Must0(vars.Set("resolve", rt.ToValue(makeResolver(func(tpl string) string {
		return model.Conf.Variables.Resolve(tpl)
	}))))
	lo.Must0(scribli.Set("vars", vars))

	return
}

func makeResolver(resolver func(string) string) func(goja.FunctionCall, *goja.Runtime) goja.Value {
	return func(call goja.FunctionCall, rt *goja.Runtime) goja.Value {
		if len(call.Arguments) < 1 || !goja.IsString(call.Argument(0)) {
			return rt.ToValue("")
		}
		return rt.ToValue(resolver(call.Argument(0).String()))
	}
}
