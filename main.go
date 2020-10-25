package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"
)

var (
	flagListen string
	flagDst    string
	flagFormat string
)

var (
	listenAddr *net.TCPAddr
	dstAddr    *net.TCPAddr
)

func main() {
	flag.StringVar(&flagListen, "l", "0.0.0.0:9001", `[ip:]port`)
	flag.StringVar(&flagDst, "d", "127.0.0.1:9002", `ip/dns:port`)
	flag.StringVar(&flagFormat, "f", "", "json")
	flag.Parse()

	var err error

	if listenAddr, err = net.ResolveTCPAddr("tcp", flagListen); err != nil {
		logRun(fmt.Sprintf("flagListen: %v", err))
		os.Exit(0)
	}

	if dstAddr, err = net.ResolveTCPAddr("tcp", flagDst); err != nil {
		logRun(fmt.Sprintf("flagDst: %v", err))
		os.Exit(0)
	}

	go l()
	select {}
}

func l() {
	defer func() {
		if rev := recover(); rev != nil {
			logRun(fmt.Sprintf("Listen: recover, %v", rev))
			go l()
		}
	}()

	l, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		logRun(fmt.Sprintf("Listen: %v", err))
		os.Exit(0)
	}

	for {
		srcConn, err := l.AcceptTCP()
		if err != nil {
			logRun(fmt.Sprintf("Accept: %v", err))
			continue
		}

		dstConn, err := net.DialTCP("tcp", nil, dstAddr)
		if err != nil {
			logRun(fmt.Sprintf("Dial: %v", err))
			srcConn.Close()
			continue
		}

		// go io.Copy(srcConn, dstConn)
		// go io.Copy(dstConn, srcConn)

		srcLabel := fmt.Sprintf("src(%s)", srcConn.RemoteAddr().String())
		dstLabel := fmt.Sprintf("dst(%s)", dstConn.LocalAddr().String())

		go copy(dstConn, dstLabel, srcConn, srcLabel)
		go copy(srcConn, srcLabel, dstConn, dstLabel)

	}
}

// copy copy
func copy(dstConn *net.TCPConn, dstLabel string, srcConn *net.TCPConn, srcLabel string) {
	defer func() {
		if rev := recover(); rev != nil {
			logCopy(fmt.Sprintf("recover: %v", rev), srcLabel, dstLabel)
		}
		srcConn.Close()
		dstConn.Close()
	}()

	r := bufio.NewReader(srcConn)
	for {
		line, isPrefix, err := r.ReadLine()
		if err != nil {
			logCopy(fmt.Sprintf("ReadLine: %v", err), srcLabel, dstLabel)
			return
		}
		if isPrefix {
			fmt.Printf("isPrefix: %s %s: big than 4096? , line: %s", srcLabel, dstLabel, line)
			continue
		}

		var lineStr = ""
		switch flagFormat {
		case "json":
			var prettyJSON bytes.Buffer
			err = json.Indent(&prettyJSON, line, "", " ")
			if err != nil {
				lineStr = string(line)
			} else {
				lineStr = string(prettyJSON.Bytes())
			}
		default:
			lineStr = string(line)
		}
		logCopy(lineStr, srcLabel, dstLabel)

		dstConn.Write(line)
		dstConn.Write([]byte("\n"))
	}
}

// runLog
func logRun(msg string) {
	fmt.Printf("%s : %s\n", time.Now().Format("2006-01-02 15:04:05"), msg)
}

func logCopy(msg string, srcLabel string, dstLabel string) {
	fmt.Printf("%s %s->%s\n%s\n", time.Now().Format("2006-01-02 15:04:05"), srcLabel, dstLabel, msg)
}
