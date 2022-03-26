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
	DeleteRecord(r RecordUpdateRequest)
	Close() string
}

type UpdateRequestAction int

const (
	UpdateRequestActionUpdate UpdateRequestAction = iota
	UpdateRequestActionDelete
)

// RecordUpdateRequest data representing a update request
type RecordUpdateRequest struct {
	domain      string
	ipAddr      string
	addrType    string
	ddnsKeyName string
	secret      string
	zone        string
	fqdn        string
	action      UpdateRequestAction
}

// NSUpdate holds resources need for an open nsupdate program
type NSUpdate struct {
	cmd       *exec.Cmd
	w         *bufio.Writer
	stdinPipe io.WriteCloser
	out       bytes.Buffer
	stderr    bytes.Buffer
	authSent  bool
	needSend  bool
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
	nsupdate.authSent = false
	nsupdate.needSend = false

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

	if nsupdate.needSend {
		nsupdate.Send()
	}

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
	nsupdate.Auth(r)
	nsupdate.DeleteRecord(r)
	nsupdate.write("update add %s %v %s %s\n", r.fqdn, appConfig.RecordTTL, r.addrType, escape(r.ipAddr))
	nsupdate.needSend = true
}

// DeleteRecord sends the record delete request to the nsupdate program
func (nsupdate *NSUpdate) DeleteRecord(r RecordUpdateRequest) {
	nsupdate.Auth(r)
	nsupdate.write("update delete %s %s\n", r.fqdn, escape(r.addrType))
	nsupdate.needSend = true
}

// DeleteRecord sends auth if needed
func (nsupdate *NSUpdate) Auth(r RecordUpdateRequest) {
	if nsupdate.authSent {
		return
	}
	nsupdate.write("server %s\n", appConfig.Server)
	if r.zone != "" {
		nsupdate.write("zone %s\n", r.zone+".")
	}
	if r.ddnsKeyName != "" {
		nsupdate.write("key hmac-sha256:ddns-key.%s %s\n", escape(r.ddnsKeyName), escape(r.secret))
	}
	nsupdate.authSent = true
}

func (nsupdate *NSUpdate) Send() {
	nsupdate.write("send\n")
}
