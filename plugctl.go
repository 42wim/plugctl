package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

var debug bool = false

func main() {
	device := flag.String("ip", "192.168.8.74", "ipv4 address of smartplug device")
	credentials := flag.String("credentials", "admin:admin", "credentials specify as <login>:<pass>")
	do := flag.String("do", "info", "enable/disable/info/disableAP/uptime/reboot")
	raw := flag.String("raw", "", "raw command to execute on device (via http)")
	rawt := flag.String("rawt", "", "raw command to execute on device (via telnet)")
	daemon := flag.Bool("daemon", true, "run as a (foreground) daemon with polling webserver")
	port := flag.Int("port", 8080, "webserver port (only used with -daemon)")
	mydebug := flag.Bool("debug", false, "show debug information")
	csvfile := flag.String("csvfile", "output.csv", "file to write csv output to (only used with -daemon)")
	info := flag.String("info", "W", "W/E/V/I\n\t\tW = centiWatt \n\t\tE = milliWatts/h\n\t\tV = milliVolts\n\t\tI = milliAmps")
	flag.Parse()

	debug = *mydebug

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		return
	}
	p := plug{device: *device, credentials: *credentials, csvfile: *csvfile}

	if *raw != "" {
		p.raw(*raw)
		return
	}

	if *rawt != "" {
		p.rawt(*rawt)
		return
	}

	if *daemon {
		listener, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
		if err != nil {
			log.Fatal(err)
		}
		go http.Serve(listener, nil)
		//go http.ListenAndServe(":"+strconv.Itoa(*port), nil)
		p.daemon()
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
