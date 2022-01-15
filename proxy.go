package load_balancer

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Proxy struct {
	Host string
	Port int
	Scheme string
	Servers []Server
}

func (proxy Proxy) origin() string {
	return proxy.Scheme + "://" + proxy.Host + ":" + strconv.Itoa(proxy.Port)
}

func(proxy Proxy) chooseServer(ignoreList []string) *Server {
	skip := false
	min := -1
	minIndex := 0
	for index, server := range proxy.Servers{
		for _, ignoreServer := range ignoreList {
			if server.Name == ignoreServer{
				skip = true
				break
			}
		}

		if skip{
			continue
		}

		conn := server.Connections
		if min == -1{
			min = conn
			minIndex = index
		} else if conn < min{
			min = conn
			minIndex = index
		}
	}

	return &proxy.Servers[minIndex]
}

func(proxy Proxy) reverseProxy(rw http.ResponseWriter, req *http.Request, server *Server) (int, error) {

	u, err := url.Parse(server.Url() + req.RequestURI)
	if err != nil{
		log.Error(err.Error())
	}

	req.URL = u
	req.Header.Set("X-Forwarded-Host", req.Host)
	req.Header.Set("Origin", proxy.origin())
	req.Host = server.Url()
	req.RequestURI = ""

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil{
		log.Error("Connection refused")
		return 0, err
	}

	log.Info("Received response: " + strconv.Itoa(resp.StatusCode))

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		log.Error("Proxy: failed to read response body")
		http.NotFound(rw, req)
		return 0, err
	}

	buffer := bytes.NewBuffer(bodyBytes)
	for k, v := range resp.Header {
		fmt.Println(k,v)
		rw.Header().Set(k, strings.Join(v, ";"))
	}

	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, buffer)
	return resp.StatusCode,nil
}

func (proxy Proxy) attemptServers(rw http.ResponseWriter, req *http.Request, ignoreList []string)  {
	if float64(len(ignoreList)) >= math.Min(float64(3), float64(len(proxy.Servers))) {
		log.Info("Failed to find server..")
		http.NotFound(rw, req)
		return
	}

	server := proxy.chooseServer([]string{})
	log.Info("Got request: " + req.RequestURI)
	log.Info("Sending to server:" + server.Name)

	server.Connections +=1
	_, err := proxy.reverseProxy(rw, req, server)
	server.Connections -=1

	if err != nil && strings.Contains(err.Error(), "Connection refused") {

		log.Info("Server did not respond:" + server.Name)
		proxy.attemptServers(rw, req, append(ignoreList, server.Name))
		return

	}

	log.Info("Responded to request successfully")
}

func(proxy Proxy) Handler(w http.ResponseWriter, r * http.Request)  {
	proxy.attemptServers(w, r, []string{})
}