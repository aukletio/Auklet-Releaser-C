package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/Shopify/sarama"
)

func connect() (sarama.SyncProducer, error) {
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile("ck_ca")
	dsc := "ck_cert"
	dpk := "ck_private_key"
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

	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = &tc
	config.ClientID = "ProfileTest"

	broker := map[string][]string{
		// new
		"steamer": {
			"steamer-01.srvs.cloudkafka.com:9093",
			"steamer-02.srvs.cloudkafka.com:9093",
			"steamer-03.srvs.cloudkafka.com:9093",
		},
		// new/qa
		"dogsled": {
			"dogsled-01.srvs.cloudkafka.com:9093",
			"dogsled-02.srvs.cloudkafka.com:9093",
			"dogsled-03.srvs.cloudkafka.com:9093",
		},
		// prod
		"prod": {
			"striped-water-wagon-01.srvs.cloudkafka.com:9093",
			"striped-water-wagon-02.srvs.cloudkafka.com:9093",
			"striped-water-wagon-03.srvs.cloudkafka.com:9093",
		},
	}
	return sarama.NewSyncProducer(broker["prod"], config)
}
