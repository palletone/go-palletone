package utils

import (
	"bytes"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/vm/common"
	"io"
	"strings"
	"time"
)

type UccInterface interface {
	GetCPUUsageTotalUsage(cc *list.CCInfo) (uint64, error)
	GetMemoryStatsLimit(cc *list.CCInfo) (uint64, error)
	GetMemoryStatsUsage(cc *list.CCInfo) (uint64, error)
}

//type Resource struct {
//
//}

func getRT(cc *list.CCInfo) {
	//r := &Resource{}
	cupusage, _ := GetCPUUsageTotalUsage(cc)
	log.Infof("================================================%d\n\n", cupusage)
	limit, _ := GetMemoryStatsLimit(cc)
	log.Infof("================================================%d\n\n", limit)
	usage, _ := GetMemoryStatsUsage(cc)
	log.Infof("================================================%d\n\n", usage)
}

func GetCPUUsageTotalUsage(cc *list.CCInfo) (uint64, error) {
	stats, err := getResourceUses(cc)
	if err != nil {
		return uint64(0), nil
	}
	return stats.CPUStats.CPUUsage.TotalUsage, nil
}
func GetMemoryStatsLimit(cc *list.CCInfo) (uint64, error) {
	stats, err := getResourceUses(cc)
	if err != nil {
		return uint64(0), nil
	}
	return stats.MemoryStats.Limit, nil
}
func GetMemoryStatsUsage(cc *list.CCInfo) (uint64, error) {
	stats, err := getResourceUses(cc)
	if err != nil {
		return uint64(0), nil
	}
	return stats.MemoryStats.Usage, nil
}

func getResourceUses(cc *list.CCInfo) (*docker.Stats, error) {
	if !cc.SysCC {
		name := fmt.Sprintf("%s:%s:%s", cc.Name, cc.Version, contractcfg.GetConfig().ContractAddress)
		newName := strings.Replace(name, ":", "-", -1)
		client, err := util.NewDockerClient()
		if err != nil {
			log.Infof("util.NewDockerClient err: %s\n", err.Error())
			return nil, err
		}
		//info, err := client.Info()
		//if err != nil {
		//	log.Infof("----------------------2--------------%s\n\n", err.Error())
		//	return nil,err
		//}
		con, err := client.InspectContainer(newName)
		if err != nil {
			log.Infof("client.InspectContainer err: %s\n", err.Error())
			return nil, err
		}
		errC := make(chan error, 1)
		statsC := make(chan *docker.Stats)
		done := make(chan bool)
		defer close(done)
		go func() {
			errC <- client.Stats(docker.StatsOptions{ID: con.ID, Stats: statsC, Stream: false, Done: done})
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
			//log.Infof("----------------------------------------%#v\n\n", stats)
			//log.Infof("----------------------------------------%#v\n\n", stats.Read)
			//log.Infof("-------------------stats.CPUStats.CPUUsage.PercpuUsage---------------------%d\n\n", stats.CPUStats.CPUUsage.PercpuUsage)
			//log.Infof("-------------------stats.CPUStats.CPUUsage.TotalUsage---------------------%d\n\n", stats.CPUStats.CPUUsage.TotalUsage)
			//log.Infof("-------------------stats.CPUStats.CPUUsage.UsageInKernelmode---------------------%d\n\n", stats.CPUStats.CPUUsage.UsageInKernelmode)
			//log.Infof("-------------------stats.CPUStats.CPUUsage.UsageInUsermode---------------------%d\n\n", stats.CPUStats.CPUUsage.UsageInUsermode)
			//log.Infof("-------------------stats.CPUStats.CPUUsage.UsageInUsermode---------------------%d\n\n", stats.CPUStats.SystemCPUUsage)
			//log.Infof("----------------------stats.MemoryStats.Stats.Swap------------------%d\n\n", stats.MemoryStats.Stats.Swap)
			//log.Infof("----------------------stats.MemoryStats.Limit------------------%d\n\n", stats.MemoryStats.Limit)
			//log.Infof("----------------------stats.MemoryStats.MaxUsage------------------%d\n\n", stats.MemoryStats.MaxUsage)
			//log.Infof("----------------------stats.MemoryStats.Usage------------------%d\n\n", stats.MemoryStats.Usage)
		}
	}
	return nil, fmt.Errorf("get container stats error")
}

//  通过容器名称获取容器里面的错误信息，返回最后一条
func GetLogFromContainer(name string) string {
	client, err := util.NewDockerClient()
	if err != nil {
		log.Info("util.NewDockerClient", "error", err)
		return ""
	}
	var buf bytes.Buffer
	logsO := docker.LogsOptions{
		Container:         name,
		ErrorStream:       &buf,
		Follow:            true,
		Stderr:            true,
		InactivityTimeout: time.Duration(3 * time.Second),
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

func GetAllExitedContainer(client *docker.Client) ([]common.Address, error) {
	cons, err := client.ListContainers(docker.ListContainersOptions{All: true})
	if err != nil {
		log.Infof("client.ListContainers err: %s\n", err.Error())
		return nil, err
	}
	addr := make([]common.Address, 0)
	if len(cons) > 0 {
		for _, v := range cons {
			if strings.Contains(v.Names[0][1:3], "PC") && strings.Contains(v.Status, "Exited") {
				name := v.Names[0][1:36]
				contractAddr, err := common.StringToAddress(name)
				if err != nil {
					log.Infof("common.StringToAddress err: %s", err.Error())
					continue
				}
				addr = append(addr, contractAddr)
			}
		}
		return addr, nil
	}
	return nil, fmt.Errorf("without any container")
}
