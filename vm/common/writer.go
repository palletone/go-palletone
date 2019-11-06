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

package util

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common/log"
	"github.com/pkg/errors"
)

//var log = flogging.MustGetLogger("container")

// These filetypes are excluded while creating the tar package sent to Docker
// Generated .class and other temporary files can be excluded
var javaExcludeFileTypes = map[string]bool{
	".class": true,
}

// WriteFolderToTarPackage writes source files to a tarball.
// This utility is used for node js chaincode packaging, but not golang chaincode.
// Golang chaincode has more sophisticated file packaging, as implemented in golang/platform.go.
func WriteFolderToTarPackage(tw *tar.Writer, srcPath string, excludeDir string, includeFileTypeMap map[string]bool, excludeFileTypeMap map[string]bool) error {
	fileCount := 0
	rootDirectory := srcPath

	// trim trailing slash if it was passed
	if rootDirectory[len(rootDirectory)-1] == '/' {
		rootDirectory = rootDirectory[:len(rootDirectory)-1]
	}

	log.Infof("rootDirectory = %s", rootDirectory)

	//append "/" if necessary
	if excludeDir != "" && strings.LastIndex(excludeDir, "/") < len(excludeDir)-1 {
		excludeDir = excludeDir + "/"
	}

	rootDirLen := len(rootDirectory)
	walkFn := func(localpath string, info os.FileInfo, err error) error {
		log.Debugf("err %s", err.Error())
		// If localpath includes .git, ignore
		if strings.Contains(localpath, ".git") {
			return nil
		}

		if info.Mode().IsDir() {
			return nil
		}

		//exclude any files with excludeDir prefix. They should already be in the tar
		if excludeDir != "" && strings.Index(localpath, excludeDir) == rootDirLen+1 {
			//1 for "/"
			return nil
		}
		// Because of scoping we can reference the external rootDirectory variable
		if len(localpath[rootDirLen:]) == 0 {
			return nil
		}
		ext := filepath.Ext(localpath)

		if includeFileTypeMap != nil {
			// we only want 'fileTypes' source files at this point
			if _, ok := includeFileTypeMap[ext]; !ok {
				return nil
			}
		}

		//exclude the given file types
		if excludeFileTypeMap != nil {
			if exclude, ok := excludeFileTypeMap[ext]; ok && exclude {
				return nil
			}
		}

		var packagepath string

		// if file is metadata, keep the /META-INF directory, e.g: META-INF/statedb/couchdb/indexes/indexOwner.json
		// otherwise file is source code, put it in /src dir, e.g: src/marbles_chaincode.js
		if strings.HasPrefix(localpath, filepath.Join(rootDirectory, "META-INF")) {
			//packagepath = localpath[rootDirLen+1:]
			//
			//// Split the tar packagepath into a tar package directory and filename
			//packageDir, filename := filepath.Split(packagepath)
			//
			//// Hidden files are not supported as metadata, therefore ignore them.
			//// User often doesn't know that hidden files are there, and may not be able to delete them, therefore warn user rather than error out.
			//if strings.HasPrefix(filename, ".") {
			//	log.Warningf("Ignoring hidden file in metadata directory: %s", packagepath)
			//	return nil
			//}
			//
			//fileBytes, err := ioutil.ReadFile(localpath)
			//if err != nil {
			//	return err
			//}
			//
			//// Validate metadata file for inclusion in tar
			//// Validation is based on the passed metadata directory, e.g. META-INF/statedb/couchdb/indexes
			//// Clean metadata directory to remove trailing slash
			////glh
			//err = metadata.ValidateMetadataFile(filename, fileBytes, filepath.Clean(packageDir))
			//if err != nil {
			//	return err
			//}

		} else { // file is not metadata, include in src
			packagepath = fmt.Sprintf("src%s", localpath[rootDirLen:])
		}

		err = WriteFileToPackage(localpath, packagepath, tw)
		if err != nil {
			return fmt.Errorf("Error writing file to package: %s", err)
		}
		fileCount++

		return nil
	}

	if err := filepath.Walk(rootDirectory, walkFn); err != nil {
		log.Infof("Error walking rootDirectory: %s", err)
		return err
	}
	// return error if no files were found
	if fileCount == 0 {
		return errors.Errorf("no source files found in '%s'", srcPath)
	}
	return nil
}

//Package Java project to tar file from the source path
func WriteJavaProjectToPackage(tw *tar.Writer, srcPath string) error {

	log.Debugf("Packaging Java project from path %s", srcPath)

	if err := WriteFolderToTarPackage(tw, srcPath, "", nil, javaExcludeFileTypes); err != nil {

		log.Errorf("Error writing folder to tar package %s", err)
		return err
	}
	// Write the tar file out
	if err := tw.Close(); err != nil {
		return err
	}
	return nil

}

//WriteFileToPackage writes a file to the tarball
func WriteFileToPackage(localpath string, packagepath string, tw *tar.Writer) error {
	//log.Debug("Writing file to tarball:", packagepath)
	fd, err := os.Open(localpath)
	if err != nil {
		return fmt.Errorf("%s: %s", localpath, err)
	}
	defer fd.Close()

	is := bufio.NewReader(fd)
	return WriteStreamToPackage(is, localpath, packagepath, tw)

}

//WriteStreamToPackage writes bytes (from a file reader) to the tarball
func WriteStreamToPackage(is io.Reader, localPath string, packagePath string, tw *tar.Writer) error { //localpath string, packagepath
	localpath := strings.Replace(localPath, "\\", "/", -1)
	packagepath := strings.Replace(packagePath, "\\", "/", -1)

	info, err := os.Stat(localpath)
	if err != nil {
		return fmt.Errorf("%s: %s", localpath, err)
	}
	header, err := tar.FileInfoHeader(info, localpath)
	if err != nil {
		return fmt.Errorf("Error getting FileInfoHeader: %s", err)
	}

	//Let's take the variance out of the tar, make headers identical by using zero time
	oldname := header.Name
	var zeroTime time.Time
	header.AccessTime = zeroTime
	header.ModTime = zeroTime
	header.ChangeTime = zeroTime
	header.Name = packagepath
	header.Mode = 0100644
	header.Uid = 500
	header.Gid = 500

	if err = tw.WriteHeader(header); err != nil {
		return fmt.Errorf("Error write header for (path: %s, oldname:%s,newname:%s,sz:%d) : %s", localpath, oldname, packagepath, header.Size, err)
	}
	if _, err := io.Copy(tw, is); err != nil {
		return fmt.Errorf("Error copy (path: %s, oldname:%s,newname:%s,sz:%d) : %s", localpath, oldname, packagepath, header.Size, err)
	}

	return nil
}

func WriteBytesToPackage(name string, payload []byte, tw *tar.Writer) error {
	//Make headers identical by using zero time
	var zeroTime time.Time
	tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(payload)), ModTime: zeroTime, AccessTime: zeroTime, ChangeTime: zeroTime})
	tw.Write(payload)

	return nil
}
