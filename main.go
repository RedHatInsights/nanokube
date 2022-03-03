package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"

	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

type ManglerConfig struct {
	TLS rest.TLSClientConfig
}

type Mangler struct {
	URL    *url.URL
	Config *rest.Config
}

func NewMangler(cfg *rest.Config) *Mangler {
	m := &Mangler{
		Config: cfg,
	}
	return m
}

func (m *Mangler) modifier(request *http.Request) {
	fmt.Printf("\n\n%v\n\n", request)
	murl, _ := url.Parse(m.Config.Host)
	request.URL.Host = murl.Host
	request.URL.Scheme = murl.Scheme

	request.Host = murl.Host
	fmt.Printf("\n\n%v\n\n", request)
}

func main() {
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("static"), // added to the project manually
		},
	}

	fmt.Print("Starting server....\n")

	cfg, err := testEnv.Start()

	if err != nil {
		fmt.Print(err)
		os.Exit(127)
	}

	fmt.Printf("KubeAPI started!\n")

	k8sClient, _ := client.New(cfg, client.Options{Scheme: clientgoscheme.Scheme})

	ctx, _ := context.WithCancel(context.Background())

	nsSpec := &core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "kafka"}}
	err = k8sClient.Create(ctx, nsSpec)

	if err != nil {
		fmt.Print(err)
		os.Exit(127)
	}

	fmt.Printf("Started... ")
	fmt.Printf("%v", testEnv.Config.Host)

	os.WriteFile("/tmp/cert.cert", cfg.TLSClientConfig.CertData, 0664)
	os.WriteFile("/tmp/ca.cert", cfg.TLSClientConfig.CAData, 0664)
	os.WriteFile("/tmp/key.key", cfg.TLSClientConfig.KeyData, 0664)

	mangler := NewMangler(
		cfg,
	)

	cert, err := tls.X509KeyPair(cfg.TLSClientConfig.CertData, cfg.TLSClientConfig.KeyData)
	if err != nil {
		fmt.Printf("BADNESS CERT")
		os.Exit(127)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cfg.TLSClientConfig.CAData)

	proxy := httputil.ReverseProxy{
		Director: mangler.modifier,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}
	server := http.Server{
		Addr:    "0.0.0.0:8090",
		Handler: &proxy,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		fmt.Printf("\nloop")
		for sig := range c {
			fmt.Printf("\n%v", sig)
			fmt.Println("\nQuitting...")
			testEnv.Stop()
			server.Shutdown(context.Background())
			os.Exit(0)
		}
	}()

	server.ListenAndServe()

}
