package main

import (
    "fmt"
    "crypto/rand"
    "encoding/base64"
)

func randomString() string {
    random := make([]byte, 32)
    _, err := rand.Read(random)
    if err != nil {
        fmt.Println("error:", err)
        return "8Passw0RT!"
    }
    return base64.StdEncoding.EncodeToString(random)
}
