package utils

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/crypto"
	"github.com/palletone/go-palletone/common/log"
	util2 "github.com/palletone/go-palletone/common/util"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	core2 "github.com/palletone/go-palletone/contracts/core"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/contracts/manger"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/vm/common"
	"strings"
	"time"
)

type PalletOneDocker struct {
	DockerClient *docker.Client
	dag          dag.IDag
	jury         core2.IAdapterJury
}

func NewPalletOneDocker(dag dag.IDag, jury core2.IAdapterJury) *PalletOneDocker {
	log.Debug("start NewPalletOneDocker")
	defer log.Debug("end NewPalletOneDocker")
	dockerClient, err := util.NewDockerClient()
	if err != nil {
		log.Debugf("NewDockerClient error %s", err.Error())
		return &PalletOneDocker{DockerClient: nil}
	}
	return &PalletOneDocker{DockerClient: dockerClient, dag: dag, jury: jury}
}

func (pD *PalletOneDocker) CreateDefaultUserContractNetWork() {
	log.Debug("start CreateDefaultUserContractNetWork")
	defer log.Debug("end CreateDefaultUserContractNetWork")
	network, err := pD.DockerClient.NetworkInfo(core.DefaultUccNetworkMode)
	if err != nil {
		//  不存在，需要创建
		log.Debugf("networkInfo error: %s", err.Error())
		network, err = pD.DockerClient.CreateNetwork(docker.CreateNetworkOptions{Name: core.DefaultUccNetworkMode, Driver: "bridge"})
		if err != nil {
			//  打印一些错误信息
			log.Debugf("createNetwork error: %s", err.Error())
			return
		}
		log.Debugf("network name = %s is created", network.Name)
		return
	}
	log.Debugf("network name = %s had already existed", network.Name)
}

func (pD *PalletOneDocker) PullUserContractImages(n chan struct{}) {
	defer close(n)
	log.Debug("start PullUserContractImages")
	defer log.Debug("end PullUserContractImages")
	goimgNameTag := contractcfg.Goimg + ":" + contractcfg.GptnVersion
	_, err := pD.DockerClient.InspectImage(goimgNameTag)
	if err != nil {
		log.Debugf("go image name = %s does not exist locally, attempt pull", goimgNameTag)
		err = pD.DockerClient.PullImage(docker.PullImageOptions{Repository: contractcfg.Goimg, Tag: contractcfg.GptnVersion}, docker.AuthConfiguration{})
		if err != nil {
			log.Debugf("Failed to pull %s, error: %s", goimgNameTag, err)
			return
		}
		log.Debugf("go image name = %s is pulled", goimgNameTag)
		return
	}
	log.Debugf("go image name = %s had already existed", goimgNameTag)
}

func (pD *PalletOneDocker) RestartUserContractsWhenStartGptn(n1 chan struct{}, n2 chan struct{}) {
	log.Debug("waiting pull image")
	<-n1
	log.Debug("start RestartUserContractsWhenStartGptn")
	defer log.Debug("end RestartUserContractsWhenStartGptn")
	defer close(n2)
	//  获取本地用户合约列表
	juryAddrs := pD.jury.GetLocalJuryAddrs()
	juryAddr := common.Address{}
	if len(juryAddrs) != 0 {
		juryAddr = juryAddrs[0]
	}
	contracts := pD.dag.GetContractsWithJuryAddr(juryAddr)
	//  启动退出的容器，包括本地有的和本地没有的
	if len(contracts) != 0 {
		log.Debugf("contracts length = %d", len(contracts))
		for _, c := range contracts {
			log.Debugf("contract id = %v", c.ContractId)
			rd, _ := crypto.GetRandomBytes(32)
			txid := util2.RlpHash(rd)
			//  启动 gptn 时启动Jury对应的没有过期的用户合约容器
			log.Debugf("create time %v", time.Unix(int64(c.CreationTime), 0).UTC())
			log.Debugf("now time %v", time.Now().UTC())
			log.Debugf("now during time = %d, contract during time %d", time.Now().Unix()-int64(c.CreationTime), c.DuringTime)
			duration := time.Now().Unix() - int64(c.CreationTime)
			if uint64(duration) < c.DuringTime {
				//expiredTime := time.Unix(time.Now().Unix()+int64(c.DuringTime), 0).UTC()
				//nowTime := time.Now().UTC()
				//log.Debugf("expiredTime = %s", expiredTime)
				//log.Debugf("nowTime = %s", nowTime)
				//isExpired := time.Now().UTC().After(expiredTime)
				//if !isExpired {
				//if c.Status == 1 {
				log.Debugf("restart container %s with jury address %s", c.Name, juryAddr.String())
				address := common.NewAddress(c.ContractId, common.ContractHash)
				_, _ = manger.RestartContainer(pD.dag, "palletone", address, txid.String())
			} else {
				log.Debugf("contract name = %s was expired", c.Name)
			}
		}
		return
	}
	log.Debugf("contracts length =%d", len(contracts))
}

