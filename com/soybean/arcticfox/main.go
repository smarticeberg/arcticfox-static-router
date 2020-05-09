package main

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var windows_absolute_url = "D:\\webroot\\card.dushu.io\\"
var linux_absolute_url = "/webroot/card.dushu.io/"

var ip_tables = map[string]string{
	"192.168.2.1":  "test17",
	"192.168.2.12": "test18",
}

var zk_host = []string{"192.168.1.59:2181"}

func main() {
	// deal static resource route
	http.HandleFunc("/", dealStaticRoute)
	// set listening port.
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenService erro", err)
	}

	fmt.Printf("this is arcticfox static router.")
}

func init() {
	conn, _, err := zk.Connect(zk_host, time.Second*5)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
}

/**
handle linux or windows absolute url.
** linux:/webroot/card.dushu.io/test17/...
** windows: D:\\webroot\\card.dushu.io\\test17\\...
*/
func decodeAbsoluteUrl() string {
	dir, err := filepath.Abs(filepath.Dir(windows_absolute_url))
	if err != nil {
		log.Fatal(err)
	}

	return strings.Replace(dir, "\\", "/", -1)
}

func dealStaticRoute(w http.ResponseWriter, r *http.Request) {

	url := decodeAbsoluteUrl()
	ip := clientPublicIP(r)
	if ip == "" {
		ip = clientIp(r)
	}

	env := ip_tables[ip]
	if env == "" {
		env = "default"
	}
	// acquire static resources absolute path.
	file := url + "/" + env + r.URL.Path
	fmt.Println(file)

	f, err := os.Open(file)
	defer f.Close()

	if err != nil && os.IsNotExist(err) {
		fmt.Fprintln(w, "File not exist")
		return
	}

	http.ServeFile(w, r, file)
	return

}

/**
get client realip
*/
func clientIp(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

func clientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !HasLocalIPddr(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" && !HasLocalIPddr(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if !HasLocalIPddr(ip) {
			return ip
		}
	}

	return ""
}

// HasLocalIPddr 检测 IP 地址字符串是否是内网地址
func HasLocalIPddr(ip string) bool {
	return HasLocalIP(net.ParseIP(ip))
}

// HasLocalIP 检测 IP 地址是否是内网地址
func HasLocalIP(ip net.IP) bool {
	return ip.IsLoopback()
}
