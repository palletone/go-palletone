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
	"bytes"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/comm"
	"github.com/palletone/go-palletone/contracts/utils"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	cutil "github.com/palletone/go-palletone/vm/common"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

//var log = flogging.MustGetLogger("util")

//ComputeHash computes contents hash based on previous hash
func ComputeHash(contents []byte, hash []byte) []byte {
	newSlice := make([]byte, len(hash)+len(contents))

	//copy the contents
	copy(newSlice[0:len(contents)], contents[:])

	//add the previous hash
	copy(newSlice[len(contents):], hash[:])

	//compute new hash
	hash = util.ComputeSHA256(newSlice)

	return hash
}

//HashFilesInDir computes h=hash(h,file bytes) for each file in a directory
//Directory entries are traversed recursively. In the end a single
//hash value is returned for the entire directory structure
func HashFilesInDir(rootDir string, dir string, hash []byte, tw *tar.Writer) ([]byte, error) {
	currentDir := filepath.Join(rootDir, dir)
	log.Debugf("hashFiles %s", currentDir)
	//ReadDir returns sorted list of files in dir
	fis, err := ioutil.ReadDir(currentDir)
	if err != nil {
		return hash, fmt.Errorf("ReadDir failed %s\n", err)
	}
	for _, fi := range fis {
		name := filepath.Join(dir, fi.Name())
		if fi.IsDir() {
			var err error
			hash, err = HashFilesInDir(rootDir, name, hash, tw)
			if err != nil {
				return hash, err
			}
			continue
		}
		fqp := filepath.Join(rootDir, name)
		buf, err := ioutil.ReadFile(fqp)
		if err != nil {
			log.Errorf("Error reading %s\n", err)
			return hash, err
		}

		//get the new hash from file contents
		hash = ComputeHash(buf, hash)

		if tw != nil {
			is := bytes.NewReader(buf)
			if err = cutil.WriteStreamToPackage(is, fqp, filepath.Join("src", name), tw); err != nil {
				return hash, fmt.Errorf("Error adding file to tar %s", err)
			}
		}
	}
	return hash, nil
}

//IsCodeExist checks the chaincode if exists
func IsCodeExist(tmppath string) error {
	file, err := os.Open(tmppath)
	if err != nil {
		return fmt.Errorf("Could not open file %s", err)
	}

	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("Could not stat file %s", err)
	}

	if !fi.IsDir() {
		return fmt.Errorf("File %s is not dir\n", file.Name())
	}

	return nil
}

type DockerBuildOptions struct {
	Image        string
	Env          []string
	Cmd          string
	InputStream  io.Reader
	OutputStream io.Writer
}

