package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

const (
	apiPrefix = "/api/v2"
)

var _ = FDescribe("Test Custom Root CA For MITM Reencryption", func() {
	It("", func() {
		// generate our "corporate" certificates.
		publicKey, privateKey := generateRootCA()

		fakeLogger := &loggerfakes.FakeLogger{}

		// save them.
		pubKeyFile, err := ioutil.TempFile(".", "test-certs")
		Expect(err).NotTo(HaveOccurred())

		_, err = pubKeyFile.Write(publicKey.Bytes())
		Expect(err).NotTo(HaveOccurred())

		defer os.Remove(pubKeyFile.Name())
		privKeyFile, err := ioutil.TempFile(".", "test-certs")
		Expect(err).NotTo(HaveOccurred())

		_, err = privKeyFile.Write(privateKey.Bytes())
		Expect(err).NotTo(HaveOccurred())

		defer os.Remove(privKeyFile.Name())

		fmt.Printf("publicKey location: %s\n", pubKeyFile.Name())
		fmt.Println(string(publicKey.Bytes()))
		fmt.Printf("privateKey location: %s\n", privKeyFile.Name())
		fmt.Println(string(privateKey.Bytes()))

		// our middle man server.
		testProxyServer := &http.Server{
			Addr: ":8888",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("Connected to proxy server")
				if r.Method == http.MethodConnect {
					handleTunneling(w, r)
				} else {
					handleHTTP(w, r)
				}
			}),
			TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		}
		go testProxyServer.ListenAndServeTLS(pubKeyFile.Name(), privKeyFile.Name())

		// save the current https proxy so we can reset it after the test, set our temp server as the actual proxy.
		currentProxy := os.Getenv("HTTPS_PROXY")
		err = os.Setenv("HTTPS_PROXY", fmt.Sprintf("%s:%s", "https://localhost", testProxyServer.Addr))
		Expect(err).NotTo(HaveOccurred())

		// test release payload can be converted to json.
		releases := pivnet.ReleasesResponse{Releases: []pivnet.Release{
			{
				ID:      1,
				Version: "1234",
			},
			{
				ID:      99,
				Version: "some-other-version",
			},
		}}

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = json.NewEncoder(w).Encode(releases)
			Expect(err).NotTo(HaveOccurred())
		}))

		token := "my-auth-refreshToken"
		userAgent := "pivnet-resource/0.1.0 (some-url)"
		newClientConfig := pivnet.ClientConfig{
			Host:       server.URL,
			Token:      token,
			UserAgent:  userAgent,
			RootCAPath: pubKeyFile.Name(),
			//SkipSSLValidation: true,
		}

		client := pivnet.NewClient(newClientConfig, fakeLogger)

		_, err = client.MakeRequest(
			"GET",
			"/foo",
			http.StatusOK,
			nil,
		)
		Expect(err).NotTo(HaveOccurred())

		err = os.Setenv("HTTPS_PROXY", currentProxy)
		Expect(err).NotTo(HaveOccurred(), "cannot reset HTTPS_PROXY back to %s", currentProxy)

		err = testProxyServer.Shutdown(context.Background())
		Expect(err).NotTo(HaveOccurred())
	})
})

func handleTunneling(w http.ResponseWriter, r *http.Request) {
	destConn, err := net.DialTimeout("tcp", r.Host, 10*time.Second)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(destConn, clientConn)
	go transfer(clientConn, destConn)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
}

func handleHTTP(w http.ResponseWriter, req *http.Request) {
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// generates a new root ca on-demand for the test
// returns: public, private
func generateRootCA() (bytes.Buffer, bytes.Buffer) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization:  []string{"Pivotal"},
			Country:       []string{"US"},
			Province:      []string{"Colorado"},
			Locality:      []string{"Boulder"},
			StreetAddress: []string{""},
			PostalCode:    []string{"80301"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	Expect(err).NotTo(HaveOccurred(), "create ca failed: %s", err)

	// Public key
	var publicKeyOutFile bytes.Buffer
	err = pem.Encode(&publicKeyOutFile, &pem.Block{Type: "CERTIFICATE", Bytes: ca_b})
	Expect(err).NotTo(HaveOccurred())

	// Private key
	var privateKeyOutFile bytes.Buffer
	err = pem.Encode(&privateKeyOutFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	Expect(err).NotTo(HaveOccurred())

	return publicKeyOutFile, privateKeyOutFile
}