func (pD *PalletOneDocker) GetAllContainers() ([]docker.APIContainers, error) {
	log.Debug("start GetAllContainers")
	defer log.Debug("end GetAllContainers")
	return pD.DockerClient.ListContainers(docker.ListContainersOptions{All: true, Size: true})
}

func (pD *PalletOneDocker) RestartExitedAndUnExpiredContainers(cons []docker.APIContainers) {
	log.Debug("start RestartExitedAndUnExpiredContainers")
	defer log.Debug("end RestartExitedAndUnExpiredContainers")
	//  获取所有退出容器
	containerNames, err := pD.GetAllContainersAddrsWithStatus(cons, "Exited")
	if err != nil {
		log.Debugf("client.GetAllExitedContainer err: %s", err.Error())
		return
	}
	log.Debugf("the exited containers len = %d", len(containerNames))
	//  判断是否是担任jury地址
	juryAddrs := pD.jury.GetLocalJuryAddrs()
	juryAddr := common.Address{}
	if len(juryAddrs) != 0 {
		juryAddr = juryAddrs[0]
	}
	contracts := pD.dag.GetContractsWithJuryAddr(juryAddr)
	if len(contracts) == 0 {
		log.Debugf("without any contact")
		return
	}
	log.Debugf("jury address = %s, the contracts len = %d", juryAddr.String(), len(contracts))
	for _, v := range containerNames {
		for _, c := range contracts {
			name := c.Name + ":" + c.Version + ":" + contractcfg.GetConfig().ContractAddress
			name = strings.Replace(name, ":", "-", -1)
			log.Debugf("name1 = %s,name2 = %s", v, name)
			if name == v {
				rd, _ := crypto.GetRandomBytes(32)
				txid := util2.RlpHash(rd)
				log.Debugf("container name = %s,user contract address = %s", v, c.Name)
				addr, _ := common.StringToAddress(c.Name)
				_, err = manger.RestartContainer(pD.dag, "palletone", addr, txid.String())
				if err != nil {
					log.Debugf("RestartContainer err: %s", err.Error())
					continue
				}
			}
		}
	}
}

//删除所有过期容器
func (pD *PalletOneDocker) RemoveExpiredContainers(cons []docker.APIContainers) {
	log.Debug("start RemoveExpiredContainers")
	defer log.Debug("end RemoveExpiredContainers")
	//  让用户合约容器一直在线，true and 0
	p := pD.dag.GetChainParameters()
	if p.RmExpConFromSysParam && p.UccDuringTime == 0 {
		log.Debug("keeping user contract containers online")
		return
	}
	//获取容器id，以及对应用户合约的地址，更新状态
	idStrMap := pD.RetrieveExpiredContainers(cons)
	if len(idStrMap) > 0 {
		for id, str := range idStrMap {
			err := pD.DockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: id, Force: true})
			if err != nil {
				log.Debugf("client.RemoveContainer id=%s error=%s", id, err.Error())
				continue
			}
			c, err := pD.dag.GetContract(str.Bytes())
			if err != nil {
				log.Debugf("get contract error %s", err.Error())
				continue
			}
			c.Status = 0
			err = pD.dag.SaveContract(c)
			if err != nil {
				log.Debugf("save contract error %s", err.Error())
			}
		}
	}
}

