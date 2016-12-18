package main

import (
    "log"
    "fmt"
    "net/http"
    "io/ioutil"
    "os"
    "bufio"
    "os/exec"
    "bytes"
    "encoding/json"
    "net"

    "github.com/gorilla/mux"
)

var appConfig = &Config{}

type WebserviceResponse struct {
    Success bool
    Message string
}

func main() {
    appConfig.LoadConfig("/etc/dyndns.json")

    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/update", Update).Methods("GET")

    log.Println(fmt.Sprintf("Serving dyndns REST services on 0.0.0.0:8080..."))
    log.Fatal(http.ListenAndServe(":8080", router))
}

func validIP4(ipAddress string) bool {
    testInput := net.ParseIP(ipAddress)
    if testInput == nil {
        return false
    }

    return (testInput.To4() != nil)
}

func validIP6(ip6Address string) bool {
    testInputIP6 := net.ParseIP(ip6Address)
    if testInputIP6 == nil {
        return false
    }

    return (testInputIP6.To16() != nil)
}

func Update(w http.ResponseWriter, r *http.Request) {
    response := WebserviceResponse{}

    var sharedSecret string
    var domain string
    var address string

    vals := r.URL.Query()
    sharedSecret = vals["secret"][0]
    domain = vals["domain"][0]
    address = vals["addr"][0]

    if sharedSecret != appConfig.SharedSecret {
		log.Println(fmt.Sprintf("Invalid shared secret: %s", sharedSecret))
        response.Success = false
        response.Message = "Invalid Credentials"
        json.NewEncoder(w).Encode(response)
        return;
    }

    w.Header().Set("Content-Type", "application/json")

    var addrType string

    if validIP4(address) {
        addrType = "A"
    } else if validIP6(address) {
        addrType = "AAAA"
    } else {
        response.Success = false
        response.Message = fmt.Sprintf("%s is neither a valid IPv4 nor IPv6 address", address)
    }

    if addrType != "" {
		if domain == "" {
	        response.Success = false
	        response.Message = fmt.Sprintf("Domain not set", address)
			log.Println(fmt.Sprintf("Domain not set"))
			return;
		}

        result := UpdateRecord(domain, address, addrType)

        if result == "" {
            response.Success = true
            response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", addrType, domain, address)
        } else {
            response.Success = false
            response.Message = result
        }
    }

    json.NewEncoder(w).Encode(response)
}

func UpdateRecord(domain string, ipaddr string, addrType string) string {
    log.Println(fmt.Sprintf("%s record update request: %s -> %s", addrType, domain, ipaddr))

    f, err := ioutil.TempFile(os.TempDir(), "dyndns")
    if err != nil {
        return err.Error()
    }

    defer os.Remove(f.Name())
    w := bufio.NewWriter(f)

    w.WriteString(fmt.Sprintf("server %s\n", appConfig.Server))
    w.WriteString(fmt.Sprintf("zone %s\n", appConfig.Zone))
    w.WriteString(fmt.Sprintf("update delete %s.%s A\n", domain, appConfig.Domain))
    w.WriteString(fmt.Sprintf("update delete %s.%s AAAA\n", domain, appConfig.Domain))
    w.WriteString(fmt.Sprintf("update add %s.%s %v %s %s\n", domain, appConfig.Domain, appConfig.RecordTTL, addrType, ipaddr))
    w.WriteString("send\n")

    w.Flush()
    f.Close()

    cmd := exec.Command(appConfig.NsupdateBinary, f.Name())
    var out bytes.Buffer
    var stderr bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &stderr
    err = cmd.Run()
    if err != nil {
        return err.Error() + ": " + stderr.String()
    }

    return out.String()
}
