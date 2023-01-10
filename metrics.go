package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Memory struct {
	Total   int     `json:"total"`
	Free    int     `json:"free"`
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
	Data      Data   `json:"data"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

func (res *Respones) ErrorRes(p string, err string) {
	res.ErrorCode = 1
	msg := p + ": " + err
	if res.ErrorMsg == "" {
		res.ErrorMsg = msg
	}
	res.ErrorMsg = res.ErrorMsg + ";" + msg
}

func GetCPUTemp() (int, string, error) {
	// 执行命令错误后会直接退出进程，需要优化
	cmd := exec.Command("cat", "/sys/devices/virtual1/thermal/thermal_zone0/temp")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
		fmt.Println(fmt.Sprint(err) + ":" + stderr.String())
		return -1, stderr.String(), err
	}
	tempStr := strings.Replace(out.String(), "\n", "", -1)
	temp, err := strconv.Atoi(tempStr)
	if err != nil {
		log.Fatalln("Failed to change string to int: ", err)
		return -1, "Failed to change string to int", err
	}
	temp = temp / 1000
	return temp, "", nil
}

func (r *Respones) GetCPUInfo() {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		r.Data.CPU.Percent = -1
		r.ErrorRes("cpu", err.Error())
	}
	r.Data.CPU.Percent = percent[0]

	temp, err_msg, err := GetCPUTemp()
	if err != nil {
		r.ErrorRes("cpu", err_msg)
	}
	r.Data.CPU.Temp = temp
}

func (r *Respones) GetMemInfo() {
	// byte
	vm, err := mem.VirtualMemory()
	if err != nil {
		r.Data.Memory.Free = -1
		r.Data.Memory.Total = -1
		r.Data.Memory.Percent = -1
		r.ErrorRes("mem", err.Error())
	}

	r.Data.Memory.Total = int(vm.Total)
	r.Data.Memory.Free = int(vm.Free)
	r.Data.Memory.Percent = vm.UsedPercent
}

func (r *Respones) GetDiskInfo() {
	// cmd := exec.Command("lsblk", "-S", "-J", "-o", "NAME")
	cmd := exec.Command("lsblk", "-S", "-J")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		r.Data.Disk.Status = "error"
		r.ErrorRes("disk", err.Error())
	}

	var t = out.String()
	if t == "" {
		r.Data.Disk.Status = "no_sata_disk"
	} else {
		// encode json
		var e SCSI
		err = json.Unmarshal([]byte(t), &e)
		if err != nil {
			r.Data.Disk.Status = "error"
			r.ErrorRes("disk", err.Error())
		}
		r.Data.Disk.Status, err = GetDiskLog(e.Blockdevices[0].Name)
		if err != nil {
			r.Data.Disk.Status = "error"
			r.ErrorRes("disk", err.Error())
		}
	}
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

func metrics() Respones {
	var res Respones

	res.GetMemInfo()
	res.GetCPUInfo()
	res.GetDiskInfo()

	return res
}
