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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */
//File and Folder operations
package files

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//Check file Is Exist
//判断文件是否存在，存在返回true，不存在返回false
func IsExist(path string) bool {
	var exist = true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

//Check path is a folder or not
// 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

//Check path is a file or not
// 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

//Check folder is empty or not
func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

//Remove this file ,and if the parent folder is empty remove parent folder.
func RemoveFileAndEmptyFolder(filePath string) error {
	if IsFile(filePath) {
		if IsExist(filePath) {
			os.Remove(filePath)
		}
	} else {
		if emp, _ := IsEmpty(filePath); emp {
			os.Remove(filePath)
		} else {
			return nil
		}
	}
	var parentFolder = filepath.Dir(filePath)
	return RemoveFileAndEmptyFolder(parentFolder)
}

//MakeDirAndFile the path of file ,if the path is not exist.
//if the path is stdout or stderr, don't create file
func MakeDirAndFile(filePath string) error {
	if filePath == "stdout" || filePath == "stderr" {
		return nil
	}
	// log.Println("log file path:" + filePath)
	if !IsExist(filePath) {
		// log.Println("create folder and file:" + filePath)
		err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			return err
		}
		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

	}
	return nil
}

//Copy on file
func CopyFile(dst, src string) error {
	if IsDir(src) {
		return errors.New("CopyFile function only can copy file, but your path is a folder:" + src)
	}
	rf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer rf.Close()
	rstat, err := rf.Stat()
	if err != nil {
		return err
	}

	wf, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, rstat.Mode())
	if err != nil {
		return err
	}
	if _, err := io.Copy(wf, rf); err != nil {
		wf.Close()
		return err
	}
	return wf.Close()
}

//CopyAll Copy all files from source to destination
func CopyAll(dst, src string) error {
	return filepath.Walk(src, makeFileCopyWalkFunc(dst, src))
}
func makeFileCopyWalkFunc(dst, src string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, strings.TrimPrefix(path, src))
		if info.IsDir() {
			return os.Mkdir(dstPath, info.Mode())
		}
		return CopyFile(dstPath, path)
	}
}
