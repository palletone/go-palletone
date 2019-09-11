package utils

import (
	"bytes"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/vm/common"
	"io"
	"strings"
	"time"
)

type UccInterface interface {
	//  获取容器使用全部资源
	GetResourcesWhenInvokeContainer(cc *list.CCInfo)
	GetAllResourceUsageByContainerName(name string) (*docker.Stats, error)
	//  获取CPU使用
	GetCPUUsageTotalUsage(stats *docker.Stats) (uint64, error)
	//  获取内存使用上限
	GetMemoryStatsLimit(stats *docker.Stats) (uint64, error)
	//  获取当前内存使用
	GetMemoryStatsUsage(stats *docker.Stats) (uint64, error)
	//  通过容器名称获取容器里面的错误信息，返回最后一条
	GetLogFromContainer(name string) string
	//  获取所以用户合约使用的磁盘容量
	GetDiskForEachContainer(client *docker.Client, disk int64)
	//  获取用户合约异常退出的监听函数
	GetAllExitedContainer(client *docker.Client) ([]common.Address, error)
	//  当调用合约时，发生超时，即停止掉容器
	StopContainerWhenInvokeTimeOut(name string)
	//  编译超时，移除容器
	RemoveContainerWhenGoBuildTimeOut(client *docker.Client, id string)
}

func GetResourcesWhenInvokeContainer(cc *list.CCInfo) {
	log.Debugf("enter GetResourcesWhenInvokeContainer")
	defer log.Debugf("exit GetResourcesWhenInvokeContainer")
	if !cc.SysCC {
		name := cc.Name + ":" + cc.Version
		name = strings.Replace(name, ":", "-", -1)
		stats, err := GetAllResourceUsageByContainerName(name)
		if err != nil {
			return
		}
		cupusage, _ := GetCPUUsageTotalUsage(stats)
		log.Infof("================================================%d\n\n", cupusage)
		limit, _ := GetMemoryStatsLimit(stats)
		log.Infof("================================================%d\n\n", limit)
		usage, _ := GetMemoryStatsUsage(stats)
		log.Infof("================================================%d\n\n", usage)
	}
}

func GetAllResourceUsageByContainerName(name string) (*docker.Stats, error) {
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return nil, err
	}
	//  通过容器名称获取容器id
	con, err := client.InspectContainer(name)
	if err != nil {
		log.Infof("client.InspectContainer err: %s\n", err.Error())
		return nil, err
	}
	errC := make(chan error, 1)
	statsC := make(chan *docker.Stats)
	done := make(chan bool)
	defer close(done)
	go func() {
		errC <- client.Stats(docker.StatsOptions{ID: con.ID, Stats: statsC, Stream: false, Done: done,
			InactivityTimeout: 3 * time.Second, Timeout: 3 * time.Second})
		close(errC)
	}()
	var resultStats []*docker.Stats
	for {
		stats, ok := <-statsC
		if !ok {
			break
		}
		resultStats = append(resultStats, stats)
	}
	err = <-errC
	if err != nil {
		return nil, err
	}
	if len(resultStats) == 0 {
		return nil, fmt.Errorf("get container stats error")
	} else {
		stats := resultStats[0]
		return stats, nil
	}
	//return nil, fmt.Errorf("get container stats error")
}
func GetCPUUsageTotalUsage(stats *docker.Stats) (uint64, error) {
	return stats.CPUStats.CPUUsage.TotalUsage, nil
}
func GetMemoryStatsLimit(stats *docker.Stats) (uint64, error) {
	return stats.MemoryStats.Limit, nil
}
func GetMemoryStatsUsage(stats *docker.Stats) (uint64, error) {
	return stats.MemoryStats.Usage, nil
}

//  通过容器名称获取容器里面的错误信息，返回最后一条
func GetLogFromContainer(name string) string {
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return ""
	}
	var buf bytes.Buffer
	logsO := docker.LogsOptions{
		Container:         name,
		ErrorStream:       &buf,
		Follow:            true,
		Stderr:            true,
		InactivityTimeout: 3 * time.Second,
	}
	log.Debugf("start docker logs")
	err = client.Logs(logsO)
	log.Debugf("end docker logs")
	if err != nil {
		log.Infof("get log from container %s error: %s", name, err.Error())
		return ""
	}
	errArray := make([]string, 0)
	for {
		line, err := buf.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return ""
		}
		line = strings.TrimSpace(line)
		if strings.Contains(line, "panic: runtime error") || strings.Contains(line, "fatal error: runtime") {
			log.Infof("container %s error %s", name, line)
			errArray = append(errArray, line)
		}
	}
	if len(errArray) != 0 {
		return errArray[len(errArray)-1]
	}
	return ""
}

//  获取所以用户合约使用的磁盘容量
func GetDiskForEachContainer(client *docker.Client, disk int64) {
	log.Debugf("Limit each container disk to %d", disk)
	diskUsage, err := client.DiskUsage(docker.DiskUsageOptions{})
	if err != nil {
		log.Infof("client.DiskUsage err: %s\n", err.Error())
		return
	}
	if diskUsage != nil {
		for _, c := range diskUsage.Containers {
			if strings.Contains(c.Names[0][1:3], "PC") {
				//log.Infof("=======%#v\n", c)
				log.Debugf("Current usage of container disk is %d", c.SizeRw)
				if c.SizeRw > disk {
					//  移除掉
					err := client.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID, Force: true})
					if err != nil {
						log.Debugf("client.RemoveContainer error %s", err.Error())
					}
				}
			}
		}
	}
}

