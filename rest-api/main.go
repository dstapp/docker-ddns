package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
)

var appConfig = &Config{}
var db *bolt.DB = nil

func main() {
	appConfig.LoadConfig("/etc/dyndns.json")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/update", Update).Methods("GET")

	/* DynDNS compatible handlers. Most routers will invoke /nic/update */
	router.HandleFunc("/nic/update", DynUpdate).Methods("GET")
	router.HandleFunc("/v2/update", DynUpdate).Methods("GET")
	router.HandleFunc("/v3/update", DynUpdate).Methods("GET")

	db, _ = bolt.Open("dyndns.db", 0600, nil)
	defer db.Close()

	go databaseMaintenance(db)

	log.Println(fmt.Sprintf("Serving dyndns REST services on 0.0.0.0:8080..."))
	log.Fatal(http.ListenAndServe(":8080", router))
}

func DynUpdate(w http.ResponseWriter, r *http.Request) {
	extractor := RequestDataExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("myip") },
		Secret: func(r *http.Request) string {
			_, sharedSecret, ok := r.BasicAuth()
			if !ok || sharedSecret == "" {
				sharedSecret = r.URL.Query().Get("password")
			}

			return sharedSecret
		},
		Domain: func(r *http.Request) string { return r.URL.Query().Get("hostname") },
	}
	response := BuildWebserviceResponseFromRequest(r, appConfig, extractor)

	if response.Success == false {
		if response.Message == "Domain not set" {
			w.Write([]byte("notfqdn\n"))
		} else {
			w.Write([]byte("badauth\n"))
		}
		return
	}

	for _, domain := range response.Domains {
		result := UpdateRecord(domain, response.Address, response.AddrType)

		if result != "" {
			response.Success = false
			response.Message = result

			w.Write([]byte("dnserr\n"))
			return
		}
	}

	response.Success = true
	response.Message = fmt.Sprintf("Updated %s record for %s to IP address %s", response.AddrType, response.Domain, response.Address)

	w.Write([]byte(fmt.Sprintf("good %s\n", response.Address)))
}

func Update(w http.ResponseWriter, r *http.Request) {
	extractor := RequestDataExtractor{
		Address: func(r *http.Request) string { return r.URL.Query().Get("addr") },
		Secret:  func(r *http.Request) string { return r.URL.Query().Get("secret") },
		Domain:  func(r *http.Request) string { return r.URL.Query().Get("domain") },
	}
	response := BuildWebserviceResponseFromRequest(r, appConfig, extractor)

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
	w.WriteString(fmt.Sprintf("update delete %s.%s %s\n", domain, appConfig.Domain, addrType))
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

	/* Create a resource record in the database */
	if err := db.Update(func(tx *bolt.Tx) error {
		rr, err := tx.CreateBucketIfNotExists([]byte(domain))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = rr.Put([]byte("domain"), []byte(domain))
		err = rr.Put([]byte("zone"), []byte(appConfig.Domain))
		err = rr.Put([]byte("ttl"), []byte(fmt.Sprintf("%v", appConfig.RecordTTL)))
		err = rr.Put([]byte("type"), []byte(addrType))
		err = rr.Put([]byte("address"), []byte(ipaddr))

		t := time.Now()
		err = rr.Put([]byte("expiry"), []byte(t.Add(time.Second * time.Duration(appConfig.RecordExpiry)).Format(time.RFC3339)))
		err = rr.Put([]byte("created"), []byte(t.Format(time.RFC3339)))

		return nil
	}); err != nil {
		log.Print(err)
	}

	return out.String()
}

/* GO func to clean up expired entries */
func databaseMaintenance(db *bolt.DB) {
	cleanupTicker := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-cleanupTicker.C:
			now := []byte(time.Now().Format(time.RFC3339))
			key := []byte("expiry")

			if err := db.View(func(tx *bolt.Tx) error {
				/* Iterate through all buckets (each is a resource record) */
				err := tx.ForEach(func(name []byte, b *bolt.Bucket) error {
					c := b.Cursor()

					if k, v := c.Seek(key); k != nil && bytes.Equal(k, key) {
						// Check for expiry
						if bytes.Compare(v, now) < 0 {
							if k, v := c.Seek([]byte("type")); k != nil {
								log.Printf("Expired RR(%s): '%s'. Deleting.", string(v), string(name))
								go deleteRecord(db, string(name), string(v))
							}
						}
					}

					return nil
				})

				if err != nil {
					log.Print(err)
				}
				return nil
			}); err != nil {
				log.Print(err)
			}
		}
	}
}

/* GO func to delete an entry once expired */
func deleteRecord(db *bolt.DB, name string, addrType string) {
	db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(name))
	})

	f, _ := ioutil.TempFile(os.TempDir(), "dyndns_cleanup")

	defer os.Remove(f.Name())
	w := bufio.NewWriter(f)

	w.WriteString(fmt.Sprintf("server %s\n", appConfig.Server))
	w.WriteString(fmt.Sprintf("zone %s\n", appConfig.Zone))
	w.WriteString(fmt.Sprintf("update delete %s.%s %s\n", name, appConfig.Domain, addrType))
	w.WriteString("send\n")

	w.Flush()
	f.Close()

	cmd := exec.Command(appConfig.NsupdateBinary, f.Name())
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	cmd.Run()
}