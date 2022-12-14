package main

import (
	"bytes"
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

type Info struct {
	Memory Memory
	CPU    CPU
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

func metrics() Info {
	return Info{Memory: GetMemInfo(), CPU: GetCPUInfo()}
}
