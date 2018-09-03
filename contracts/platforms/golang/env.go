/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package golang

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Env map[string]string

func getEnv() Env {
	env := make(Env)
	for _, entry := range os.Environ() {
		tokens := strings.SplitN(entry, "=", 2)
		if len(tokens) > 1 {
			env[tokens[0]] = tokens[1]
		}
	}

	return env
}

func getGoEnv() (Env, error) {
	env := getEnv()

	goenvbytes, err := runProgram(env, 10*time.Second, "go", "env")
	if err != nil {
		return nil, err
	}

	goenv := make(Env)

	envout := strings.Split(string(goenvbytes), "\n")
	for _, entry := range envout {
		tokens := strings.SplitN(entry, "=", 2)
		if len(tokens) > 1 {
			goenv[tokens[0]] = strings.Trim(tokens[1], "\"")
		}
	}

	return goenv, nil
}

func flattenEnv(env Env) []string {
	result := make([]string, 0)
	for k, v := range env {
		result = append(result, k+"="+v)
	}

	return result
}

type Paths map[string]bool

func splitEnvPaths(value string) Paths {
	_paths := filepath.SplitList(value)
	paths := make(Paths)
	for _, path := range _paths {
		paths[path] = true
	}
	return paths
}

func flattenEnvPaths(paths Paths) string {

	_paths := make([]string, 0)
	for path := range paths {
		_paths = append(_paths, path)
	}

	return strings.Join(_paths, string(os.PathListSeparator))
}
