package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/emojify-app/emojify"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}
var bindAddress = "localhost:9004"
var consulAddress = "consul:8500"
var serviceName = "emojify-emojify"
var cacheService = "emojify-cache"
var faceboxAddress = "http://emojify-facebox:8080"
var imageURL = "https://www.rollingstone.com/wp-content/uploads/2018/06/rs-nirvana-e9e22e4b-f7d9-4fc7-bd94-23c30084ce94.jpg"

var server *emojify.Server

var response string
var responseImage []byte

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func theServerIsRunning() error {
	server = emojify.NewServer(bindAddress, cacheService, consulAddress, faceboxAddress, serviceName, "../images/")
	go server.Start()
	time.Sleep(5 * time.Second)

	return nil
}

func iPostAnImageURLToTheEndpoint() error {
	resp, err := http.Post("http://"+bindAddress, "", bytes.NewReader([]byte(imageURL)))
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	response = string(data)

	return nil
}

func iExpectABase64URLToBeReturned() error {
	_, err := base64.StdEncoding.DecodeString(response)
	if err != nil {
		return err
	}
	return nil
}

func iCallGetWithTheImageURL() error {
	resp, err := http.Get("http://" + bindAddress + "/" + response)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	responseImage = data

	return nil
}

func iExpectAValidImageToBeReturned() error {
	contentType := http.DetectContentType(responseImage)

	if contentType != "image/png" {
		return fmt.Errorf("Expected a valid png to be returned, got: %s", contentType)
	}

	return nil
}

func FeatureContext(s *godog.Suite) {
	s.AfterScenario(func(interface{}, error) {
		server.Stop()
	})

	s.Step(`^the server is running$`, theServerIsRunning)
	s.Step(`^i post an image url to the endpoint$`, iPostAnImageURLToTheEndpoint)
	s.Step(`^i expect a base64 url to be returned$`, iExpectABase64URLToBeReturned)
	s.Step(`^i call get with the image url$`, iCallGetWithTheImageURL)
	s.Step(`^i expect a valid image to be returned$`, iExpectAValidImageToBeReturned)
}
