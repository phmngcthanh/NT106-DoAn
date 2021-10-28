package main

import (
	"Tiamat/log"
	"Tiamat/prisma"
	"Tiamat/process"
	"crypto/tls"
	"fmt"
)

/* hard-coded certs*/


const serverKey = `-----BEGIN RSA PRIVATE KEY-----
REDACTED
-----END RSA PRIVATE KEY-----`

const serverCert = `-----BEGIN CERTIFICATE-----
MIIHDDCCBfSgAwIBAgIQDAHyfOtQgf/6F2sFNGRwdzANBgkqhkiG9w0BAQsFADBu
MQswCQYDVQQGEwJVUzEVMBMGA1UEChMMRGlnaUNlcnQgSW5jMRkwFwYDVQQLExB3
d3cuZGlnaWNlcnQuY29tMS0wKwYDVQQDEyRFbmNyeXB0aW9uIEV2ZXJ5d2hlcmUg
RFYgVExTIENBIC0gRzEwHhcNMjEwNTI3MDAwMDAwWhcNMjIwNTI3MjM1OTU5WjAc
MRowGAYDVQQDExFwaG1uZ2N0aGFuaC5jb2RlczCCAiIwDQYJKoZIhvcNAQEBBQAD
ggIPADCCAgoCggIBAMhsB5OtBw2Wj+So6g3L6teSN3EOeq5yTkfZ7h0IXge9hhvZ
2sLfWr8fteE57WQzT+gt6ukyLAXK9krGMzzKiZS5bejbZARhhErnURwn5EqiNfIp
boPlajj9r61w24skr6lxV6Umv1HmlsVqINKtKgldsAuYa2EODJ3docBFhjljrmGw
+N+XQgFI/Tqjm+OGn6S2o3eGR53aqcO4kyU0LDA9sD9q+IG7lifFvJiPcTwjN/wG
LGTgEH1WoZH204PqRoCGhiLPMvHSzZ8fw/BVl1vXwUMinVKj9hIn8MR6xHaCn7cK
gqMFgdOOO32Y4By4R4NagODiXtPTr3u2uQS3a+GDTW8myQVC/MYevZjyD99PM2dn
JOsSE6CKNn31TWlocyzuwSbkBaa6kt7LRX8mdZm2YW2ry5e8IaRYPHL8ykcmfNRX
6UgjGcHCRhRbegAxq3ofOqVPW+SyB3+e1TzZCJPfjjjuDqXOn2xmxg5BvoZ7NUeh
pjPgs8ak2d5w50mfnCeAT1Q71o6rfwB/1rGQAzImJKfbpcYoToy6DXMZOTGIMzKz
vaSm7orBBpy28KuWqz3lG/W0Mss/GgFDF4nj2eP1TPy5fHRCwSYDYUw+sgh1h8rd
nhWfsmZhw9ZM/IWVKHQAeLxjrZxczgr1PhxPxJaAVQzOnh/+VIYPiUzuCu8BAgMB
AAGjggL2MIIC8jAfBgNVHSMEGDAWgBRVdE+yck/1YLpQ0dfmUVyaAYca1zAdBgNV
HQ4EFgQUM+MOWJMF/p8xMNR9uoO4G2KRqUkwMwYDVR0RBCwwKoIRcGhtbmdjdGhh
bmguY29kZXOCFXd3dy5waG1uZ2N0aGFuaC5jb2RlczAOBgNVHQ8BAf8EBAMCBaAw
HQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMCMD4GA1UdIAQ3MDUwMwYGZ4EM
AQIBMCkwJwYIKwYBBQUHAgEWG2h0dHA6Ly93d3cuZGlnaWNlcnQuY29tL0NQUzCB
gAYIKwYBBQUHAQEEdDByMCQGCCsGAQUFBzABhhhodHRwOi8vb2NzcC5kaWdpY2Vy
dC5jb20wSgYIKwYBBQUHMAKGPmh0dHA6Ly9jYWNlcnRzLmRpZ2ljZXJ0LmNvbS9F
bmNyeXB0aW9uRXZlcnl3aGVyZURWVExTQ0EtRzEuY3J0MAkGA1UdEwQCMAAwggF8
BgorBgEEAdZ5AgQCBIIBbASCAWgBZgB2ACl5vvCeOTkh8FZzn2Old+W+V32cYAr4
+U1dJlwlXceEAAABea0UyKYAAAQDAEcwRQIhAOZjyIXWL3FgMdSIXsPGut21QCSx
936viBkAeXGwsaE2AiB3CLx5uW873a+oGZcTgmlm9Ht0DfrrSKpY9Xc7IYEvswB1
ACJFRQdZVSRWlj+hL/H3bYbgIyZjrcBLf13Gg1xu4g8CAAABea0UyPkAAAQDAEYw
RAIgDTBoNLM7In6WZLPXzA2dA91dPZkbQXh29YpHwnNBOAMCIDbTKK/9wEdFDJND
OnE30XsWfLQDpXgFiu6hz7SmjPfDAHUAUaOw9f0BeZxWbbg3eI8MpHrMGyfL956I
QpoN/tSLBeUAAAF5rRTJNQAABAMARjBEAiAs/NKULw3lpKLuga4gd6qJuvkxtM0d
zwObSnErYIF6OgIgWUz3uAUik+eis20eb+3N+48x2CI5HeL5juLsN0O5lMkwDQYJ
KoZIhvcNAQELBQADggEBAIVt1nlskt3dyTr4xNGUMglL4pBQVhZdJ6sntco/9HER
DBd2Ee6qoXYIxQQJPpmU+ldQuhJZckvcChmw6WbVunnF5JaXLnHesr/qrP5QoTMF
s9uCcR80VBHFus3jt+OAEL4GkKpjE5O2/iN4faCZ3HY3xC+J7PY04Z+Z8P2498fq
yLPfFC0VAbZNr85idLLWWLcVBPpY633uOZXo6RoyCKqanpMXdpFCkS0cFVZ3bx9g
fdG+31KzqD/LgLc2Pfvx1OHXuhaBXNLCW6ajEQRFgygez3uvWy+TPQQvmUyOA/oU
I78diIqTRKy+XCSTqVWVaRoa0spcpQ+fcfngB6dPRrQ=
-----END CERTIFICATE-----
`


