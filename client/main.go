package main

import (
	"bufio"
	"bytes"
	"common"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
)

var port int

func init() {
	flag.IntVar(&port, "port", common.SERVER_DEFAULT_PORT, "server port")
}

func usage() {
	flag.PrintDefaults()
}

var url string

func main() {
	url = fmt.Sprintf("http://localhost:%d", port)
	fmt.Println("欢迎光临线上购物书城")
	for {
		fmt.Println()
		fmt.Println("购物车:")
		fmt.Println()
		fmt.Println("请问您要：")
		fmt.Println("1. 刷新购物车")
		fmt.Println("2. 看商品")
		fmt.Println("3. 移除某项商品")
		fmt.Println("4. 清空购物车")
		fmt.Println("5. 结算购物车")
		input := <-readStdin()
		switch input {
		case "1":
			var jsonStr = []byte(`{"name":"book"}`)
			resp := request("POST", "/refresh", jsonStr)
			fmt.Println(resp)
		case "2":
			var jsonStr = []byte(`{"name":"book"}`)
			resp := request("POST", "/additem", jsonStr)
			fmt.Println(resp)
		case "3":
			fmt.Println("你想移除哪项商品商品?")
			input := <-readStdin()
			fmt.Println(input)
		case "4":
			// store.ClearShoppingCart()
		case "5":
			// store.SettleShoppingCart()
		default:
			break
		}
	}
}

func request(method, api string, j []byte) *http.Response {
	req, err := http.NewRequest(method, url+api, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	return resp
}

func readStdin() chan string {
	cb := make(chan string)
	input := bufio.NewScanner(os.Stdin)
	go func() {
		if input.Scan() {
			cb <- input.Text()
		}
	}()
	return cb
}

func getIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func findIPAddress(input string) string {
	validIpAddressRegex := "([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})"
	re := regexp.MustCompile(validIpAddressRegex)
	return re.FindString(input)
}
