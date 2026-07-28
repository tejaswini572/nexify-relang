package main
import (
    "fmt"
    "rsc.io/qr/coding"
)
func main() {
    v := coding.Version(2)
    l := coding.L
    
    var b coding.Bits
    coding.String("https://example.com").Encode(&b, v)
    b.AddCheckBytes(v, l)
    
    bytes := b.Bytes()
    for _, byteVal := range bytes {
        fmt.Printf("%02x ", byteVal)
    }
    fmt.Println()
}
