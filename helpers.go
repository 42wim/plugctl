package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
)

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

func readcsv(csvfile string) []byte {
	contents, err := ioutil.ReadFile(csvfile)
	if err != nil {
		if os.IsNotExist(err) {
			// fmt.Println("file does not exist")
		} else {
			fmt.Println("error", err)
		}
	}
	return contents
}

// webhandlers

func webHistoryHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, webHistory)
}

func webStreamHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, webStream)
}

func webQuitHandler(diskwriter *csv.Writer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "flushing to disk and shutting down")
		diskwriter.Flush()
		os.Exit(0)
	}
}

func webReadCsvHandler(p *plug) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		backup := p.buffer
		for {
			output, err := backup.ReadString('\n')
			if err != nil {
				break
			}
			fmt.Fprintf(w, string(output))
		}
	}
}

func webReadJsonHandler(p *plug) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(p.latestEntry))
	}
}
