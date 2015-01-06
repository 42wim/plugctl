package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
)

const plugEnable = "GpioForCrond+1"
const plugDisable = "GpioForCrond+0"
const plugInfo = "GetInfo+"
const plugDisableAP = "ifconfig+ra0+down"

const plugURI = "/goform/SystemCommand?command="
const plugReadResult = "/adm/system_command.asp"

type plug struct {
	device      string
	credentials string
}

func (p *plug) enable() {
	fmt.Println("enabling plug.")
	url := "http://" + p.credentials + "@" + p.device + plugURI + plugEnable
	_, err := http.Get(url)
	if err != nil {
		log.Fatal("connection failed")
	}
}

func (p *plug) disable() {
	fmt.Println("disabling plug.")
	url := "http://" + p.credentials + "@" + p.device + plugURI + plugDisable
	_, err := http.Get(url)
	if err != nil {
		log.Fatal("connection failed")
	}
}

func (p *plug) info(info string) string {
	url := "http://" + p.credentials + "@" + p.device + plugURI + plugInfo + info
	_, err := http.Get(url)
	if err != nil {
		log.Fatal("connection failed!")
	}
	resp, err := http.Get("http://" + p.credentials + "@" + p.device + plugReadResult)
	if err != nil {
		log.Fatal("connection failed!")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	re := regexp.MustCompile("01(I|V|W|E)[0-9]+ 0*([0-9]+)")
	result := re.FindStringSubmatch(string(body))
	// if we don't have 2 matches something is wrong
	if len(result) > 2 {
		return (string(result[2]))
	} else {
		return ("error")
	}
}

func (p *plug) disableAP() {
	fmt.Println("disabling AP.")
	url := "http://" + p.credentials + "@" + p.device + plugURI + plugDisableAP
	_, err := http.Get(url)
	if err != nil {
		log.Fatal("connection failed")
	}
}

func main() {
	device := flag.String("ip", "192.168.8.74", "ipv4 address of smartplug device")
	credentials := flag.String("credentials", "admin:admin", "credentials specify as <login>:<pass>")
	do := flag.String("do", "", "enable/disable/info/disableAP")
	info := flag.String("info", "", "W/E/V/I\n\t\tW = centiWatt \n\t\tE = milliWatts/h\n\t\tV = milliVolts\n\t\tI = milliAmps")
	flag.Parse()
	if len(os.Args) == 1 {
		flag.PrintDefaults()
		return
	}
	p := plug{*device, *credentials}
	switch *do {
	case "enable":
		p.enable()
	case "disable":
		p.disable()
	case "disableAP":
		p.disableAP()
	default:
		fmt.Println(p.info(*info), *info)
	}
}
