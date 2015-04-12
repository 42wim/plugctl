package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"errors"
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

func (p *plug) info(info string) (string, error) {
	textarea := p.exec(plugInfo + info)
	re := regexp.MustCompile("01(I|V|W|E)[0-9]+ 0*([0-9]+)")
	result := re.FindStringSubmatch(textarea)
	// if we don't have 2 matches something is wrong
	if len(result) > 2 {
		return string(result[2]), nil
	} else {
		return "", errors.New("info not found")
	}
}

func (p *plug) toggle(reallytoggle bool) error {
	status, err := p.status()
	if err != nil {
		return errors.New("Can't get power status")
	}
	if status == 1 {
		fmt.Println("Power is on")
		if reallytoggle {
			p.disable()
		}
	} else {
		fmt.Println("Power is off")
		if reallytoggle {
			p.enable()
		}
	}
	return nil
}

func (p *plug) status() (int, error) {
	result, err := p.info("W")
	if err != nil {
		return -1, err
	}
	resint, _ := strconv.Atoi(result)
	if resint > 100 {
		return 1, nil
	} else {
		return 0, nil
	}
}

func (p *plug) infofull() string {
	resultStr := ""
	conn, err := p.DialTimeout("tcp", p.device, time.Duration(time.Second*3))
	if err != nil {
		log.Fatal("can't connect")
	}
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
		case 0:
			// mAmp/10 -> Amp
			value = strconv.FormatFloat(temp/10000, 'f', 2, 32)
			value = value + " Ampere - "
			// centiWatt -> Watt
		case 1:
			value = strconv.FormatFloat(temp/100, 'f', 2, 32)
			value = value + " Watt - "
			// mWatt/h -> Watt/h
		case 2:
			value = strconv.FormatFloat(temp/1000, 'f', 2, 32)
			value = value + " Watt/hour - "
			// mVolt -> Volt
		case 3:
			value = strconv.FormatFloat(temp/1000, 'f', 2, 32)
			value = value + " Volt"
		}
		resultStr = resultStr + value
	}
	return resultStr
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
	result = p.exec(plugGetAP)
	fmt.Print("saving state...")
	if result == plugResultAPundef || result == plugResultAPon {
		_ = p.exec(plugSetAP + "+0")
		result = p.exec(plugGetAP)
		if result == plugResultAPoff {
			fmt.Println("success")
		} else {
			fmt.Println("failed")
		}
	} else {
		fmt.Println("already set")
	}
}

func (p *plug) enableAP() {
	fmt.Print("enabling AP...")
	p.exec(plugEnableAP)
	result := p.exec(plugIfconfig)
	if strings.Contains(result, "ra0") {
		fmt.Println("success")
	} else {
		fmt.Println("failed")
	}
	result = p.exec(plugGetAP)
	fmt.Print("saving state...")
	if result == plugResultAPundef || result == plugResultAPoff {
		_ = p.exec(plugSetAP + "+1")
		result = p.exec(plugGetAP)
		if result == plugResultAPon {
			fmt.Println("success")
		} else {
			fmt.Println("failed")
		}
	} else {
		fmt.Println("already set")
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
	var diskwriter *csv.Writer
	var gzipwriter *gzip.Writer
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

	// compressed or not
	if strings.Contains(p.csvfile, ".gz") {
		gzipwriter, _ = gzip.NewWriterLevel(csvfile, gzip.BestCompression)
		defer gzipwriter.Close()
		// wrap csv around gzipwriter
		diskwriter = csv.NewWriter(gzipwriter)
	} else {
		// create a diskwriter (appends to csv on disk)
		diskwriter = csv.NewWriter(csvfile)
	}

	// connect via telnet to the device and login
	conn, err := p.DialTimeout("tcp", p.device, time.Duration(time.Second*30))
	if err != nil {
		log.Fatal("can't connect")
	}

	// create http handlers
	http.HandleFunc("/quit", webQuitHandler(diskwriter, gzipwriter, csvfile))
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
			case 0:
				// mAmp/10 -> Amp
				value = strconv.FormatFloat(temp/10000, 'f', 2, 32)
			// centiWatt -> Watt
			case 1:
				value = strconv.FormatFloat(temp/100, 'f', 2, 32)
			// mWatt/h -> Watt/h | mVolt -> Volt
			case 2, 3:
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
			if strings.Contains(p.csvfile, ".gz") {
				gzipwriter.Flush()
			}
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
