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

package dockercontroller

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
	container "github.com/palletone/go-palletone/vm/api"
	"github.com/palletone/go-palletone/vm/ccintf"
	com "github.com/palletone/go-palletone/vm/common"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/comm"
	"strconv"
	"github.com/fsouza/go-dockerclient"
)

var (
	//log = flogging.MustGetLogger("dockercontroller")
	hostConfig  *docker.HostConfig
	vmRegExp    = regexp.MustCompile("[^a-zA-Z0-9-_.]")
	imageRegExp = regexp.MustCompile("^[a-z0-9]+(([._-][a-z0-9]+)+)?$")
)

func getClientMe() (dockerClient, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	return client, err
}

// getClient returns an instance that implements dockerClient interface
type getClient func() (dockerClient, error)

//DockerVM is a vm. It is identified by an image id
type DockerVM struct {
	id           string
	getClientFnc getClient
}

// dockerClient represents a docker client
type dockerClient interface {
	// CreateContainer creates a docker container, returns an error in case of failure
	CreateContainer(opts docker.CreateContainerOptions) (*docker.Container, error)
	// UploadToContainer uploads a tar archive to be extracted to a path in the
	// filesystem of the container.
	UploadToContainer(id string, opts docker.UploadToContainerOptions) error
	// StartContainer starts a docker container, returns an error in case of failure
	StartContainer(id string, cfg *docker.HostConfig) error
	// AttachToContainer attaches to a docker container, returns an error in case of
	// failure
	AttachToContainer(opts docker.AttachToContainerOptions) error
	// BuildImage builds an image from a tarball's url or a Dockerfile in the input
	// stream, returns an error in case of failure
	BuildImage(opts docker.BuildImageOptions) error
	// RemoveImageExtended removes a docker image by its name or ID, returns an
	// error in case of failure
	RemoveImageExtended(id string, opts docker.RemoveImageOptions) error
	// StopContainer stops a docker container, killing it after the given timeout
	// (in seconds). Returns an error in case of failure
	StopContainer(id string, timeout uint) error
	// KillContainer sends a signal to a docker container, returns an error in
	// case of failure
	KillContainer(opts docker.KillContainerOptions) error
	// RemoveContainer removes a docker container, returns an error in case of failure
	RemoveContainer(opts docker.RemoveContainerOptions) error
}

// NewDockerVM returns a new DockerVM instance
func NewDockerVM() *DockerVM {
	vm := DockerVM{}
	vm.getClientFnc = getDockerClient
	return &vm
}

func getDockerClient() (dockerClient, error) {
	return com.NewDockerClient()
}

func GetInt64FromDb(key string) int64 {
	//DefaultUccMemory  = "104857600" //物理内存  104857600  100m
	//DefaultUccMemorySwap  = "104857600"//内存交换区，不设置默认为memory的两倍
	//DefaultUccCpuShare  = "1024"//CPU占用率，相对的  CPU 利用率权重，默认为 1024
	//DefaultCpuPeriod  = "50000"// 限制CPU --cpu-period=50000 --cpu-quota=25000
	//DefaultUccCpuQuota  = "25000"//限制CPU 周期设为 50000，将容器在每个周期内的 CPU 配额设置为 25000，表示该容器每 50ms 可以得到 50% 的 CPU 运行时间
	//DefaultUccCpuSetCpus  = "0-3"//限制使用某些CPUS  "1,3"  "0-3"
	dag, err := comm.GetCcDagHand()
	resultStr, _, err := dag.GetConfig(key)
	if err != nil {
		log.Infof("dag.GetConfig err: %s", err.Error())
		return 0
	}
	resultInt64, err := strconv.ParseInt(string(resultStr), 10, 64)
	if err != nil {
		log.Infof("strconv.ParseInt err: %s", err.Error())
		return 0
	}
	return resultInt64
}

