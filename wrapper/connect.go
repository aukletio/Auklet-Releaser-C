package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
)

// https://github.com/eclipse/paho.mqtt.golang/blob/master/cmd/sample/main.go
func connect() (mqtt.Client, error) {
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

	o := mqtt.NewClientOptions()
	o.AddBroker("tcps://a18ej8ow70rah.iot.us-west-2.amazonaws.com:8883")
	o.SetCleanSession(true)
	o.SetClientID("ProfileTest")
	o.SetTLSConfig(&tc)

	c := mqtt.NewClient(o)
	t := c.Connect()
	t.Wait()
	check(t.Error())

	return c, t.Error()
}

func publish(c mqtt.Client, payload []byte) {
	topic := "sdkTest/sub"
	qos := byte(0)
	t := c.Publish(topic, qos, false, payload)
	t.Wait()
	if t.Error() != nil {
		fmt.Println(t.Error())
	}
}
