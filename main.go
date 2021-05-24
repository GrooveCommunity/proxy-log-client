package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func main() {
	help := flag.Bool("help", false, "Opcional, Imprime informação para o usuário")
	srvhost := flag.String("srvhost", "", "Endereço do servidor")
	caCertFile := flag.String("cacert", "", "Requerido, CA que assinou o certificado do servidor.")
	clientCertFile := flag.String("clientcert", "", "Requerido, Certificado do cliente")
	clientKeyFile := flag.String("clientkey", "", "Requerido, Chave privada do cliente")

	flag.Parse()

	usage := `usage:
	
proxylogclient -host -clientcert <clientCertificateFile> -cacert <cafile> -clientkey <clientPrivateKeyFile> [ -help]

Options: 
  -help			Opcional, Imprime essa mensagem
  -srvhost		Requerido, Endereço do servidor
  -clientcert	Requerido, Chave do cliente
  -clientkey		Requerido, Certificado do cliente
  -cacert		Requerido, O nome do CA que assina o certificado do servidor`

	if *help == true {
		fmt.Println(usage)
		return
	}

	if *caCertFile == "" || *srvhost == "" || *clientCertFile == "" || *clientKeyFile == "" {
		log.Fatalf("Um ou mais campos requeridos estão faltando: \n%s", usage)
	}

	var cert tls.Certificate
	var err error

	cert, err = tls.LoadX509KeyPair(*clientCertFile, *clientKeyFile)
	if err != nil {
		log.Fatalf("Error creating x509 keypair from client cert file %s and client key file %s", *clientCertFile, *clientKeyFile)
	}

	log.Printf("CAFile: %s", *caCertFile)

	caCert, err := ioutil.ReadFile(*caCertFile)
	if err != nil {
		log.Fatalf("Error opening cert file %s, Error: %s", *caCertFile, err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	t := &http.Transport{
		TLSClientConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		},
	}

	client := http.Client{Transport: t, Timeout: 15 * time.Second}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("https://%s", *srvhost), bytes.NewBuffer([]byte("World")))

	if err != nil {
		log.Fatalf("unable to create http request due to error %s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		switch e := err.(type) {
		case *url.Error:
			log.Fatalf("url.Error received on http request: %s", e)
		default:
			log.Fatalf("Unexpected error received: %s", err)
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatalf("unexpected error reading response body: %s", err)
	}

	fmt.Printf("\nResponse from server: \n\tHTTP status: %s\n\tBody: %s\n", resp.Status, body)
}
