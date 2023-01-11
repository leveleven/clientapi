package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"

	"log"
	"net"
	"os"

	"strings"
)

type Net_Response struct {
	Data      NetworkInfo `json:"data"`
	ErrorCode int         `json:"error_code"`
	ErrorMsg  string      `json:"error_msg"`
}

type NetworkInfo struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	IP      string `json:"ip"`
	Netmask int    `json:"netmask"`
}

func (res *Net_Response) ErrorRes(p string, err string) {
	res.ErrorCode = 1
	msg := p + ": " + err
	if res.ErrorMsg == "" {
		res.ErrorMsg = msg
	} else {
		res.ErrorMsg = res.ErrorMsg + ";" + msg
	}
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

// func GetNetplanFile() (filename string) {
// 	fileInfos, err := ioutil.ReadDir("/etc/netplan/")
// 	if err != nil {
// 		log.Fatalln("Failed to open configuration file, ", err)
// 		return ""
// 	}

// 	for _, info := range fileInfos {
// 		if strings.Contains(info.Name(), ".yaml") {
// 			log.Println("Get network configure ", info.Name())
// 			return info.Name()
// 		}
// 	}
// 	return
// }

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

func NetworkConfig(address string, netmask string, gateway string, dns []string, need_reboot bool) bool {
	f, newfile_err := os.CreateTemp("/tmp/", "config")
	if newfile_err != nil {
		// log.Fatalln("Failed to write configuration file", newfile_err)
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
		// log.Fatalf("Failed to remove file %s to /etc/network/interfaces: %s", f.Name(), mv_err)
		return false
	}
	// log.Printf("move %s to /etc/network/interfaces", f.Name())

	log.Println(need_reboot)
	if need_reboot {
		cmd := "nmcli c reload | nmcli c up ifname eth0"
		exc := exec.Command("bash", "-c", cmd)
		err := exc.Run()
		if err != nil {
			// log.Fatalln(err)
			return false
		}
	}
	return true
}

func (n *Net_Response) GetNetworkCfg(netjs NetworkInfo) {
	n.Data = netjs
}

func GetNetwork() Net_Response {
	var nif Net_Response

	cmd := exec.Command("bash", "-c", "ip -j addr | jq '.[1] | {Name: .ifname, Address:.address, ip:.addr_info[0].local, netmask:.addr_info[0].prefixlen}'")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		nif.ErrorRes("net", stderr.String())
		log.Fatalln("Failed to run commed: ", err)
	}
	var t = stdout.String()

	var netjs NetworkInfo
	err = json.Unmarshal([]byte(t), &netjs)
	nif.GetNetworkCfg(netjs)
	if err != nil {
		nif.ErrorRes("net", stderr.String())
		log.Fatalln("Failed to encoding: ", err)
	}
	return nif
}