//TODO
func getDockerHostConfig() *docker.HostConfig {

	if hostConfig != nil {
		return hostConfig
	}
	dockerKey := func(key string) string {
		return "vm.docker.hostConfig." + key
	}
	getInt64 := func(key string) int64 {
		defer func() {
			if err := recover(); err != nil {
				log.Debugf("load vm.docker.hostConfig.%s failed, error: %v", key, err)
			}
		}()
		n := viper.GetInt(dockerKey(key))
		return int64(n)
	}

	var logConfig docker.LogConfig
	err := viper.UnmarshalKey(dockerKey("LogConfig"), &logConfig)
	if err != nil {
		log.Debugf("load docker HostConfig.LogConfig failed, error: %s", err.Error())
	}
	networkMode := viper.GetString(dockerKey("NetworkMode"))
	if networkMode == "" {
		//networkMode = "bridge"
		networkMode = "host"
	}
	log.Debugf("docker container hostconfig NetworkMode: %s", networkMode)
	portBindings := make(map[docker.Port][]docker.PortBinding)
	hosts := strings.Split(contractcfg.GetConfig().ContractAddress, ":")

	portBinding := docker.PortBinding{
		HostIP:   hosts[0],
		HostPort: hosts[1],
	}
	portBindings[docker.Port(hosts[1]+"/tcp")] = []docker.PortBinding{portBinding}
	hostConfig = &docker.HostConfig{
		CapAdd:       viper.GetStringSlice(dockerKey("CapAdd")),
		CapDrop:      viper.GetStringSlice(dockerKey("CapDrop")),
		PortBindings: portBindings,
		DNS:          viper.GetStringSlice(dockerKey("Dns")),
		DNSSearch:    viper.GetStringSlice(dockerKey("DnsSearch")),
		ExtraHosts:   viper.GetStringSlice(dockerKey("ExtraHosts")),
		NetworkMode:  networkMode,
		IpcMode:      viper.GetString(dockerKey("IpcMode")),
		PidMode:      viper.GetString(dockerKey("PidMode")),
		UTSMode:      viper.GetString(dockerKey("UTSMode")),
		LogConfig:    logConfig,

		ReadonlyRootfs: viper.GetBool(dockerKey("ReadonlyRootfs")),
		SecurityOpt:    viper.GetStringSlice(dockerKey("SecurityOpt")),
		CgroupParent:   viper.GetString(dockerKey("CgroupParent")),
		Memory:         GetInt64FromDb("UccMemory"),
		MemorySwap:     GetInt64FromDb("UccMemorySwap"),
		//Memory:           int64(104857600), //100mB
		//MemorySwap:       int64(20971520),
		MemorySwappiness: getInt64("MemorySwappiness"),
		OOMKillDisable:   viper.GetBool(dockerKey("OomKillDisable")),
		CPUShares:        GetInt64FromDb("UccCpuShares"),
		CPUSet:           viper.GetString(dockerKey("Cpuset")),
		CPUSetCPUs:       viper.GetString(dockerKey("CpusetCPUs")),
		CPUSetMEMs:       viper.GetString(dockerKey("CpusetMEMs")),
		CPUQuota:         GetInt64FromDb("UccCpuQuota"),
		CPUPeriod:        GetInt64FromDb("UccCpuPeriod"),
		BlkioWeight:      getInt64("BlkioWeight"),
	}

	return hostConfig
}

func (vm *DockerVM) createContainer(ctxt context.Context, client dockerClient,
	imageID string, containerID string, args []string,
	env []string, attachStdout bool) error {
	config := docker.Config{Cmd: args, Image: imageID, Env: env, AttachStdout: attachStdout, AttachStderr: attachStdout}
	copts := docker.CreateContainerOptions{Name: containerID, Config: &config, HostConfig: getDockerHostConfig()}
	log.Debugf("Create container: %s", containerID)
	_, err := client.CreateContainer(copts)
	if err != nil {
		return err
	}
	//log.Debugf("Created container: %s", imageID)
	return nil
}

func (vm *DockerVM) deployImage(client dockerClient, ccid ccintf.CCID,
	args []string, env []string, reader io.Reader) error {
	id, err := vm.GetImageId(ccid)
	if err != nil {
		return err
	}
	outputbuf := bytes.NewBuffer(nil)
	opts := docker.BuildImageOptions{
		Name:         id,
		Pull:         viper.GetBool("chaincode.pull"),
		InputStream:  reader,
		OutputStream: outputbuf,
	}
	//glh
	/*
		opts := docker.BuildImageOptions{
			Name:         "ubuntu",
			Pull:         false,
			InputStream:  reader,
			OutputStream: outputbuf,
		}
	*/
	if err := client.BuildImage(opts); err != nil {
		log.Errorf("Error building images: %s", err)
		log.Errorf("Image Output:\n********************\n%s\n********************", outputbuf.String())
		return err
	}

	log.Debugf("Created image: %s", id)

	return nil
}

//Deploy use the reader containing targz to create a docker image
//for docker inputbuf is tar reader ready for use by docker.Client
//the stream from end client to peer could directly be this tar stream
//talk to docker daemon using docker Client and build the image
func (vm *DockerVM) Deploy(ctxt context.Context, ccid ccintf.CCID,
	args []string, env []string, reader io.Reader) error {

	client, err := vm.getClientFnc()
	switch err {
	case nil:
		if err = vm.deployImage(client, ccid, args, env, reader); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Error creating docker client: %s", err)
	}
	return nil
}

