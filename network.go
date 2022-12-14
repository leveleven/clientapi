package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"

	// "fmt"
	"io/ioutil"
	"log"
	"net"
	"os"

	// "strconv"
	"strings"
)

type NetworkInfo struct {
	Name    string
	Address string
	IP      string
	Netmask int
}

func FlagMatch(flags []string, tags []string) (match bool) {
	match = false
	index := 0
	for _, tag := range tags {
		for i := 0; i < len(flags); i++ {
			if flags[i] == tag {
				index++
				break
			}
		}
		if index == 2 {
			match = true
			return
		}
	}
	return
}

func GetInterfaseName() (name string) {
	tags := [2]string{"up", "broadcast"}
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Fatalln("Failed to get broadcast,", err)
	}
	for _, iface := range ifaces {
		flags := strings.Split(iface.Flags.String(), "|")

		if FlagMatch(flags, tags[:]) {
			log.Println("Get interface", iface.Name)
			return iface.Name
		}
	}
	return
}

func GetNetplanFile() (filename string) {
	fileInfos, err := ioutil.ReadDir("/etc/netplan/")
	if err != nil {
		log.Fatalln("Failed to open configuration file, ", err)
		return ""
	}

	for _, info := range fileInfos {
		if strings.Contains(info.Name(), ".yaml") {
			log.Println("Get network configure ", info.Name())
			return info.Name()
		}
	}
	return
}

// 掩码转网络长度
// func SubNetMaskToLen(netmask string) (string, error) {
// 	ipSplitArr := strings.Split(netmask, ".")
// 	if len(ipSplitArr) != 4 {
// 		return "", fmt.Errorf("netmask:%s is not valid, pattern should like: 255.255.255.0", netmask)
// 	}
// 	ipv4MaskArr := make([]byte, 4)
// 	for i, value := range ipSplitArr {
// 		intValue, err := strconv.Atoi(value)
// 		if err != nil {
// 			log.Fatalf("ipMaskToInt call strconv.Atoi error:[%v] string value is: [%s]", err, value)
// 		}
// 		if intValue > 255 {
// 			log.Fatalf("netmask cannot greater than 255, current value is: [%s]", value)
// 		}
// 		ipv4MaskArr[i] = byte(intValue)
// 	}
// 	ones, _ := net.IPv4Mask(ipv4MaskArr[0], ipv4MaskArr[1], ipv4MaskArr[2], ipv4MaskArr[3]).Size()
// 	return strconv.Itoa(ones), nil
// }

func NetworkConfig(address string, netmask string, gateway string, dns []string) bool {
	f, newfile_err := os.CreateTemp("/tmp/", "config")
	if newfile_err != nil {
		log.Fatalln("Failed to write configuration file", newfile_err)
		return false
	}
	defer os.Remove(f.Name())
	intf := GetInterfaseName()

	// 开始编辑文件
	write := bufio.NewWriter(f)

	write.WriteString("auto " + intf + "\n")
	write.WriteString("iface " + intf + " inet static\n")
	write.WriteString("\taddress " + address + "\n")
	write.WriteString("\tgateway " + gateway + "\n")
	write.WriteString("\tnetmask " + netmask + "\n")
	write.WriteString("\tdns-nameservers ")
	for _, tag := range dns {
		write.WriteString(tag + " ")
	}
	write.Flush()

	// filename := GetNetplanFile()
	os.Chmod(f.Name(), 0644)
	mv_err := os.Rename(f.Name(), "/etc/network/interfaces")
	if mv_err != nil {
		log.Fatalf("Failed to remove file %s to /etc/network/interfaces: %s", f.Name(), mv_err)
		return false
	}
	log.Printf("move %s to /etc/network/interfaces", f.Name())
	return true
}

func GetNetwork() NetworkInfo {
	cmd := exec.Command("bash", "-c", "ip -j addr | jq '.[2] | {Name: .ifname, Address:.address, ip:.addr_info[0].local, netmask:.addr_info[0].prefixlen}'")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		log.Fatalln("Failed to run commed: ", err)
	}

	var netjs NetworkInfo
	err = json.Unmarshal([]byte(out.String()), &netjs)
	if err != nil {
		log.Fatalln("Failed to decoding: ", err)
	}
	return netjs
}
