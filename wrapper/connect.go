package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/Shopify/sarama"
)

func connect() (sarama.AsyncProducer, error) {
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile("RootCA.crt")
	dsc := "fd0a4b895f-certificate.pem.crt"
	dpk := "fd0a4b895f-private.pem.key"
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	cert, err := tls.LoadX509KeyPair(dsc, dpk)
	check(err)

	tc := tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}

	// https://godoc.org/github.com/Shopify/sarama#ex-AsyncProducer--Goroutines
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tc
	config.ClientID = "ProfileTest"

	return sarama.NewAsyncProducer([]string{"localhost:9092"}, config)
}