//Start starts a container using a previously created docker image
//根据之前指定的镜像文件启动容器，如果镜像文件不存在则新创建，成功后启动容器
//这里还可以指定对容器日志的输出
func (vm *DockerVM) Start(ctxt context.Context, ccid ccintf.CCID,
	args []string, env []string, filesToUpload map[string][]byte, builder container.BuildSpecFactory, prelaunchFunc container.PrelaunchFunc) error {
	//获取本地基础镜像
	imageID, err := vm.GetImageId(ccid)
	if err != nil {
		log.Errorf("get image id error: %s", err)
		return err
	}
	//获取docker客户端
	client, err := vm.getClientFnc()
	if err != nil {
		log.Debugf("start - cannot create client: ", err)
		return err
	}

	containerID, err := vm.GetContainerId(ccid)
	if err != nil {
		log.Debugf("get container %s error: ", containerID, err)
		return err
	}

	attachStdout := viper.GetBool("vm.docker.attachStdout")

	//stop,force remove if necessary
	log.Debugf("Cleanup container %s", containerID)
	//停止容器
	//err = vm.stopInternal(ctxt, client, containerID, 0, false, false)
	//if err != nil {
	//	return err
	//}
	//创建容器
	log.Debugf("Start container %s", containerID)
	err = vm.createContainer(ctxt, client, imageID, containerID, args, env, attachStdout)
	//var reader io.Reader
	//var err1 error
	//var isInit = false
	if err != nil {
		//if image not found try to create image and retry
		if err == docker.ErrNoSuchImage {
			log.Debugf("start-could not find image <%s> (container id <%s>), because of <%s>..."+
				"attempt to recreate image", imageID, containerID, err)
			log.Errorf("no such base image with image name is %s, should pull this image from docker hub.", imageID)
			//-----------------------------------------------------------------------------------
			// Ensure the image exists locally, or pull it from a registry if it doesn't
			//确认镜像是否存在或从远程拉取
			//-----------------------------------------------------------------------------------
			client1, err := com.NewDockerClient()
			_, err = client1.InspectImage(imageID)
			if err != nil {
				log.Debugf("Image %s does not exist locally, attempt pull", imageID)

				err = client1.PullImage(docker.PullImageOptions{Repository: imageID}, docker.AuthConfiguration{})
				if err != nil {
					return fmt.Errorf("Failed to pull %s: %s", imageID, err)
				}
			}
			err = vm.createContainer(ctxt, client, imageID, containerID, args, env, attachStdout)
			if err != nil {
				return fmt.Errorf("no such base image with image name is %s, should pull this image from docker hub.", imageID)
			}
			//if builder != nil {
			//	log.Debugf("start-could not find image <%s> (container id <%s>), because of <%s>..."+
			//		"attempt to recreate image", imageID, containerID, err)
			//	//获取上传container的文件
			//	reader, err1 = builder()
			//	if err1 != nil {
			//		log.Errorf("Error creating image builder for image <%s> (container id <%s>), "+
			//			"because of <%s>", imageID, containerID, err1)
			//	}
			//	////创建镜像
			//	if err1 = vm.deployImage(client, ccid, args, env, reader); err1 != nil {
			//		return err1
			//	}
			//	//imageID = "2bdd5196cd89"
			//	//args = []string{"/bin/sh","-c","cd / && tar binpackage.tar && mv chaincode /usr/local/bin && cd /usr/local/bin && ./chaincode"}
			//	isInit = true
			//	log.Debug("start-recreated image successfully")
			//	//再次创建容器
			//	if err1 = vm.createContainer(ctxt, client, imageID, containerID, args, env, attachStdout); err1 != nil {
			//		log.Errorf("start-could not recreate container post recreate image: %s", err1)
			//		return err1
			//	}
			//} else {
			//	log.Errorf("start-could not find image <%s>, because of %s", imageID, err)
			//	return err
			//}
		} else {
			log.Errorf("start-could not recreate container <%s> with image id <%s>, because of %s", containerID, imageID, err)
			return err
		}
	}

	if attachStdout {
		// Launch a few go-threads to manage output streams from the container.
		// They will be automatically destroyed when the container exits
		attached := make(chan struct{})
		r, w := io.Pipe()

		go func() {
			// AttachToContainer will fire off a message on the "attached" channel once the
			// attachment completes, and then block until the container is terminated.
			// The returned error is not used outside the scope of this function. Assign the
			// error to a local variable to prevent clobbering the function variable 'err'.
			err := client.AttachToContainer(docker.AttachToContainerOptions{
				Container:    containerID,
				OutputStream: w,
				ErrorStream:  w,
				Logs:         true,
				Stdout:       true,
				Stderr:       true,
				Stream:       true,
				Success:      attached,
			})

			// If we get here, the container has terminated.  Send a signal on the pipe
			// so that downstream may clean up appropriately
			_ = w.CloseWithError(err)
		}()

		go func() {
			// Block here until the attachment completes or we timeout
			select {
			case <-attached:
				// successful attach
			case <-time.After(10 * time.Second):
				log.Errorf("Timeout while attaching to IO channel in container %s", containerID)
				return
			}

			// Acknowledge the attachment?  This was included in the gist I followed
			// (http://bit.ly/2jBrCtM).  Not sure it's actually needed but it doesn't
			// appear to hurt anything.
			attached <- struct{}{}

			// Establish a buffer for our IO channel so that we may do readline-style
			// ingestion of the IO, one log entry per line
			is := bufio.NewReader(r)

			// Acquire a custom logger for our chaincode, inheriting the level from the peer
			//containerLogger := flogging.MustGetLogger(containerID)
			//logging.SetLevel(logging.GetLevel("peer"), containerID)

			for {
				// Loop forever dumping lines of text into the containerLogger
				// until the pipe is closed
				line, err2 := is.ReadString('\n')
				if err2 != nil {
					switch err2 {
					case io.EOF:
						log.Debugf("Container %s has closed its IO channel", containerID)
					default:
						log.Errorf("Error reading container output: %s", err2)
					}

					return
				}

				log.Debugf(line)
			}
		}()
	}

	// upload specified files to the container before starting it
	// this can be used for configurations such as TLS key and certs
	if len(filesToUpload) != 0 {
		// the docker upload API takes a tar file, so we need to first
		// consolidate the file entries to a tar
		payload := bytes.NewBuffer(nil)
		gw := gzip.NewWriter(payload)
		tw := tar.NewWriter(gw)

		for path, fileToUpload := range filesToUpload {
			com.WriteBytesToPackage(path, fileToUpload, tw)
		}

		// Write the tar file out
		if err = tw.Close(); err != nil {
			return fmt.Errorf("Error writing files to upload to Docker instance into a temporary tar blob: %s", err)
		}

		gw.Close()

		err = client.UploadToContainer(containerID, docker.UploadToContainerOptions{
			InputStream:          bytes.NewReader(payload.Bytes()),
			Path:                 "/",
			NoOverwriteDirNonDir: false,
		})

		if err != nil {
			return fmt.Errorf("Error uploading files to the container instance %s: %s", containerID, err)
		}
	}

	if prelaunchFunc != nil {
		if err = prelaunchFunc(); err != nil {
			return err
		}
	}
	//获取上传container的文件
	reader, err1 := builder()
	if err1 != nil {
		log.Errorf("Error creating image builder for image <%s> (container id <%s>), "+
			"because of <%s>", imageID, containerID, err1)
		return err1
	}
	//上传文件到容器
	err = client.UploadToContainer(containerID, docker.UploadToContainerOptions{
		InputStream:          reader,
		Path:                 "/",
		NoOverwriteDirNonDir: false,
	})
	if err != nil {
		return fmt.Errorf("Error uploading files to the container instance %s: %s", containerID, err)
	}
	// start container with HostConfig was deprecated since v1.10 and removed in v1.2
	err = client.StartContainer(containerID, nil)
	if err != nil {
		log.Errorf("start-could not start container: %s", err)
		return err
	}

	log.Debugf("Started container %s", containerID)
	return nil
}

