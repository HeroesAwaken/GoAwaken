package main

import (
	"flag"
	"os"
	"os/signal"

	"net/http"
	_ "net/http/pprof"

	gs "github.com/ReviveNetwork/GoRevive/GameSpy"
	log "github.com/ReviveNetwork/GoRevive/Log"
	"github.com/kabukky/httpscerts"
)

var (
	// BuildTime of the build provided by the build command
	BuildTime = "Not provided"
	// GitHash of build provided by the build command
	GitHash = "Not provided"
	// GitBranch of the build provided by the build command
	GitBranch = "Not provided"
	// compileVersion we are receiving by the build command
	CompileVersion = "0"
	// Version of the Application
	Version = "0.0.0"
)

func CheckAndGenerateHTTPSCertificate(certFile, keyFile string) {
	// If we have no cert and key file, create one
	err := httpscerts.Check(certFile, keyFile)
	if err != nil {
		log.Noteln("Creating a new certificate for 127.0.0.1")
		err = httpscerts.Generate(certFile, keyFile, "127.0.0.1")
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}
}

func main() {
	var (
		logLevel     = flag.String("logLevel", "error", "LogLevel [error|warning|note|debug]")
		certFileFlag = flag.String("cert", "cert.pem", "[HTTPS] Location of your certification file. Env: LOUIS_HTTPS_CERT")
		keyFileFlag  = flag.String("key", "key.pem", "[HTTPS] Location of your private key file. Env: LOUIS_HTTPS_KEY")
	)
	flag.Parse()

	go func() {
		log.Noteln(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	if CompileVersion != "0" {
		Version = Version + "." + CompileVersion
	}

	log.SetLevel(*logLevel)
	log.Notef("Starting up v%s", Version)

	// Startup done

	// Generate session key

	CheckAndGenerateHTTPSCertificate(*certFileFlag, *keyFileFlag)

	test3 := new(gs.Socket)
	eventsChannel, err := test3.New("Testing", "42127", false)
	if err != nil {
		log.Errorln(err)
	}

	for {
		select {
		case event := <-eventsChannel:
			switch {
			case event.Name == "newClient":
				log.Debugln(event)
			case event.Name == "error":
				log.Debugln(event)
			case event.Name == "close":
				log.Debugln(event)
				os.Exit(0)
			default:
				log.Debugln(event)
			}
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for sig := range c {
		log.Noteln("Captured" + sig.String() + ". Shutting down.")
		os.Exit(0)
	}

}