func main() {
	prisma.InitPrisma()
	process.InitSocket()
    
    /* trying to implemenmt load-balancing
	fmt.Println("Setup GRPC addr:")
	s := ""
	_, _ = fmt.Scanf("%s", &s)
	go process.SocketServer(s)
	time.Sleep(100 * time.Millisecond)
	n := 0
	fmt.Println("Enter the number server")
	_, _ = fmt.Scanf("%d", &n)
	fmt.Println("Enter other server's addr")
	for i := 0; i < n; i++ {
		fmt.Println("Server", i+1, ": ")
		s := ""
		_, _ = fmt.Scanf("%s", &s)
		process.ServerFriend[s] = s

	}
    
	process.BroadCastServer("1", "", "", s)
	time.Sleep(100 * time.Millisecond)
*/
	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		fmt.Println(err)
	}
    AcceptedCipherSuite := []uint16{
    tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
    tls.TLS_RSA_WITH_AES_128_CBC_SHA,
    tls.TLS_RSA_WITH_AES_256_CBC_SHA,
    tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
    tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
    }
	
    config := &tls.Config{Certificates: []tls.Certificate{cer},CipherSuites: AcceptedCipherSuite}

	addr := "0.0.0.0:17749"
	listener, _ := tls.Listen("tcp", addr, config)
	fmt.Println("Running on " + addr + "/tcp ........")
	for { /*while true*/
		conn, err := listener.Accept()
		if err != nil {
			log.Log.Error(err.Error())
			continue
		}

		log.Log.Println(conn.RemoteAddr().String() + " connect .....")
		go process.AcceptConnect(conn)
	}
}
