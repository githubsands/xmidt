package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/Comcast/webpa-common/wrp"
)

const (
	AuthHeaderName    = "Authorization"
	TypeHeaderName    = "X-Xmidt-Message-Type"
	ContentHeaderName = "X-Xmidt-Content-Type"
	TransIdHeaderName = "X-Xmidt-Transaction-Uuid"
	SourceHeaderName  = "X-Xmidt-Source"
	DestHeaderName    = "X-Webpa-Device-Name"
	PathURL           = "https://api.xmidt.comcast.net/api/v2/device"
	ContentType       = "Content-Type"
	XScytaleBuild     = "X-Scytale-Build"
	//XmidtStatus       = "X-Xmidt-Status"
	XScytaleFlavor = "X-Scytale-Flavor"
	XScytaleRegion = "X-Scytale-Region"
	XScytaleServer = "X-Scytale-Server"
)

var auth string

func makeRequest(requestType, messageType, source, transId, dest, contentType, payload string, client *http.Client) {

	headers := make(map[string]string)

	headers[AuthHeaderName] = auth
	var req *http.Request
	var err error

	if "POST" == requestType {
		headers[TypeHeaderName] = messageType
		headers[ContentHeaderName] = contentType
		headers[TransIdHeaderName] = transId
		headers[SourceHeaderName] = source
		headers[DestHeaderName] = dest

		req, err = http.NewRequest(requestType, PathURL, bytes.NewBufferString(payload))
	}

	if err != nil {
		fmt.Printf("Request %s failed: %s, %v\n", requestType, PathURL, err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Printf("Failed to obtain request: %v\n", err)
	} else {

		if 200 != resp.StatusCode {
			fmt.Printf("Failed. resp.StatusCode %v \n", resp.StatusCode)
		} else if 200 == resp.StatusCode {
			/*Decode Msgpack response to JSON format*/
			decoder := wrp.NewDecoder(resp.Body, wrp.Msgpack)
			var buffer bytes.Buffer
			encoder := wrp.NewEncoder(&buffer, wrp.JSON)

			if message, err := wrp.TranscodeMessage(encoder, decoder); err != nil {
				fmt.Println("Error while converting:", wrp.Msgpack, "to", wrp.JSON, err)
			} else {
				//Headers
				fmt.Println(ContentType, ":", resp.Header.Get(ContentType))
				//fmt.Println(XmidtStatus, ":", resp.Header.Get(XmidtStatus))
				fmt.Println(XScytaleBuild, ":", resp.Header.Get(XScytaleBuild))
				fmt.Println(XScytaleFlavor, ":", resp.Header.Get(XScytaleFlavor))
				fmt.Println(XScytaleRegion, ":", resp.Header.Get(XScytaleRegion))
				fmt.Println(XScytaleServer, ":", resp.Header.Get(XScytaleServer))
				//Response from cloud
				fmt.Println("WRP response", string(buffer.Bytes()))
				fmt.Println("payload :", string(message.Payload))
			}

		} else {
			fmt.Println("Request failed with status ", resp.StatusCode)
		}
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}
}

func main() {
	if "" == os.Getenv("WEBPA_SAT") {
		fmt.Printf("Environment variable WEBPA_SAT must be set a valid SAT token.")
		return
	}

	auth = "Bearer " + os.Getenv("WEBPA_SAT")

	var maxConnReuse uint
	var requestType, messageType, source, dest, contentType, transId, payload string

	flag.StringVar(&messageType, "messageType", "", "type of the request to send")
	flag.StringVar(&payload, "payload", "", "the payload to send to apply")
	flag.StringVar(&source, "source", "", "source value for CRUD operations")
	flag.StringVar(&dest, "dest", "", "CRUD dest value to apply")
	flag.StringVar(&transId, "transId", "", "transId for CRUD operations")
	flag.StringVar(&contentType, "contentType", "", "contentType for CRUD operations")
	flag.UintVar(&maxConnReuse, "max-reuse", 100, "the maximum number of connection re-use attempts -- (advanced)")

	flag.Parse()

	if "" == messageType {
		fmt.Printf("Please provide messageType option\n")
		return
	} else if "" != messageType {
		requestType = "POST"
		if "Create" == messageType || "Update" == messageType {
			if "" == payload {
				fmt.Printf("Please provide payload option\n")
				return
			}
		}

		if "" == source || "" == dest || "" == transId || "" == contentType {
			fmt.Printf("Please provide the required arguments: source, dest, transId, contentType\n")
			return
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: int(maxConnReuse),
		},
	}

	makeRequest(requestType, messageType, source, transId, dest, contentType, payload, client)

}
