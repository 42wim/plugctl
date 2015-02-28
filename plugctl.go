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

type config struct {
	Main struct {
		Ip          string
		Credentials string
		Csvfile     string
		Port        int
		Daemon      bool
		WebWidth    string
		WebHeight   string
	}
}

var debug bool = false

func main() {
	device := flag.String("ip", "", "ipv4 address of smartplug device")
	credentials := flag.String("credentials", "admin:admin", "credentials specify as <login>:<pass>")
	enable := flag.String("enable", "", "enable power/cloud/ap")
	disable := flag.String("disable", "", "disable power/cloud/ap")
	toggle := flag.String("toggle", "", "toggle power")
	show := flag.String("show", "", "show info/uptime/power")
	raw := flag.String("raw", "", "raw command to execute on device (via telnet)")
	daemon := flag.Bool("daemon", false, "run as a (foreground) daemon with polling webserver")
	port := flag.Int("port", 8080, "webserver port (only used with -daemon)")
	delay := flag.Int("delay", 1, "polling delay of statistics in seconds (only used with -daemon)")
	mydebug := flag.Bool("debug", false, "show debug information")
	csvfile := flag.String("csvfile", "output.csv", "file to write csv output to (only used with -daemon)")
	cfgfile := flag.String("conf", "", "a valid config file (uses plugctl.conf if exists)")

	flag.Parse()

	if *cfgfile == "" {
		if _, err := os.Stat("plugctl.conf"); err == nil {
			*cfgfile = "plugctl.conf"
		}
	}

	if *cfgfile != "" {
		cfg := readconfig(*cfgfile)
		if cfg.Main.Ip != "" {
			*device = cfg.Main.Ip
		}
		if cfg.Main.Credentials != "" {
			*credentials = cfg.Main.Credentials
		}
		if cfg.Main.Csvfile != "" {
			*csvfile = cfg.Main.Csvfile
		}
		if strconv.Itoa(cfg.Main.Port) != "0" {
			*port = cfg.Main.Port
		}
		if cfg.Main.Daemon {
			*daemon = true
		}
		if cfg.Main.WebHeight != "" {
			webHistory = strings.Replace(webHistory, "##WebHeight##", cfg.Main.WebHeight, -1)
			webStream = strings.Replace(webStream, "##WebHeight##", cfg.Main.WebHeight, -1)
		}
		if cfg.Main.WebWidth != "" {
			webHistory = strings.Replace(webHistory, "##WebWidth##", cfg.Main.WebWidth, -1)
			webStream = strings.Replace(webStream, "##WebWidth##", cfg.Main.WebWidth, -1)
		}
	}

	webHistory = strings.Replace(webHistory, "##WebHeight##", WebHeight, -1)
	webHistory = strings.Replace(webHistory, "##WebWidth##", WebWidth, -1)
	webStream = strings.Replace(webStream, "##WebHeight##", WebHeight, -1)
	webStream = strings.Replace(webStream, "##WebWidth##", WebWidth, -1)

	debug = *mydebug

	if len(os.Args) == 1 && !*daemon {
		flag.PrintDefaults()
		return
	}

	if strings.Contains(*device, ":") == false {
		if (!*daemon) && (*show != "info") && (*raw == "") {
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
		p.daemon()
		return
	}

	if *toggle != "" {
		p.toggle(true)
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
		case "power":
			p.toggle(false)
		}
		return
	}
	flag.PrintDefaults()
}
