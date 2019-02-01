package pivnet

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
	"github.com/pivotal-cf/go-pivnet/logger/loggerfakes"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const (
	apiPrefix = "/api/v2"
)

func TestCustomRootCAForMITMReencryption(t *testing.T) {

	// generate our "corporate" certificates.
	publicKey, privateKey := generateRootCA(t)

	fakeLogger := &loggerfakes.FakeLogger{}

	// save them.
	pubKeyFile, err := ioutil.TempFile(".", "test-certs")
	if err != nil {
		t.Error(err)
	}
	if _, err := pubKeyFile.Write(publicKey.Bytes()); err != nil {
		t.Error(err)
	}
	defer os.Remove(pubKeyFile.Name())
	privKeyFile, err := ioutil.TempFile(".", "test-certs")
	if err != nil {
		t.Error(err)
	}
	if _, err := privKeyFile.Write(privateKey.Bytes()); err != nil {
		t.Error(err)
	}
	defer os.Remove(privKeyFile.Name())

	// our middle man server.
	testProxyServer := &http.Server{
		Addr: ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if err := os.Setenv("HTTPS_PROXY", fmt.Sprintf("%s:%s", "https://localhost", testProxyServer.Addr)); err != nil {
		t.Error(err)
	}

	// test release payload.
	releases := ReleasesResponse{Releases: []Release{
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
		if err := json.NewEncoder(w).Encode(releases); err != nil {
			t.Error(err)
		}
	}))

	token := "my-auth-refreshToken"
	userAgent := "pivnet-resource/0.1.0 (some-url)"
	newClientConfig := ClientConfig{
		Host:       server.URL,
		Token:      token,
		UserAgent:  userAgent,
		RootCAPath: pubKeyFile.Name(),
	}

	client := NewClient(newClientConfig, fakeLogger)

	_, err = client.MakeRequest(
		"GET",
		"/foo",
		http.StatusOK,
		nil,
	)
	if err != nil {
		t.Error(err)
	}

	if err := os.Setenv("HTTPS_PROXY", currentProxy); err != nil {
		t.Errorf("cannot reset HTTPS_PROXY back to %s", currentProxy)
	}

	if err := testProxyServer.Shutdown(context.Background()); err != nil {
		t.Error(err)
	}
}

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
func generateRootCA(t *testing.T) (bytes.Buffer, bytes.Buffer) {
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
	if err != nil {
		t.Errorf("create ca failed: %s", err)
	}

	// Public key
	var publicKeyOutFile bytes.Buffer
	if err := pem.Encode(&publicKeyOutFile, &pem.Block{Type: "CERTIFICATE", Bytes: ca_b}); err != nil {
		t.Error(err)
	}

	// Private key
	var privateKeyOutFile bytes.Buffer
	if err := pem.Encode(&privateKeyOutFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		t.Error(err)
	}

	return publicKeyOutFile, privateKeyOutFile
}