//-------------------------------------------------------------------------------------------
// DockerBuild
//-------------------------------------------------------------------------------------------
// This function allows a "pass-through" build of chaincode within a docker container as
// an alternative to using standard "docker build" + Dockerfile mechanisms.  The plain docker
// build is somewhat limiting due to the resulting image that is a superset composition of
// the build-time and run-time environments.  This superset can be problematic on several
// fronts, such as a bloated image size, and additional security exposure associated with
// applications that are not needed, etc.
//
// Therefore, this mechanism creates a pipeline consisting of an ephemeral docker
// container that accepts source code as input, runs some function (e.g. "go build"), and
// outputs the result.  The intention is that this output will be consumed as the basis of
// a streamlined container by installing the output into a downstream docker-build based on
// an appropriate minimal image.
//
// The input parameters are fairly simple:
//      - Image:        (optional) The builder image to use or "chaincode.builder"
//      - Env:          (optional) environment variables for the build environment.
//      - Cmd:          The command to execute inside the container.
//      - InputStream:  A tarball of files that will be expanded into /chaincode/input.
//      - OutputStream: A tarball of files that will be gathered from /chaincode/output
//                      after successful execution of Cmd.
//-------------------------------------------------------------------------------------------
func DockerBuild(opts DockerBuildOptions) error {

	client, err := cutil.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return fmt.Errorf("error creating docker client: %s", err)
	}
	//if opts.Image == "" {
	//	//通用的本地编译环境
	//	opts.Image = contractcfg.GetConfig().CommonBuilder //cutil.GetDockerfileFromConfig("chaincode.builder")
	//	if opts.Image == "" {
	//		return fmt.Errorf("No image provided and \"chaincode.builder\" default does not exist")
	//	}
	//}

	log.Debugf("Attempting build with image %s", opts.Image)

	//-----------------------------------------------------------------------------------
	// Ensure the image exists locally, or pull it from a registry if it doesn't
	//确认镜像是否存在或从远程拉取
	//-----------------------------------------------------------------------------------
	_, err = client.InspectImage(opts.Image)
	if err != nil {
		log.Debugf("Image %s does not exist locally, attempt pull", opts.Image)

		err = client.PullImage(docker.PullImageOptions{Repository: opts.Image}, docker.AuthConfiguration{})
		if err != nil {
			return fmt.Errorf("Failed to pull %s: %s", opts.Image, err)
		}
	}

	//-----------------------------------------------------------------------------------
	// Create an ephemeral container, armed with our Env/Cmd
	//创建一个暂时的容器用于链码编译
	//-----------------------------------------------------------------------------------
	dag, err := comm.GetCcDagHand()
	if err != nil {
		log.Debugf("load GetCcDagHand: %s", err.Error())
	}
	cp := dag.GetChainParameters()

	hostConfig := &docker.HostConfig{
		//Memory:     dockercontroller.GetInt64FromDb("TempUccMemory"),     //1GB
		//MemorySwap: dockercontroller.GetInt64FromDb("TempUccMemorySwap"), //1GB
		//CPUShares:  dockercontroller.GetInt64FromDb("TempUccCpuShares"),
		//CPUQuota:   dockercontroller.GetInt64FromDb("TempUccCpuQuota"),
		Memory:    cp.TempUccMemory, //1GB
		CPUShares: cp.TempUccCpuShares,
		CPUQuota:  cp.TempUccCpuQuota,
	}
	log.Infof("client.CreateContainer")
	container, err := client.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        opts.Image,
			Env:          opts.Env,
			Cmd:          []string{"/bin/sh", "-c", opts.Cmd},
			AttachStdout: true,
			AttachStderr: true,
		},
		HostConfig: hostConfig,
	})
	if err != nil {
		return fmt.Errorf("Error creating container: %s", err)
	}
	defer client.RemoveContainer(docker.RemoveContainerOptions{ID: container.ID})

	//-----------------------------------------------------------------------------------
	// Upload our input stream
	//上传输入
	//-----------------------------------------------------------------------------------
	log.Infof("client.UploadToContainer")
	err = client.UploadToContainer(container.ID, docker.UploadToContainerOptions{
		Path:        "/chaincode/input", //  /chaincode/input
		InputStream: opts.InputStream,
	})
	if err != nil {
		return fmt.Errorf("Error uploading input to container: %s", err)
	}

	//-----------------------------------------------------------------------------------
	// Attach stdout buffer to capture possible compilation errors
	//-----------------------------------------------------------------------------------
	stdout := bytes.NewBuffer(nil)
	log.Infof("client.AttachToContainerNonBlocking")
	_, err = client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    container.ID,
		OutputStream: stdout,
		ErrorStream:  stdout,
		Logs:         true,
		Stdout:       true,
		Stderr:       true,
		Stream:       true,
	})
	if err != nil {
		return fmt.Errorf("Error attaching to container: %s", err)
	}

	//-----------------------------------------------------------------------------------
	// Launch the actual build, realizing the Env/Cmd specified at container creation
	//启动容器
	//-----------------------------------------------------------------------------------
	log.Infof("client.StartContainer")
	err = client.StartContainer(container.ID, nil)
	if err != nil {
		return fmt.Errorf("Error executing build: %s \"%s\"", err, stdout.String())
	}
	//解决临时容器一直运行的情况
	go utils.RemoveContainerWhenGoBuildTimeOut(container.ID)

	//-----------------------------------------------------------------------------------
	// Wait for the build to complete and gather the return value
	//等待容器返回
	//-----------------------------------------------------------------------------------
	log.Infof("client.WaitContainer")
	retval, err := client.WaitContainer(container.ID)
	if err != nil {
		return fmt.Errorf("Error waiting for container to complete: %s", err)
	}
	if retval > 0 {
		return fmt.Errorf("Error returned from build: %d \"%s\"", retval, stdout.String())
	}

	//-----------------------------------------------------------------------------------
	// Finally, download the result
	//获取容器输出
	//-----------------------------------------------------------------------------------
	log.Infof("client.DownloadFromContainer")
	err = client.DownloadFromContainer(container.ID, docker.DownloadFromContainerOptions{
		Path:         "/chaincode/output/.",
		OutputStream: opts.OutputStream,
	})
	if err != nil {
		return fmt.Errorf("Error downloading output: %s", err)
	}

	return nil
}