func (pD *PalletOneDocker) RetrieveExpiredContainers(containers []docker.APIContainers) map[string]common.Address {
	log.Debug("start RetrieveExpiredContainers")
	defer log.Debug("end RetrieveExpiredContainers")
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
				if pD.dag.GetChainParameters().RmExpConFromSysParam {
					log.Info("rm exp con from sys param............")
					containerDurTime = uint64(pD.dag.GetChainParameters().UccDuringTime)
				} else {
					log.Info("rm exp con from contact info..........")
					contract, err := pD.dag.GetContract(contractAddr.Bytes())
					if err != nil {
						log.Errorf("get contract error: %s", err.Error())
						continue
					}
					containerDurTime = contract.DuringTime
				}
				duration := time.Now().Unix() - c.Created
				log.Infof("")
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
//status: Up， Exited
func (pD *PalletOneDocker) GetAllContainersAddrsWithStatus(cons []docker.APIContainers, status string) ([]string, error) {
	log.Debug("start GetAllContainersAddrsWithStatus")
	defer log.Debug("end GetAllContainersAddrsWithStatus")
	if len(cons) != 0 {
		containerNames := make([]string, 0)
		for _, v := range cons {
			if strings.Contains(v.Names[0][1:3], "PC") && len(v.Names[0]) > 40 {
				if status != "" && !strings.Contains(v.Status, status) {
					continue
				}
				log.Debugf("find container name = %s", v.Names[0])
				containerNames = append(containerNames, v.Names[0][1:])
				//name := v.Names[0][1:36]
				//contractAddr, err := common.StringToAddress(name)
				//if err != nil {
				//	log.Debugf("common.StringToAddress err: %s", err.Error())
				//	continue
				//}
				//log.Debugf("find container name = %s", v.Names[0])
				//addr = append(addr, contractAddr)
			}
		}
		return containerNames, nil
	}
	return nil, fmt.Errorf("without any container")
}

//  调用的时候，若调用完发现磁盘使用超过系统上限，则kill掉并移除
func (pD *PalletOneDocker) RmConsOverDisk(cc *list.CCInfo) (sizeRW int64, disk int64, isOver bool) {
	log.Debug("start RmConsOverDisk")
	defer log.Debug("end RmConsOverDisk")
	cons, err := pD.GetAllContainers()
	if err != nil {
		log.Debugf("GetAllContainers error: %s", err.Error())
		return 0, 0, false
	}
	if len(cons) != 0 {
		//  获取name对应的容器
		name := cc.Name + ":" + cc.Version
		name = strings.Replace(name, ":", "-", -1)
		cp := pD.dag.GetChainParameters()
		for _, c := range cons {
			if c.Names[0][1:] == name && c.SizeRw > cp.UccDisk {
				err := pD.DockerClient.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID, Force: true})
				if err != nil {
					log.Debugf("client.RemoveContainer %s", err.Error())
					return 0, 0, false
				}
				log.Debugf("remove container %s", c.Names[0][1:36])
				return c.SizeRw, cp.UccDisk, true
			}
		}
	}
	log.Debug("without container")
	return 0, 0, false
}

//func GetResourcesWhenInvokeContainer(cc *list.CCInfo) {
//	log.Debugf("enter GetResourcesWhenInvokeContainer")
//	defer log.Debugf("exit GetResourcesWhenInvokeContainer")
//	if !cc.SysCC {
//		name := cc.Name + ":" + cc.Version
//		name = strings.Replace(name, ":", "-", -1)
//		stats, err := GetAllResourceUsageByContainerName(name)
//		if err != nil {
//			return
//		}
//		cupusage, _ := GetCPUUsageTotalUsage(stats)
//		log.Infof("================================================%d\n\n", cupusage)
//		limit, _ := GetMemoryStatsLimit(stats)
//		log.Infof("================================================%d\n\n", limit)
//		usage, _ := GetMemoryStatsUsage(stats)
//		log.Infof("================================================%d\n\n", usage)
//	}
//}
//
//func GetAllResourceUsageByContainerName(name string) (*docker.Stats, error) {
//	client, err := util.NewDockerClient()
//	if err != nil {
//		log.Error("util.NewDockerClient", "error", err)
//		return nil, err
//	}
//	//  通过容器名称获取容器id
//	con, err := client.InspectContainer(name)
//	if err != nil {
//		log.Infof("client.InspectContainer err: %s\n", err.Error())
//		return nil, err
//	}
//	errC := make(chan error, 1)
//	statsC := make(chan *docker.Stats)
//	done := make(chan bool)
//	defer close(done)
//	go func() {
//		errC <- client.Stats(docker.StatsOptions{ID: con.ID, Stats: statsC, Stream: false, Done: done,
//			InactivityTimeout: 3 * time.Second, Timeout: 3 * time.Second})
//		close(errC)
//	}()
//	var resultStats []*docker.Stats
//	for {
//		stats, ok := <-statsC
//		if !ok {
//			break
//		}
//		resultStats = append(resultStats, stats)
//	}
//	err = <-errC
//	if err != nil {
//		return nil, err
//	}
//	if len(resultStats) == 0 {
//		return nil, fmt.Errorf("get container stats error")
//	} else {
//		stats := resultStats[0]
//		return stats, nil
//	}
//	//return nil, fmt.Errorf("get container stats error")
//}
//func GetCPUUsageTotalUsage(stats *docker.Stats) (uint64, error) {
//	return stats.CPUStats.CPUUsage.TotalUsage, nil
//}
//func GetMemoryStatsLimit(stats *docker.Stats) (uint64, error) {
//	return stats.MemoryStats.Limit, nil
//}
//func GetMemoryStatsUsage(stats *docker.Stats) (uint64, error) {
//	return stats.MemoryStats.Usage, nil
//}

