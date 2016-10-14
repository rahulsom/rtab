package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func GetNetstatOutput() string {
	netstat := exec.Command("netstat", "-rnf", "inet")
	var out bytes.Buffer
	netstat.Stdout = &out
	err := netstat.Run()
	if err != nil {
		log.Fatal(err)
	}
	return out.String()
}

type NetstatLine struct {
	destination string
	gateway     string
	flags       string
	refs        int
	use         int
	netif       string
	expire      int
	seq         int
}
type NetstatLines []*NetstatLine

func (m *NetstatLine) Display() string {
	return fmt.Sprintf("%15s %8s %3d", m.destination, m.flags, m.seq)
}

func (m *NetstatLine) Route() string {
	return fmt.Sprintf("%15s (%s)", m.gateway, m.netif)
}

func parse(line string, seq int) *NetstatLine {
	parts := strings.Fields(line)
	var expire int = -1
	if len(parts) > 6 {
		expire1, err := strconv.Atoi(parts[6])
		if err == nil {
			expire = expire1
		}
	}

	refs, err := strconv.Atoi(parts[3])
	if err != nil {
		refs = -1
	}

	use, err := strconv.Atoi(parts[4])
	if err != nil {
		use = -1
	}

	retval := &NetstatLine{
		destination: parts[0],
		gateway:     parts[1],
		flags:       parts[2],
		refs:        refs,
		use:         use,
		netif:       parts[5],
		expire:      expire,
		seq:         seq,
	}
	return retval
}

func Lines(data string) []*NetstatLine {
	lines := strings.Split(data, "\n")
	startRecording := false
	retval := []*NetstatLine{}
	for i := 0; i < len(lines); i++ {
		if startRecording && len(lines[i]) > 0 {
			retval = append(retval, parse(lines[i], i))
		}
		if strings.Contains(lines[i], "Destination") {
			startRecording = true
		}
	}
	return retval
}

func printGateway(gateway string) {
	fmt.Println("                                 -->  ", gateway)
	fmt.Println("--------------------------------------------------------------")
	fmt.Println("")
	fmt.Println("")

}
func main() {
	netstat := GetNetstatOutput()
	lines := Lines(netstat)

	gateways := make(map[string]NetstatLines)

	for _, val := range lines {
		route := val.Route()
		if gateways[route] == nil {
			gateways[route] = make(NetstatLines, 0)
		}
		gateways[route] = append(gateways[route], val)
	}

	fmt.Println("")
	fmt.Println("")

	for k, v := range gateways {
		if !strings.Contains(k, ":") && !strings.Contains(k, "#") {
			didPrint := 0
			for _, v1 := range v {
				if v1.destination != "default" {
					fmt.Println("", v1.Display())
					didPrint = didPrint + 1
				}
			}
			if didPrint > 0 {
				printGateway(k)
			}
		}
	}

	for k, v := range gateways {
		if !strings.Contains(k, ":") && !strings.Contains(k, "#") {
			didPrint := 0
			for _, v1 := range v {
				if v1.destination == "default" {
					fmt.Println("", v1.Display())
					didPrint = didPrint + 1
				}
			}
			if didPrint > 0 {
				printGateway(k)
			}
		}
	}
}
