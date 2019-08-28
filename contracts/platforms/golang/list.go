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
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common/log"
)

//runProgram non-nil Env, timeout (typically secs or millisecs), program name and args
func runProgram(env Env, timeout time.Duration, pgm string, args ...string) ([]byte, error) {
	if env == nil {
		return nil, fmt.Errorf("<%s, %v>: nil env provided", pgm, args)
	}
	var stdOut bytes.Buffer
	var stdErr bytes.Buffer

	cmd := exec.Command(pgm, args...)
	cmd.Env = flattenEnv(env)
	cmd.Stdout = &stdOut
	cmd.Stderr = &stdErr
	cmd.Start()
	var err error

	// Create a go routine that will wait for the command to finish
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		if err = cmd.Process.Kill(); err != nil {
			return nil, fmt.Errorf("<%s, %v>: failed to kill: %s", pgm, args, err)
		} else {
			return nil, fmt.Errorf("<%s, %v>: timeout(%d msecs)", pgm, args, timeout/time.Millisecond)
		}
	case err = <-done:
		if err != nil {
			return nil, fmt.Errorf("<%s, %v>: failed with error: \"%s\"\n%s", pgm, args, err, stdErr.String())
		}

		return stdOut.Bytes(), nil
	}
}

// Logic inspired by: https://dave.cheney.net/2014/09/14/go-list-your-swiss-army-knife
func list(env Env, template, pkg string) ([]string, error) {
	if env == nil {
		env = getEnv()
	}

	log.Infof("template[%v],pkg[%s]", template, pkg)
	lst, err := runProgram(env, 60*time.Second, "go", "list", "-f", template, pkg)
	if err != nil {
		return nil, err
	}

	return strings.Split(strings.Trim(string(lst), "\n"), "\n"), nil
}

func listDeps(env Env, pkg string) ([]string, error) {
	return list(env, "{{ join .Deps \"\\n\"}}", pkg)
}

func listImports(env Env, pkg string) ([]string, error) {
	return list(env, "{{ join .Imports \"\\n\"}}", pkg)
}
