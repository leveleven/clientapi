package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
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
	TRAN string
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

type MachineInfo struct {
	Data struct {
		Serial string `json:"serial"`
	} `json:"data"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
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
	} else {
		res.ErrorMsg = res.ErrorMsg + ";" + msg
	}
}

func GetCPUTemp() (int, string, error) {
	cmd := exec.Command("cat", "/sys/devices/virtual/thermal/thermal_zone0/temp")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		return -1, stderr.String(), err
	}
	tempStr := strings.Replace(stdout.String(), "\n", "", -1)
	temp, err := strconv.Atoi(tempStr)
	if err != nil {
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
	// cmd := exec.Command("lsblk", "-S", "-J", "-o", "NAME,TRAN,FSTYPE")
	cmd := exec.Command("lsblk", "-S", "-J")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		r.Data.Disk.Status = "error"
		r.ErrorRes("disk", err.Error())
	}

	var t = stdout.String()
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
		// 硬盘优化点
		for _, device := range e.Blockdevices {
			if device.TRAN == "sata" {
				disk_status, err_msg, err := GetDiskLog(device.Name)
				if err != nil {
					r.Data.Disk.Status = "error"
					r.ErrorRes("disk", err_msg)
				} else {
					r.Data.Disk.Status = disk_status
				}
				break
			}
		}
	}
}

func GetDiskLog(device string) (string, string, error) {
	var path = "/dev/" + device
	cmd := exec.Command("smartctl", "-json", "-l", "error", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", stderr.String(), err
	}

	var s Smartctl
	err = json.Unmarshal([]byte(stdout.Bytes()), &s)
	if err != nil {
		return "", stderr.String(), err
	}

	error_count := s.AtaSmartErrorLog.Summary.Conut
	if error_count == 0 {
		return "health", "", nil
	} else if error_count > 0 {
		return "unhealth", "", nil
	} else {
		return "error", "", nil
	}
}

func getns() MachineInfo {
	sn := exec.Command("bash", "-c", "lshw -disable pci -disable usb -c system -quiet -json | jq .[0]")
	var stdout, stderr bytes.Buffer
	sn.Stdout = &stdout
	sn.Stderr = &stderr
	err := sn.Run()

	var n MachineInfo
	if err != nil {
		n.Data = struct {
			Serial string "json:\"serial\""
		}{}
		n.ErrorCode = 1
		n.ErrorMsg = stderr.String()
		return n
	}

	err = json.Unmarshal([]byte(stdout.Bytes()), &n.Data)
	if err != nil {
		n.Data = struct {
			Serial string "json:\"serial\""
		}{}
		n.ErrorCode = 1
		n.ErrorMsg = err.Error()
		return n
	}

	env_err := godotenv.Load("/root/.env")
	if env_err != nil {
		n.Data = struct {
			Serial string "json:\"serial\""
		}{}
		n.ErrorCode = 1
		n.ErrorMsg = err.Error()
	}
	env_sn := os.Getenv("SN")
	if env_sn != "" {
		n.Data.Serial = env_sn
	}

	return n
}

func metrics() Respones {
	var res Respones

	res.GetMemInfo()
	res.GetCPUInfo()
	res.GetDiskInfo()

	return res
}