//Stop stops a running chaincode
func (vm *DockerVM) Stop(ctxt context.Context, ccid ccintf.CCID, timeout uint, dontkill bool, dontremove bool) error {
	containerId, err := vm.GetContainerId(ccid)
	if err != nil {
		log.Errorf("get image id error: %s", err)
		return err
	}
	client, err := vm.getClientFnc()
	if err != nil {
		log.Debugf("stop - cannot create client %s", err)
		return err
	}
	//id = strings.Replace(id, ":", "_", -1)

	err = vm.stopInternal(ctxt, client, containerId, timeout, dontkill, dontremove)

	return err
}

func (vm *DockerVM) stopInternal(ctxt context.Context, client dockerClient,
	id string, timeout uint, dontkill bool, dontremove bool) error {
	err := client.StopContainer(id, timeout)
	if err != nil {
		log.Debugf("Stop container %s(%s)", id, err)
	} else {
		log.Debugf("Stopped container %s", id)
	}
	if !dontkill {
		err = client.KillContainer(docker.KillContainerOptions{ID: id})
		if err != nil {
			log.Debugf("Kill container %s (%s)", id, err)
		} else {
			log.Debugf("Killed container %s", id)
		}
	}
	if !dontremove {
		err = client.RemoveContainer(docker.RemoveContainerOptions{ID: id, Force: true})
		if err != nil {
			log.Debugf("Remove container %s (%s)", id, err)
		} else {
			log.Debugf("Removed container %s", id)
		}
	}
	return err
}

