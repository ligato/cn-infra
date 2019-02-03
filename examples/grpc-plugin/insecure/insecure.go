//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package insecure

import (
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"log"
)

const certPEM = `-----BEGIN CERTIFICATE-----
MIIFDTCCAvWgAwIBAgIRAPkwJic3oOBPq4nxQ8j9la0wDQYJKoZIhvcNAQELBQAw
EjEQMA4GA1UEChMHQWNtZSBDbzAeFw0xODA3MjIxNjMzNTFaFw0yODA3MTkxNjMz
NTFaMBIxEDAOBgNVBAoTB0FjbWUgQ28wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAw
ggIKAoICAQDfpqmkHYuRO35ZEXI2ouaWGc2wpI6nP6Rin2ygfgWbzKkmL31d68+P
5jEbOUj9eJ5jZeCggGr8LvIx+61G1Tp0AoabVHnMZA7xCZ3Ri1mPfFGGNAziWI0T
ltXcm8KD6nwvYpgf35FS9vbiY9jXPqT0DWJBaJ8uNlTE1BUmLyiD9oy6jOx5v40k
0C+cBufeLsHstpG1q03zZ+o3Q0pfhYGWHBpnv1Nes2EeZxVrg0811F44Ii4UQyEk
ie45wz5uZzj/3Mkpo9Bd5VjUdjIyY0bJZarJHe8JBWjHysE4Im7X4NGYvhTBtvMZ
wk9GZ2v5KWICH5S0b/bj3dYhE97wRScMkDpdgeVXbuOMy02gHWqXc+uPZqpozH6D
xIz52hW1AwsijlDvzB7xG88Uf0RnA7ZJtUhfOFFg7dn8g2SeZ+4TB3jrJfzgkUES
qsyM05HRV4P3EIdHUtShBTs7hnzYdn0DYRJ8v243XMC2D5Qd3y2P+FAG9LK/25no
3GeNWWpy9k04AY4bRZTDi2byGcT4qARp9egEpEs71EFQkpkbUapGR1Zx5yf+4Iyr
ETmmiC2m66DHtFtZ8+4U26er4eAzyN6ZoZ+TM/3UZ2eF42glt41JH/OgkQS0Vaql
4K6s0T74/j4HohbA2l0TodvdPqWsZmBTXYaKdaCSlG8BOS46i3olzwIDAQABo14w
XDAOBgNVHQ8BAf8EBAMCAqQwHQYDVR0lBBYwFAYIKwYBBQUHAwEGCCsGAQUFBwMC
MA8GA1UdEwEB/wQFMAMBAf8wGgYDVR0RBBMwEYIJbG9jYWxob3N0hwR/AAABMA0G
CSqGSIb3DQEBCwUAA4ICAQCfJ69SXjDZ6qmSZPba9//oUmks9GhNn6XJvhxKTuel
O0eyMzBe1BBwoVMDQc43oquSBh0w5oN6rozjxFN48QfGrO3Ba/U7JWY9R1WD8ftd
FPveR6IjWSkj0TBn3u9WGzfOvIWRsbtC505sx8kesA5jOjJ1dfY5knY2S/9Zr2Nn
SqXuaIRhhMUyrXrST6sedrYyKU8xPS2dwkhALiwSEwSRWlErIZ1ZBsSrMSIIsKmr
X0TNoF+B4Z4Tgx2JO3aZlz9DYn6CXSkxwMcIqi8eScq8mYKjXd1Q3YttVnWmxKXi
Nr8nE7sJ9nF1CXk0AXSpP5/0KiNqs5UUKK+CWFUECyEAjy/tOH/ZRkmfDulLCJyd
cZVOwpd2CMlXvfEwjHrjhxnx6wQNpbB4ztQUgnQ2iTVHOCNGbw8M/AQFiuG++xM1
r3nPXWZSmTsoCbmCqxh2fN7YB4aDx9oenuQUmZd1OxvL4zO0kgSwIVifRiLo7jWe
3LHGATJeQDgX26dJ7Kplqe1h15JAL2EvjGWrZxpgaKFffoQfZm6j/M2i+1ab7zVT
VOEkJn5s43GClrG2QSez3RtjfrIuQURcutMz81+kfpOz9sJuvlZyt0fxKlvgRNdg
ZHUZe+1kPi32wi9RPFsNBqs1RhaIaczwJrvtz7giPUXvhhrQuQPLE0ImPqbAY/49
Rg==
-----END CERTIFICATE-----
`

const keyPEM = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

var (
	// Key is the private key
	Key crypto.PrivateKey
	// Cert is a self signed certificate
	Cert tls.Certificate
	// CertPool contains the self signed certificate
	CertPool *x509.CertPool
)

func init() {
	var err error
	Cert, err = tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		log.Fatalln("Failed to parse key pair:", err)
	}
	Cert.Leaf, err = x509.ParseCertificate(Cert.Certificate[0])
	if err != nil {
		log.Fatalln("Failed to parse certificate:", err)
	}

	CertPool = x509.NewCertPool()
	CertPool.AddCert(Cert.Leaf)

	Key = Cert.PrivateKey
}
