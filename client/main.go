package main

import (
	"bufio"
	"bytes"
	"common"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var port int
var responseTime chan time.Duration

func init() {
	flag.IntVar(&port, "port", 10000, "server port")
}

func usage() {
	flag.PrintDefaults()
}

var url string

var titleMap = []string{
	"新增商品",
	"看商品",
	"从购物车减少商品",
	"清空购物车",
	"结算购物车",
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	url = fmt.Sprintf("http://localhost:%d", port)

	responseTime = make(chan time.Duration, 1)

	fmt.Println("欢迎光临线上购物书城")
	for {
		fmt.Println()
		fmt.Println("购物车:")
		printCartList()
		fmt.Println()
		fmt.Println("请问您要：")
		for k, v := range titleMap {
			if common.TypeMap[k] == common.BLUE {
				fmt.Printf("%c[%d;%d;%dm%s %d. %s%c[0m ", 0x1B, 0, 40, 36, "", k+1, v, 0x1B)
				fmt.Println()
			} else {
				fmt.Printf("%c[%d;%d;%dm%s %d. %s%c[0m ", 0x1B, 0, 40, 31, "", k+1, v, 0x1B)
				fmt.Println()
			}
		}
		fmt.Print("请选择：")
		input := <-readStdin()
		switch input {
		case "1":
			AddItemOption()
		case "2":
			ReadItemListOption()
		case "3":
			RmItemOption()
		case "4":
			ClearCartOption()
		case "5":
			CheckoutOption()
		default:
			break
		}
		fmt.Printf("%c[%d;%d;%dm%s===========================================%c[0m ", 0x1B, 0, 40, 31, "", 0x1B)
	}
}

func request(method, api string, j []byte, benchmark bool) *http.Response {
	var start time.Time
	if benchmark {
		start = time.Now()
	}
	req, err := http.NewRequest(method, url+api, bytes.NewBuffer(j))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if benchmark {
		elapsed := time.Since(start)
		//fmt.Println()
		//fmt.Printf("%c[%d;%d;%dm%s耗时: %s%c[0m ", 0x1B, 0, 40, 31, "", elapsed, 0x1B)
		//fmt.Println()
		responseTime <- elapsed
	}
	return resp
}

func printResponceTime() {
	fmt.Println()
	fmt.Printf("%c[%d;%d;%dm%s耗时: %s%c[0m ", 0x1B, 0, 40, 31, "", <-responseTime, 0x1B)
	fmt.Println()
	fmt.Println()

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

func printCartList() []common.Item {
	cartList := []common.Item{}
	resp := request("GET", "/mycarts", []byte{}, false)
	decoder := json.NewDecoder(resp.Body)
	var response common.Response
	err := decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}
	msgs := response.Msg.([]interface{})
	sum := 0
	for k, value := range msgs {
		v := value.(map[string]interface{})
		item := common.Item{v["name"].(string), uint32(v["volume"].(float64)), v["id"].(string), uint32(v["price"].(float64))}
		cartList = append(cartList, item)
		fmt.Print(k + 1)
		fmt.Print(". ", item.Name, " - ")
		fmt.Print(item.Volume)
		fmt.Print(" X ")
		fmt.Println(item.Price, "元 =", item.Volume*item.Price, "元")
		sum += int(item.Volume * item.Price)
	}
	fmt.Println("总共: ", sum, "元")
	return cartList
}

func AddItemOption() {
	fmt.Println()
	fmt.Print("商品名称: ")
	name := <-readStdin()
	fmt.Print("商品售价: ")
	input := <-readStdin()
	price, _ := strconv.Atoi(input)
	fmt.Print("商品数量: ")
	input = <-readStdin()
	volume, _ := strconv.Atoi(input)

	jsonStr := fmt.Sprintf(`{"name":"%s","price":%d ,"volume":%d}`, name, price, volume)

	resp := request("POST", "/newitem", []byte(jsonStr), true)

	decoder := json.NewDecoder(resp.Body)
	var response common.Response
	err := decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}
	if response.Succeed {
		fmt.Println("	新增商品成功！")
	} else {
		fmt.Println("	新增商品失败！")
	}

	printResponceTime()
}