////  通过容器名称获取容器里面的错误信息，返回最后一条
//func GetLogFromContainer(name string) string {
//	client, err := util.NewDockerClient()
//	if err != nil {
//		log.Error("util.NewDockerClient", "error", err)
//		return ""
//	}
//	var buf bytes.Buffer
//	logsO := docker.LogsOptions{
//		Container:         name,
//		ErrorStream:       &buf,
//		Follow:            true,
//		Stderr:            true,
//		InactivityTimeout: 3 * time.Second,
//	}
//	log.Debugf("start docker logs")
//	err = client.Logs(logsO)
//	log.Debugf("end docker logs")
//	if err != nil {
//		log.Infof("get log from container %s error: %s", name, err.Error())
//		return ""
//	}
//	errArray := make([]string, 0)
//	for {
//		line, err := buf.ReadString('\n')
//		if err != nil {
//			if err == io.EOF {
//				break
//			}
//			return ""
//		}
//		line = strings.TrimSpace(line)
//		if strings.Contains(line, "panic: runtime error") || strings.Contains(line, "fatal error: runtime") {
//			log.Infof("container %s error %s", name, line)
//			errArray = append(errArray, line)
//		}
//	}
//	if len(errArray) != 0 {
//		return errArray[len(errArray)-1]
//	}
//	return ""
//}

//  获取所以用户合约使用的磁盘容量
//func GetDiskForEachContainer(client *docker.Client, disk int64) {
//	log.Debugf("Limit each container disk to %d", disk)
//	diskUsage, err := client.DiskUsage(docker.DiskUsageOptions{})
//	if err != nil {
//		log.Infof("client.DiskUsage err: %s\n", err.Error())
//		return
//	}
//	if diskUsage != nil {
//		for _, c := range diskUsage.Containers {
//			if strings.Contains(c.Names[0][1:3], "PC") {
//				//log.Infof("=======%#v\n", c)
//				log.Debugf("Current usage of container disk is %d", c.SizeRw)
//				if c.SizeRw > disk {
//					//  移除掉
//					err := client.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID, Force: true})
//					if err != nil {
//						log.Debugf("client.RemoveContainer error %s", err.Error())
//					}
//				}
//			}
//		}
//	}
//}

//获取所有容器
//func GetAllContainers(client *docker.Client) ([]docker.APIContainers, error) {
//	cons, err := client.ListContainers(docker.ListContainersOptions{All: true})
//	if err != nil {
//		log.Infof("client.ListContainers err: %s\n", err.Error())
//		return nil, err
//	}
//	return cons, nil
//}

//  获取所有过期的容器ID(通过交易上的)
//func RetrieveExpiredContainers(idag dag.IDag, containers []docker.APIContainers, rmExpConFromSysParam bool) map[string]common.Address {
//	log.Debugf("enter RetrieveExpiredContainers func")
//	idStr := make(map[string]common.Address)
//	if len(containers) > 0 {
//		for _, c := range containers {
//			if strings.Contains(c.Names[0][1:3], "PC") && len(c.Names[0]) > 40 {
//				contractName := c.Names[0][1:36]
//				contractAddr, err := common.StringToAddress(contractName)
//				if err != nil {
//					log.Errorf("string to address error: %s", err.Error())
//					continue
//				}
//				containerDurTime := uint64(0)
//				if rmExpConFromSysParam {
//					log.Info("rm exp con from sys param............")
//					containerDurTime = uint64(idag.GetChainParameters().UccDuringTime)
//				} else {
//					log.Info("rm exp con from contact info..........")
//					contract, err := idag.GetContract(contractAddr.Bytes())
//					if err != nil {
//						log.Errorf("get contract error: %s", err.Error())
//						continue
//					}
//					containerDurTime = contract.DuringTime
//				}
//				duration := time.Now().Unix() - c.Created
//				if uint64(duration) >= containerDurTime {
//					log.Infof("container name = %s was expired.", c.Names[0])
//					idStr[c.ID] = contractAddr
//				}
//			}
//		}
//	}
//	return idStr
//}

