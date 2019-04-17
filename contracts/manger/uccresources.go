package manger

import (
	"github.com/fsouza/go-dockerclient"
	"fmt"
	"strings"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/list"
	"github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/vm/common"
)

type UccInterface interface {
	GetCPUUsageTotalUsage(cc *list.CCInfo) (uint64,error)
	GetMemoryStatsLimit(cc *list.CCInfo) (uint64,error)
	GetMemoryStatsUsage(cc *list.CCInfo) (uint64,error)
}

//type Resource struct {
//
//}

func getRT(cc *list.CCInfo){
	//r := &Resource{}
	cupusage,_ := GetCPUUsageTotalUsage(cc)
	log.Infof("================================================%d\n\n",cupusage)
	limit,_ := GetMemoryStatsLimit(cc)
	log.Infof("================================================%d\n\n",limit)
	usage,_ := GetMemoryStatsUsage(cc)
	log.Infof("================================================%d\n\n",usage)
}

func GetCPUUsageTotalUsage(cc *list.CCInfo)(uint64,error) {

stats,err := getResourceUses(cc)
if err != nil {
	return uint64(0),nil
}
return stats.CPUStats.CPUUsage.TotalUsage,nil
}
func GetMemoryStatsLimit(cc *list.CCInfo)(uint64,error) {
	stats,err := getResourceUses(cc)
	if err != nil {
		return uint64(0),nil
	}
	return stats.MemoryStats.Limit,nil
}
func  GetMemoryStatsUsage(cc *list.CCInfo)(uint64,error) {
	stats,err := getResourceUses(cc)
	if err != nil {
		return uint64(0),nil
	}
	return stats.MemoryStats.Usage,nil
}

func getResourceUses(cc *list.CCInfo)(*docker.Stats,error) {
	if !cc.SysCC {
		name := fmt.Sprintf("%s:%s:%s", cc.Name, cc.Version, contractcfg.GetConfig().ContractAddress)
		newName := strings.Replace(name, ":", "-", -1)
		client, err := util.NewDockerClient()
		if err != nil {
			log.Infof("-------------------1---------------------%s\n\n", err.Error())
			return nil,err
		}
		//info, err := client.Info()
		//if err != nil {
		//	log.Infof("----------------------2--------------%s\n\n", err.Error())
		//	return nil,err
		//}
		con, err := client.InspectContainer(newName)
		if err != nil {
			log.Infof("----------------------3--------------%s\n\n", err.Error())
			return nil,err
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
			log.Infof("----------------------------------------%s\n\n", err.Error())
			return nil,err
		}
		if len(resultStats) == 0 {
			return nil,fmt.Errorf("get container stats error")
		} else {
			stats := resultStats[0]
				return stats,nil
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
	return nil,fmt.Errorf("get container stats error")
}