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

    "github.com/gorilla/mux"
)

var appConfig = &Config{}

func main() {
    appConfig.LoadConfig("/etc/dyndns.json")

    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/update", Update).Methods("GET")

    log.Println(fmt.Sprintf("Serving dyndns REST services on 0.0.0.0:8080..."))
    log.Fatal(http.ListenAndServe(":8080", router))
}

func Update(w http.ResponseWriter, r *http.Request) {
    response := BuildWebserviceResponseFromRequest(r, appConfig)

    if response.Success == false {
        json.NewEncoder(w).Encode(response)
        return
    }

    for _, domain := range response.Domains {
        result := UpdateRecord(domain, response.Address, response.AddrType)

        if result != "" {
            response.Success = false
            response.Message = result

            json.NewEncoder(w).Encode(response)
            return
        }
    }

    response.Success = true
    response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)

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