//  获取用户合约异常退出的监听函数
//status: Up， Exited
//func GetAllContainerAddr(cons []docker.APIContainers, status string) ([]common.Address, error) {
//	if len(cons) > 0 {
//		addr := make([]common.Address, 0)
//		for _, v := range cons {
//			if strings.Contains(v.Names[0][1:3], "PC") && len(v.Names[0]) > 40 {
//				if status != "" && !strings.Contains(v.Status, status) {
//					continue
//				}
//				name := v.Names[0][1:36]
//				contractAddr, err := common.StringToAddress(name)
//				if err != nil {
//					log.Infof("common.StringToAddress err: %s", err.Error())
//					continue
//				}
//				log.Infof("find container name = %s", v.Names[0])
//				addr = append(addr, contractAddr)
//			}
//		}
//		return addr, nil
//	}
//	return nil, fmt.Errorf("without any container")
//}

//  当调用合约时，发生超时，即停止掉容器
//func StopContainerWhenInvokeTimeOut(name string) {
//	log.Debugf("enter StopContainerWhenInvokeTimeOut name = %s", name)
//	defer log.Debugf("exit StopContainerWhenInvokeTimeOut name = %s", name)
//	client, err := util.NewDockerClient()
//	if err != nil {
//		log.Error("util.NewDockerClient", "error", err)
//		return
//	}
//	err = client.StopContainer(name, 3)
//	if err != nil {
//		log.Infof("stop container error: %s", err.Error())
//		return
//	}
//}

//  编译超时，移除容器
//func RemoveContainerWhenGoBuildTimeOut(id string) {
//	client, err := util.NewDockerClient()
//	if err != nil {
//		log.Error("util.NewDockerClient", "error", err)
//		return
//	}
//	<-time.After(contractcfg.GetConfig().ContractDeploytimeout)
//	err = client.RemoveContainer(docker.RemoveContainerOptions{ID: id, Force: true})
//	if err != nil {
//		log.Infof("remove container error: %s", err.Error())
//	}
//	//select {
//	//case <-time.After(contractcfg.GetConfig().ContractDeploytimeout):
//	//	err := client.RemoveContainer(docker.RemoveContainerOptions{ID: id, Force: true})
//	//	if err != nil {
//	//		log.Infof("remove container error: %s", err.Error())
//	//	}
//	//	return
//	//}
//}

//
////  调用的时候，若调用完发现磁盘使用超过系统上限，则kill掉并移除
//func RemoveConWhenOverDisk(cc *list.CCInfo, dag dag.IDag) (sizeRW int64, disk int64, isOver bool) {
//	log.Debugf("start KillAndRmWhenOver")
//	defer log.Debugf("end KillAndRmWhenOver")
//	client, err := util.NewDockerClient()
//	if err != nil {
//		log.Error("util.NewDockerClient", "error", err)
//		return 0, 0, false
//	}
//	//  获取所有容器
//	allCon, err := client.ListContainers(docker.ListContainersOptions{All: true, Size: true})
//	if err != nil {
//		log.Debugf("client.ListContainers %s", err.Error())
//		return 0, 0, false
//	}
//	if len(allCon) > 0 {
//		//  获取name对应的容器
//		name := cc.Name + ":" + cc.Version
//		name = strings.Replace(name, ":", "-", -1)
//		cp := dag.GetChainParameters()
//		for _, c := range allCon {
//			if c.Names[0][1:] == name && c.SizeRw > cp.UccDisk {
//				err := client.RemoveContainer(docker.RemoveContainerOptions{ID: c.ID, Force: true})
//				if err != nil {
//					log.Debugf("client.RemoveContainer %s", err.Error())
//					return 0, 0, false
//				}
//				log.Debugf("remove container %s", c.Names[0][1:36])
//				return c.SizeRw, cp.UccDisk, true
//			}
//		}
//	}
//	return 0, 0, false
//}

//判断容器是否正在运行
//func IsRunning(name string) bool {
//	client, err := util.NewDockerClient()
//	if err != nil {
//		log.Errorf(err.Error())
//		return false
//	}
//	c, err := client.InspectContainer(name)
//	if err != nil {
//		log.Errorf(err.Error())
//		return false
//	}
//	return c.State.Running
//}
//func CreateGptnNet(client *docker.Client) {
//	_, err := client.NetworkInfo(core.DefaultUccNetworkMode)
//	if err != nil {
//		log.Debugf("client.NetworkInfo error: %s", err.Error())
//		_, err := client.CreateNetwork(docker.CreateNetworkOptions{Name: core.DefaultUccNetworkMode, Driver: "bridge"})
//		if err != nil {
//			log.Debugf("client.CreateNetwork error: %s", err.Error())
//		}
//	}
//}
