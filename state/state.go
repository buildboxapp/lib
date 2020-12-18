package state

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"math"
)

type StateHost struct {
	PercentageCPU,
	PercentageMemory,
	PercentageDisk,
	TotalCPU,
	TotalMemory,
	TotalDisk,
	UsedCPU,
	UsedMemory,
	UsedDisk float64
	Info interface{}
}

func (c *StateHost) Tick()  {
	var pcpu, i float64

	memoryStat, _ 	:= mem.VirtualMemory()
	percentage, _ 	:= cpu.Percent(0, true)
	//phost, _ 		:= host.Info()
	diskStat, _ 	:= disk.Usage("/")

	for _, cpupercent := range percentage {
		pcpu = (pcpu + cpupercent)
		i ++
	}

	c.PercentageCPU 	= math.Round(pcpu / i)
	c.PercentageMemory 	= math.Round(memoryStat.UsedPercent)
	c.PercentageDisk 	= math.Round(diskStat.UsedPercent)
	//c.Info = phost

	return
}
