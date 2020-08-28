package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"unicode"
)

// NSUpdateInterface is the interface to a client which can update a DNS record
type NSUpdateInterface interface {
	UpdateRecord(r RecordUpdateRequest)
	Close() string
}

// RecordUpdateRequest data representing a update request
type RecordUpdateRequest struct {
	domain   string
	ipaddr   string
	addrType string
	ddnskey  string
}

// NSUpdate holds resources need for an open nsupdate program
type NSUpdate struct {
	cmd       *exec.Cmd
	w         *bufio.Writer
	stdinPipe io.WriteCloser
	out       bytes.Buffer
	stderr    bytes.Buffer
}

// NewNSUpdate starts the nsupdate program
func NewNSUpdate() *NSUpdate {
	var err error

	var nsupdate = &NSUpdate{}
	nsupdate.cmd = exec.Command(appConfig.NsupdateBinary)

	nsupdate.stdinPipe, err = nsupdate.cmd.StdinPipe()
	if err != nil {
		log.Println(err.Error() + ": " + nsupdate.stderr.String())
		return nil
	}

	nsupdate.cmd.Stdout = &nsupdate.out
	nsupdate.cmd.Stderr = &nsupdate.stderr
	err = nsupdate.cmd.Start()
	if err != nil {
		log.Println(err.Error() + ": " + nsupdate.stderr.String())
		return nil
	}
	nsupdate.w = bufio.NewWriter(nsupdate.stdinPipe)

	return nsupdate
}

func (nsupdate *NSUpdate) write(format string, a ...interface{}) {
	command := fmt.Sprintf(format, a...)
	if appConfig.LogLevel >= 1 {
		logCommand := strings.Replace(command, "\n", "\\n", -1) // ReplaceAll
		log.Println("nsupdate: " + logCommand)
	}
	nsupdate.w.WriteString(command)
}

// Close sends the quit command and waits for the response which is then returned.
func (nsupdate *NSUpdate) Close() string {
	var err error

	nsupdate.write("quit\n")
	nsupdate.w.Flush()
	nsupdate.stdinPipe.Close()

	err = nsupdate.cmd.Wait()
	if err != nil {
		return err.Error() + ": " + nsupdate.stderr.String()
	}

	return nsupdate.out.String()
}

func isRune(r rune, allow string) bool {
	for _, c := range allow {
		if r == c {
			return true
		}
	}
	return false
}

func escape(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !isRune(r, ".+-_/=")
	})
}

// UpdateRecord sends the record update request to the nsupdate program
func (nsupdate *NSUpdate) UpdateRecord(r RecordUpdateRequest) {
	fqdn := escape(r.domain)
	if appConfig.Zone != "" {
		fqdn = escape(r.domain) + "." + appConfig.Zone
	}

	if r.ddnskey != "" {
		fqdnN := strings.TrimLeft(fqdn, ".")
		nsupdate.write("key hmac-sha256:ddns-key.%s %s\n", fqdnN, escape(r.ddnskey))
	}

	nsupdate.write("server %s\n", appConfig.Server)
	nsupdate.write("zone %s\n", appConfig.Zone)
	nsupdate.write("update delete %s %s\n", fqdn, r.addrType)
	nsupdate.write("update add %s %v %s %s\n", fqdn, appConfig.RecordTTL, r.addrType, escape(r.ipaddr))
	nsupdate.write("send\n")
}
