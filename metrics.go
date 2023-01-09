package main

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Memory struct {
	Total   uint64  `json:"total"`
	Free    uint64  `json:"free"`
	Percent float64 `json:"percent"`
}

type CPU struct {
	Percent float64 `json:"percent"`
	Temp    int     `json:"temp"`
}

type Disk struct {
	Status string `json:"status"`
}

type SCSI struct {
	Blockdevices []BlockDevice
}

type BlockDevice struct {
	Name string
}

type Smartctl struct {
	AtaSmartErrorLog Log
}

type Log struct {
	Summary Summary
}

type Summary struct {
	Conut int
}

type Data struct {
	Memory Memory `json:"memory"`
	CPU    CPU    `json:"cpu"`
	Disk   Disk   `json:"disk"`
}

type Respones struct {
	Data      Data     `json:"data"`
	ErrorCode int      `json:"error_code"`
	ErrorMsg  []string `json:"error_msg"`
}

func GetCPUTemp() (int, error) {
	cmd := exec.Command("cat", "/sys/devices/virtual/thermal/thermal_zone0/temp")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
		return 0, err
	}
	tempStr := strings.Replace(out.String(), "\n", "", -1)
	temp, err := strconv.Atoi(tempStr)
	if err != nil {
		log.Fatalln("Failed to change string to int: ", err)
		return 0, err
	}
	temp = temp / 1000
	return temp, nil
}

func (r *Respones) GetCPUInfo() (string, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return "cpu", err
	}

	temp, err := GetCPUTemp()
	if err != nil {
		return "cpu", err
	}

	r.Data.CPU.Percent = percent[0]
	r.Data.CPU.Temp = temp

	return "cpu", nil
	// return CPU{Percent: percent[0], Temp: temp}
}

func (r *Respones) GetMemInfo() (string, error) {
	// byte
	vm, err := mem.VirtualMemory()
	if err != nil {
		return "mem", err
	}

	r.Data.Memory.Total = vm.Total
	r.Data.Memory.Free = vm.Free
	r.Data.Memory.Percent = vm.UsedPercent
	return "mem", nil
	// return Memory{Total: vm.Total, Free: vm.Free, Percent: vm.UsedPercent}
}

func (r *Respones) GetDiskInfo() (string, error) {
	// cmd := exec.Command("lsblk", "-S", "-J", "-o", "NAME")
	cmd := exec.Command("lsblk", "-S", "-J")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
		return "disk", err
		// return Disk{Status: "falded to get disk information"}
	}

	var t = out.String()
	if t == "" {
		r.Data.Disk.Status = "no_sata_disk"
		// return Disk{Status: "no_sata_disk"}
	} else {
		// encode json
		var e SCSI
		err = json.Unmarshal([]byte(t), &e)
		if err != nil {
			log.Fatalln("Failed to encoding: ", err)
			return "disk", err
		}
		// fmt.Println(e.Blockdevices[0].Name)
		r.Data.Disk.Status, err = GetDiskLog(e.Blockdevices[0].Name)
		if err != nil {
			return "disk", err
		}
		// return Disk{Status: GetDiskLog(e.Blockdevices[0].Name)}
		// return "", nil
	}
	return "disk", nil
}

func GetDiskLog(device string) (string, error) {
	var path = "/dev/" + device
	cmd := exec.Command("smartctl", "-json", "-l", "error", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
		return "", err
		// return Disk{Status: "falded to get disk information"}
	}

	var s Smartctl
	err = json.Unmarshal([]byte(out.Bytes()), &s)
	if err != nil {
		log.Fatalln("Failed to encoding: ", err)
		return "", err
	}

	error_count := s.AtaSmartErrorLog.Summary.Conut
	if error_count == 0 {
		return "health", nil
	} else if error_count > 0 {
		return "unhealth", nil
	} else {
		return "error", nil
	}
}

func (res *Respones) ErrorRes(p string, err error) {
	res.ErrorCode = 1
	msg := p + ": " + err.Error()

	res.ErrorMsg = append(res.ErrorMsg, msg)
}

func metrics() Respones {
	var res Respones
	m, err := res.GetMemInfo()
	if err != nil {
		res.ErrorRes(m, err)
	}
	c, err := res.GetCPUInfo()
	if err != nil {
		res.ErrorRes(c, err)
	}
	d, err := res.GetDiskInfo()
	if err != nil {
		res.ErrorRes(d, err)
	}

	return res
	// return Respones{Data: Data{Memory:getMemInfo(), CPU: GetCPUInfo(), Disk: GetDiskInfo()}, ErrorCode: 0}
}