//Destroy destroys an image
func (vm *DockerVM) Destroy(ctxt context.Context, ccid ccintf.CCID, force bool, noprune bool) error {
	imageID, err := vm.GetImageId(ccid)
	if err != nil {
		log.Errorf("get image id error: %s", err)
		return err
	}
	log.Debugf("image id[%s]", imageID)

	client, err := vm.getClientFnc()
	if err != nil {
		log.Errorf("destroy-cannot create client %s", err)
		return err
	}
	//id = strings.Replace(id, ":", "_", -1)

	err = client.RemoveImageExtended(imageID, docker.RemoveImageOptions{Force: force, NoPrune: noprune})

	if err != nil {
		log.Errorf("error while destroying image: %s", err)
	} else {
		log.Debugf("Destroyed image %s", imageID)
	}

	return err
}

// GetVMName generates the VM name from peer information. It accepts a format
// function parameter to allow different formatting based on the desired use of
// the name.
func (vm *DockerVM) GetVMName(ccid ccintf.CCID, format func(string) (string, error)) (string, error) {
	name := ccid.GetName()

	if ccid.NetworkID != "" && ccid.PeerID != "" {
		name = fmt.Sprintf("%s-%s-%s", ccid.NetworkID, ccid.PeerID, name)
	} else if ccid.NetworkID != "" {
		name = fmt.Sprintf("%s-%s", ccid.NetworkID, name)
	} else if ccid.PeerID != "" {
		name = fmt.Sprintf("%s-%s", ccid.PeerID, name)
	}

	if format != nil {
		formattedName, err := format(name)
		if err != nil {
			return formattedName, err
		}
		name = formattedName
	}

	// replace any invalid characters with "-" (either in network id, peer id, or in the
	// entire name returned by any format function)
	name = vmRegExp.ReplaceAllString(name, "-")

	return name, nil
}
func (vm *DockerVM) GetContainerId(ccid ccintf.CCID) (string, error) {
	name := ccid.GetName()
	//if ccid.NetworkID != "" && ccid.PeerID != "" {
	//	name = fmt.Sprintf("%s-%s-%s", ccid.NetworkID, ccid.PeerID, name)
	//} else if ccid.NetworkID != "" {
	//	name = fmt.Sprintf("%s-%s", ccid.NetworkID, name)
	//} else if ccid.PeerID != "" {
	//	name = fmt.Sprintf("%s-%s", ccid.PeerID, name)
	//}
	name = contractcfg.GetConfig().ContractAddress + ":" + name
	// replace any invalid characters with "-" (either in network id, peer id, or in the
	// entire name returned by any format function)
	name = vmRegExp.ReplaceAllString(name, "-")

	return name, nil
}

func (vm *DockerVM) GetImageId(ccid ccintf.CCID) (string, error) {
	vmName := ccid.ChaincodeSpec.Type
	switch vmName {
	case 1:
		return "palletone/goimg", nil
	case 2:
		return "node", nil
	case 3:
		return "java", nil
	default:
		return "palletone", nil
	}
}

// formatImageName formats the docker image from peer information. This is
// needed to keep image (repository) names unique in a single host, multi-peer
// environment (such as a development environment). It computes the hash for the
// supplied image name and then appends it to the lowercase image name to ensure
// uniqueness.
func formatImageName(name string) (string, error) {
	hash := hex.EncodeToString(util.ComputeSHA256([]byte(name)))
	name = vmRegExp.ReplaceAllString(name, "-")
	imageName := strings.ToLower(fmt.Sprintf("%s-%s", name, hash))

	// Check that name complies with Docker's repository naming rules
	if !imageRegExp.MatchString(imageName) {
		log.Errorf("Error constructing Docker VM Name. '%s' breaks Docker's repository naming rules", name)
		return imageName, fmt.Errorf("Error constructing Docker VM Name. '%s' breaks Docker's repository naming rules", imageName)
	}

	return imageName, nil
}