func ReadItemListOption() {
	resp := request("GET", "/items", []byte{}, false)
	decoder := json.NewDecoder(resp.Body)
	var response common.Response
	err := decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println()
	msgs := response.Msg.([]interface{})
	items := []common.Item{}
	for k, value := range msgs {
		v := value.(map[string]interface{})
		item := common.Item{v["name"].(string), uint32(v["volume"].(float64)), v["id"].(string), uint32(v["price"].(float64))}
		fmt.Print(k + 1)
		fmt.Print(". ", item.Name, " - ")
		fmt.Print(item.Price)
		fmt.Print("元 - ")
		fmt.Println("剩余", item.Volume, "个")
		items = append(items, item)
	}

	var input string
	index := -1
	for index > len(items) || index <= 0 {
		if index == 0 {
			return
		}

		fmt.Print("请选择您想要加入购物车的商品：(0 退出)")
		input = <-readStdin()
		index, _ = strconv.Atoi(input)
	}

	num := -1
	for uint32(num) > items[index-1].Volume || num == -1 {
		fmt.Print("请选择数量：")
		input = <-readStdin()
		num, _ = strconv.Atoi(input)
	}

	jsonStr := fmt.Sprintf(`{"id":"%s", "volume":%d}`, items[index-1].ID, num)
	resp = request("POST", "/additem", []byte(jsonStr), true)
	decoder = json.NewDecoder(resp.Body)
	err = decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}
	if response.Succeed {
		fmt.Println("	加入购物车成功！")
	} else {
		fmt.Println("	加入购物车失败！")
	}

	printResponceTime()
}

func RmItemOption() {
	// id(string(10))
	cartList := printCartList()

	var input string
	index := -1
	for index > len(cartList) || index <= 0 {
		if index == 0 {
			return
		}

		fmt.Println("你想从购物车减少哪种商品?(0 返回)")
		input = <-readStdin()
		index, _ = strconv.Atoi(input)
	}

	num := -1
	for uint32(num) > cartList[index-1].Volume || num == -1 {
		fmt.Print("请选择数量：")
		input = <-readStdin()
		num, _ = strconv.Atoi(input)
	}

	jsonStr := fmt.Sprintf(`{"id":"%s", "volume":%d}`, cartList[index-1].ID, num)
	//var jsonStr = []byte(`{"id":"l3k4l1n3x1m3"}`)

	var response common.Response
	resp := request("POST", "/removeitem", []byte(jsonStr), true)
	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}
	if response.Succeed {
		fmt.Println("	操作购物车成功！")
	} else {
		fmt.Println("	操作购物车失败！")
	}
	printResponceTime()
}

func ClearCartOption() {
	// id(string,volume(string)
	//var jsonStr = []byte(`{"id":"l3k4l1n3x1m3,34dsd214dsd,23fsdfd123", "volume":"2,3,1"}`)
	resp := request("POST", "/clear", []byte{}, true)

	var response common.Response
	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}
	if response.Succeed {
		fmt.Println("	清空购物车成功！")
	} else {
		fmt.Println("	清空购物车失败！")
	}
	printResponceTime()

}

func CheckoutOption() {
	// id(string,volume(string)
	//var jsonStr = []byte(`{"id":"l3k4l1n3x1m3,34dsd214dsd,23fsdfd123", "volume":"2,3,1"}`)
	//var jsonStr = []byte({})

	fmt.Println("\n您打算购买:")
	printCartList()
	resp := request("POST", "/checkout", []byte{}, true)

	var response common.Response
	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&response)
	if err != nil {
		fmt.Println(err)
	}
	if response.Succeed {
		fmt.Println("	结账成功！")
	} else {
		fmt.Println("	结账失败！")
	}
	printResponceTime()

}