//获取所有容器
func GetAllContainers(client *docker.Client) ([]docker.APIContainers, error) {
	cons, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Infof("client.ListContainers err: %s\n", err.Error())
		return nil, err
	}
	return cons, nil
}

//  获取所有过期的容器ID(通过交易上的)
func RetrieveExpiredContainers(idag dag.IDag, containers []docker.APIContainers, rmExpConFromSysParam bool) map[string]common.Address {
	log.Debugf("enter RetrieveExpiredContainers func")
	idStr := make(map[string]common.Address)
	if len(containers) > 0 {
		for _, c := range containers {
			if strings.Contains(c.Names[0][1:3], "PC") && len(c.Names[0]) > 40 {
				contractName := c.Names[0][1:36]
				contractAddr, err := common.StringToAddress(contractName)
				if err != nil {
					log.Errorf("string to address error: %s", err.Error())
					continue
				}
				containerDurTime := uint64(0)
				if rmExpConFromSysParam {
					log.Info("rm exp con from sys param............")
					containerDurTime = uint64(idag.GetChainParameters().UccDuringTime)
				} else {
					log.Info("rm exp con from contact info..........")
					contract, err := idag.GetContract(contractAddr.Bytes())
					if err != nil {
						log.Errorf("get contract error: %s", err.Error())
						continue
					}
					containerDurTime = contract.DuringTime
				}
				duration := time.Now().Unix() - c.Created
				if uint64(duration) >= containerDurTime {
					log.Infof("container name = %s was expired.", c.Names[0])
					idStr[c.ID] = contractAddr
				}
			}
		}
	}
	return idStr
}

//  获取用户合约异常退出的监听函数
func GetAllExitedContainer(cons []docker.APIContainers) ([]common.Address, error) {
	if len(cons) > 0 {
		addr := make([]common.Address, 0)
		for _, v := range cons {
			if strings.Contains(v.Names[0][1:3], "PC") && strings.Contains(v.Status, "Exited") && len(v.Names[0]) > 40 {
				name := v.Names[0][1:36]
				contractAddr, err := common.StringToAddress(name)
				if err != nil {
					log.Infof("common.StringToAddress err: %s", err.Error())
					continue
				}
				log.Infof("container name = %s was exited.", v.Names[0])
				addr = append(addr, contractAddr)
			}
		}
		return addr, nil
	}
	return nil, fmt.Errorf("without any container")
}

//  当调用合约时，发生超时，即停止掉容器
func StopContainerWhenInvokeTimeOut(name string) {
	log.Debugf("enter StopContainerWhenInvokeTimeOut name = %s", name)
	defer log.Debugf("exit StopContainerWhenInvokeTimeOut name = %s", name)
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return
	}
	err = client.StopContainer(name, 3)
	if err != nil {
		log.Infof("stop container error: %s", err.Error())
		return
	}
}

//  编译超时，移除容器
func RemoveContainerWhenGoBuildTimeOut(id string) {
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return
	}
	<-time.After(contractcfg.GetConfig().ContractDeploytimeout)
	err = client.RemoveContainer(docker.RemoveContainerOptions{ID: id, Force: true})
	if err != nil {
		log.Infof("remove container error: %s", err.Error())
	}
	//select {
	//case <-time.After(contractcfg.GetConfig().ContractDeploytimeout):
	//	err := client.RemoveContainer(docker.RemoveContainerOptions{ID: id, Force: true})
	//	if err != nil {
	//		log.Infof("remove container error: %s", err.Error())
	//	}
	//	return
	//}
}

//  调用的时候，若调用完发现磁盘使用超过系统上限，则kill掉并移除
func RemoveConWhenOverDisk(cc *list.CCInfo, dag dag.IDag) (sizeRW int64, disk int64, isOver bool) {
	log.Debugf("enter KillAndRmWhenOver")
	defer log.Debugf("exit KillAndRmWhenOver")
	client, err := util.NewDockerClient()
	if err != nil {
		log.Error("util.NewDockerClient", "error", err)
		return 0, 0, false
	}
	//  获取所有容器
	allCon, err := client.ListContainers(docker.ListContainersOptions{All: true, Size: true})
	if err != nil {
		log.Debugf("client.ListContainers %s", err.Error())
		return 0, 0, false
	}
	if len(allCon) > 0 {
		//  获取name对应的容器
		name := cc.Name + ":" + cc.Version
		name = strings.Replace(name, ":", "-", -1)
		cp := dag.GetChainParameters()
		for _, c := range allCon {
			if c.Names[0][1:] == name && c.SizeRw > cp.UccDisk {
				err := client.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID, Force: true})
				if err != nil {
					log.Debugf("client.RemoveContainer %s", err.Error())
					return 0, 0, false
				}
				log.Debugf("remove container %s", c.Names[0][1:36])
				return c.SizeRw, cp.UccDisk, true
			}
		}
	}
	return 0, 0, false
}
