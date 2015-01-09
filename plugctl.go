package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const plugEnable = "GpioForCrond+1"
const plugDisable = "GpioForCrond+0"
const plugInfo = "GetInfo+"
const plugDisableAP = "ifconfig+ra0+down"
const plugIfconfig = "ifconfig"
const plugUptime = "uptime"
const plugReboot = "reboot"

const plugURI = "/goform/SystemCommand?command="
const plugReadResult = "/adm/system_command.asp"

type plug struct {
	device      string
	credentials string
}

// helper functions

func checkErr(err error) {
	if err != nil {
		log.Fatalln("Error:", err)
	}
}

func sendln(conn net.Conn, s string, wait byte) string {
	_, err := fmt.Fprintf(conn, s+"\n")
	checkErr(err)
	status, err := bufio.NewReader(conn).ReadString(wait)
	checkErr(err)
	return status
}

func parseTextArea(body string) string {
	body = strings.Replace(body, "\n", "||", -1)
	re := regexp.MustCompile("1\">(.*)</textarea>")
	result := re.FindStringSubmatch(body)
	return result[1]
}

func parseResult(p *plug) string {
	resp, err := http.Get("http://" + p.credentials + "@" + p.device + plugReadResult)
	if err != nil {
		log.Fatal("connection failed!")
	}
	body, _ := ioutil.ReadAll(resp.Body)
	textarea := parseTextArea(string(body))
	return textarea
}

func printResultSuccess(result string) {
	if strings.Contains(result, "success") {
		fmt.Println("succesful")
	} else {
		fmt.Println("failed")
	}
}

// plug methods

func (p *plug) exec(command string) string {
	url := "http://" + p.credentials + "@" + p.device + plugURI
	url = url + command
	_, err := http.Get(url)
	if err != nil {
		log.Fatal("connection failed")
	}
	return parseResult(p)
}

func (p *plug) enable() {
	fmt.Print("enabling plug..")
	result := p.exec(plugEnable)
	printResultSuccess(result)
}

func (p *plug) disable() {
	fmt.Print("disabling plug..")
	result := p.exec(plugDisable)
	printResultSuccess(result)
}

func (p *plug) uptime() {
	result := p.exec(plugUptime)
	result = strings.Replace(result, "||", "", -1)
	fmt.Println(result)
}

func (p *plug) reboot() {
	fmt.Println("rebooting.")
	p.exec(plugReboot)
}

func (p *plug) info(info string) string {
	textarea := p.exec(plugInfo + info)
	re := regexp.MustCompile("01(I|V|W|E)[0-9]+ 0*([0-9]+)")
	result := re.FindStringSubmatch(textarea)
	// if we don't have 2 matches something is wrong
	if len(result) > 2 {
		return (string(result[2]))
	} else {
		return ("error")
	}
}

func (p *plug) disableAP() {
	fmt.Print("disabling AP...")
	p.exec(plugDisableAP)
	result := p.exec(plugIfconfig)
	if strings.Contains(result, "ra0") {
		fmt.Println("failed")
	} else {
		fmt.Println("success")
	}
}

func (p *plug) raw(command string) {
	fmt.Println("executing command:", command)
	command = strings.Replace(command, " ", "+", -1)
	result := p.exec(command)
	result = strings.Replace(result, "||", "\n", -1)
	fmt.Println(result)
}

// use telnet connection
func (p *plug) rawt(command string) {
	conn, err := net.Dial("tcp", p.device+":23")
	s := strings.Split(p.credentials, ":")
	login, pass := s[0], s[1]
	if err != nil {
		log.Fatal("can't connect")
	}
	sendln(conn, "", '\n')
	sendln(conn, login, '\n')
	sendln(conn, pass, '#')
	status := sendln(conn, command, '#')
	status = strings.Replace(status, command+"\r\n", "", 1)
	status = strings.Replace(status, "#", "", 1)
	fmt.Print(status)
}

func main() {
	device := flag.String("ip", "192.168.8.74", "ipv4 address of smartplug device")
	credentials := flag.String("credentials", "admin:admin", "credentials specify as <login>:<pass>")
	do := flag.String("do", "info", "enable/disable/info/disableAP/uptime/reboot")
	raw := flag.String("raw", "", "raw command to execute (via http)")
	rawt := flag.String("rawt", "", "raw command to execute (via telnet)")
	info := flag.String("info", "W", "W/E/V/I\n\t\tW = centiWatt \n\t\tE = milliWatts/h\n\t\tV = milliVolts\n\t\tI = milliAmps")
	flag.Parse()
	if len(os.Args) == 1 {
		flag.PrintDefaults()
		return
	}
	p := plug{*device, *credentials}

	if *raw != "" {
		p.raw(*raw)
		return
	}

	if *rawt != "" {
		p.rawt(*rawt)
		return
	}

	switch *do {
	case "enable":
		p.enable()
	case "disable":
		p.disable()
	case "disableAP":
		p.disableAP()
	case "uptime":
		p.uptime()
	case "reboot":
		p.reboot()
	case "info":
		fmt.Println(p.info(*info), *info)
	default:
		flag.PrintDefaults()
	}
}
