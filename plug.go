package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type plug struct {
	device      string
	credentials string
	csvfile     string
	latestEntry []byte
	buffer      bytes.Buffer
	delay       int
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

func (p *plug) disableCloud() {
	fmt.Println("disabling Cloud Access...")
	for _, address := range plugDisableCloudAddresses {
		fmt.Print("blocking " + address + " ...")
		p.exec(plugRoute + "+add+" + address + "+dev+lo")
		result := p.exec(plugRoute)
		if strings.Contains(result, address) {
			fmt.Println("success")
		} else {
			fmt.Println("failed")
		}
	}
}

func (p *plug) enableCloud() {
	fmt.Println("enabling Cloud Access...")
	for _, address := range plugDisableCloudAddresses {
		fmt.Print("unblocking " + address + " ...")
		p.exec(plugRoute + "+del+" + address + "+dev+lo")
		result := p.exec(plugRoute)
		if strings.Contains(result, address) {
			fmt.Println("failed")
		} else {
			fmt.Println("success")
		}
	}
}

func (p *plug) raw(command string) {
	fmt.Println("executing command:", command)
	command = strings.Replace(command, " ", "+", -1)
	result := p.exec(command)
	result = strings.Replace(result, "||", "\n", -1)
	fmt.Println(result)
}

func (p *plug) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	s := strings.Split(p.credentials, ":")
	login, pass := s[0], s[1]
	if err != nil {
		return nil, err
	}
	sendln(conn, "", '\n')
	sendln(conn, login, '\n')
	sendln(conn, pass, '#')
	return conn, err
}

// use telnet connection
func (p *plug) rawt(command string) {
	conn, err := p.DialTimeout("tcp", p.device, time.Duration(time.Second*30))
	if err != nil {
		log.Fatal("can't connect")
	}
	status := sendln(conn, command, '#')
	status = strings.Replace(status, command+"\r\n", "", 1)
	status = strings.Replace(status, "#", "", 1)
	fmt.Print(status)
}

func (p *plug) daemon() {
	fmt.Println("starting foreground daemon ;-)")

	// write csv from disk into the buffer
	fmt.Println("loading history (" + p.csvfile + ")")
	p.buffer.Write(readcsv(p.csvfile))

	// create/append the csvfile on disk
	csvfile, err := os.OpenFile(p.csvfile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatal("Error:", err)
	}
	defer csvfile.Close()

	// create a bufferwriter (appends to csv already in p.buffer)
	bufferwriter := csv.NewWriter(&p.buffer)

	// create a diskwriter (appends to csv on disk)
	diskwriter := csv.NewWriter(csvfile)

	// connect via telnet to the device and login
	conn, err := p.DialTimeout("tcp", p.device, time.Duration(time.Second*30))
	if err != nil {
		log.Fatal("can't connect")
	}

	// create http handlers
	http.HandleFunc("/quit", webQuitHandler(diskwriter))
	http.HandleFunc("/history", webHistoryHandler)
	http.HandleFunc("/stream", webStreamHandler)
	http.HandleFunc("/read.csv", webReadCsvHandler(p))
	http.HandleFunc("/read.json", webReadJsonHandler(p))

	// needed for occasionally flushing on a newline
	recordcount := 0

	// start infinite polling loop
	for {
		// measure how long it takes
		start := time.Now()

		// specify correct format for dygraph
		record := []string{start.Format("2006/01/02 15:04:05")}

		// get statistics from device and cleanup
		status := sendln(conn, plugGetInfoStats, '#')
		status = strings.Replace(status, plugGetInfoStats+"\r\n", "", 1)
		status = strings.Replace(status, "#", "", 1)
		// split up the 4 results a newline
		results := strings.SplitN(status, "\r\n", 4)

		re := regexp.MustCompile("01(I|V|W|E)[0-9]+ 0*([0-9]+)")
		// for each GetInfo result, do a regexp match, adjust value and create a CSV record
		for i, result := range results {
			match := re.FindStringSubmatch(result)
			value := "0"
			// check if we got the right size of slice
			if len(match) == 3 {
				value = match[2]
			}

			temp, _ := strconv.ParseFloat(value, 32)

			switch i {
			// centiWatt -> Watt
			case 1:
				value = strconv.FormatFloat(temp/100, 'f', 2, 32)
			// mAmp -> Amp | mWatt/h -> Watt/h | mVolt -> Volt
			case 0, 2, 3:
				value = strconv.FormatFloat(temp/1000, 'f', 2, 32)
			}
			record = append(record, value)
			recordcount += 1
		}

		// latestentry is needed in JSON for the realtime streaming
		p.latestEntry, _ = json.Marshal(record)

		// write the record to disk
		err := diskwriter.Write(record)
		if err != nil {
			fmt.Println("Error:", err)
		}

		// write the record to buffer (in memory)
		err = bufferwriter.Write(record)
		if err != nil {
			fmt.Println("Error:", err)
		}

		// flush disk every 25 records
		if recordcount%100 == 0 {
			diskwriter.Flush()
		}
		// flush memory immediately
		bufferwriter.Flush()

		if debug {
			fmt.Print(record)
			fmt.Println(" took", time.Since(start))
		}
		// sleep the right amount of time
		time.Sleep(time.Second*time.Duration(p.delay) - time.Since(start))
	}
}
