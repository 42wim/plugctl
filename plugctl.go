package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var debug bool = false

func main() {
	device := flag.String("ip", "192.168.8.74", "ipv4 address of smartplug device")
	credentials := flag.String("credentials", "admin:admin", "credentials specify as <login>:<pass>")
	enable := flag.String("enable", "", "power/cloud/ap")
	disable := flag.String("disable", "", "power/cloud/ap")
	show := flag.String("show", "", "info/uptime")
	raw := flag.String("raw", "", "raw command to execute on device (via telnet)")
	daemon := flag.Bool("daemon", false, "run as a (foreground) daemon with polling webserver")
	port := flag.Int("port", 8080, "webserver port (only used with -daemon)")
	delay := flag.Int("delay", 1, "polling delay of statistics in seconds (only used with -daemon)")
	mydebug := flag.Bool("debug", false, "show debug information")
	csvfile := flag.String("csvfile", "output.csv", "file to write csv output to (only used with -daemon)")

	flag.Parse()

	debug = *mydebug

	if len(os.Args) == 1 {
		flag.PrintDefaults()
		return
	}

	if strings.Contains(*device, ":") == false {
		if (!*daemon) && (*show != "info") && (*raw=="") {
			*device = *device + ":80"
		} else {
			*device = *device + ":23"
		}
	}

	p := plug{device: *device, credentials: *credentials, csvfile: *csvfile, delay: *delay}

	if *raw != "" {
		p.rawt(*raw)
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

	if *enable != "" {
		switch *enable {
		case "power":
			p.enable()
		case "ap":
			p.enableAP()
		case "cloud":
			p.enableCloud()
		}
		return
	}

	if *disable != "" {
		switch *disable {
		case "power":
			p.disable()
		case "ap":
			p.disableAP()
		case "cloud":
			p.disableCloud()
		}
		return
	}

	if *show != "" {
		switch *show {
		case "uptime":
			p.uptime()
		case "info":
			fmt.Println(p.infofull())
		}
		return
	}
	flag.PrintDefaults()
}
