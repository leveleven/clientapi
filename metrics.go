package main

import (
	"bytes"
	"encoding/json"
	// "fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type Memory struct {
	Total   uint64
	Free    uint64
	Percent float64
}

type CPU struct {
	Percent float64
	Temp    int
}

type Disk struct {
	Status string
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
	Memory Memory
	CPU    CPU
	Disk   Disk
}

type Respond struct {
	Data Data
	ErrorCode int
	ErrorMsg string
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

func GetCPUInfo() CPU {
	percent, _ := cpu.Percent(time.Second, false)
	temp, _ := GetCPUTemp()
	return CPU{Percent: percent[0], Temp: temp}
}

func GetMemInfo() Memory {
	// byte
	m, _ := mem.VirtualMemory()
	return Memory{Total: m.Total, Free: m.Free, Percent: m.UsedPercent}
}

func GetDiskInfo() Disk {
	// cmd := exec.Command("lsblk", "-S", "-J", "-o", "NAME")
	cmd := exec.Command("lsblk", "-S", "-J")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
		return Disk{Status: "falded to get disk information"}
	}

	var t = out.String()
	if t == "" {
		return Disk{Status: "no_sata_disk"}
	} else {
		// encode json
		var e SCSI
		err = json.Unmarshal([]byte(t), &e)
		if err != nil {
			log.Fatalln("Failed to encoding: ", err)
		}
		// fmt.Println(e.Blockdevices[0].Name)
		
		return Disk{Status: GetDiskLog(e.Blockdevices[0].Name)}
		// return "", nil
	}
}

func GetDiskLog(device string) string {
	var path = "/dev/" + device
	cmd := exec.Command("smartctl", "-json", "-l", "error", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
		// return Disk{Status: "falded to get disk information"}
	}

	var s Smartctl
	err = json.Unmarshal([]byte(out.String()), &s)
	if err != nil {
		log.Fatalln("Failed to encoding: ", err)
	}

	var error_count = s.AtaSmartErrorLog.Summary.Conut
	if error_count == 0 {
		return "health"
	} else if error_count > 0 {
		return "unhealth"
	} else {
		return "error"
	}
}

func metrics() Respond {
	return Respond{Data: Data{Memory: GetMemInfo(), CPU: GetCPUInfo(), Disk: GetDiskInfo()}, ErrorCode: 0, ErrorMsg: ""}
}
