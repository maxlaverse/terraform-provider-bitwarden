package embedded

import (
	"encoding/json"
	"log"
	"os"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/models"
	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/webapi"
)

const (
	ServerURL = "http://127.0.0.1:8080"

	RsaPrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAlRbtt5Vyku3dUIkPR0Y3v94qvZReuwAijIwHQGQOuKw6lKVV
HL29rZ93TwCG2P5a+GKH2+2fIbT/wTMTK4K1ElxQZ/2yLN9Hfu2d/ITNTsfTzPXv
F3fao3Q0JmD7DNJS2bJqng3so28aFddOZ03H55m9T6+0ZJqrMdgE5Z1V/4I7LFF6
JxGGZ4mg99OvfQ5K8GOBM6SCI6h5eMXM5EkSM9vol9sRxvVLZmvNKH3UP9riyQwt
dcD5IxmY1y34Bg4b+a8tYaP7v90xF73uKs+287yNPLhWE9i+/gXwvApVxENG9SCP
lSrAEHd1NZPfsZhHoG+LXhyCZu2COttZess44QIDAQABAoIBAAIXgN54+qMpJ+2M
yGsdvGj3vCy9+vSyWi8Tr3icdXMrKfTVgMUlvEurcOI/Mcp+v55MF3JF0kwylh3N
pSwbV3DBHN5Hp5xu8HmtahsoWRnXo98Z4oOB7U5gAj5kBmkMtKhB/fJW+UzF/C/b
I2906Tw59Uy2XIsROzvjMeGPnddh1LbvXUb9nAmhi7napdwCUbeqvatu56GyiXxV
03DTwhbfuU+nMi/M556WPEkPbJoG4bF82WqQ+6+a1NfE2wg//cc5CzXQehC6jGx6
Pi+uNUtPhMSyTnJgpvg+Ob2r/LTiL0zic00ka29xUsi7EwXKUR2ih8JTkSPaTR06
3ezrg3ECgYEAzi/MnXfWp49jRgb4bHE9Cy63hehgaBvIcEyhFKOz5OD6nI4Lg/3z
SNDQo9YhwMqschqQLHVEtjxDT0V/RHdX4icTF+zSCl/T79EtM/R1nMT0MSIXV7IE
NtPbnqXOjrbe3vgjLvBst/cWpGHiML+znCqukHOevSn7yUlg4b1aMVECgYEAuRvL
YnbNlps/nql1EW9DUKOEV7kBvehwGzYpFfZ7pRscl6RyISTipGMOzJmOSfscYwqR
HrFpTNMNxjnyXOuv4OTC7bCIUJc6N8AZ4jm4452ibxpzktlO8Im+TbsD6mZDT8zB
d+8o/8ST9j1zy43Sb+f5vxfB6fC9vUXpBW0+KpECgYEAzQKT9cJxSVv1/mvx2Ilj
g9nompmqOfnd+2MGCuqWdS4JoV5PLudzXeRaf3zrRLGAc1fcIIhdUMFsv8Y/O8la
NcBaaMCNO8l6hoo64tzfkIf4sV3PTd/v9sACL6V3U0mbIqIhAYwG3YguGDZHW+dQ
ZCfAOFrt6/Jxqvtt/CZ1JnECgYBOxmdNZeWj/Dmc2dy6KLFq9ctyUYdOPEbJLcla
UWTZJKqMVi1DsaDJ+GXp6EdHcJfqBisv9qwrR34LJ8nehWZ5vKC/6mp4cYMTCqt5
PLtUEld4FLeufNA9SUE1bysBa7ellCuZUKwP/KZDGm/W5mnxubTs/71EQ3FbxQ6f
gpf8IQKBgGHK8j8rvxtszoQKUY+XpWFrP0x5pDLiAkmQ0bmF2KRIahq3anla6n3i
/LL5BrUdMjEnvAb+RASq+41rceq4rLcz0pA2yOWNjhbCAPFdU5MQMkJ4/zqHtzvd
GqwE00g9gizQ6CmsaNNJh7y6gNg0TBU2EGqTaQMz37fheAEt3NSt
-----END RSA PRIVATE KEY-----
`

	EncryptionKey = "Vr+KA/il3QX4z7EqFnhQ3U8TtETlQPKkXHCE2PiR75wwzDVRutR4rib/jMtgZ1S/gPyOEXbwKFju2oJq3njVLg=="
	TestPassword  = "test12341234"

	Pdkdf2Mocks = "test-pbkdf"
	Argon2Mocks = "test-argon2"
)

var (
	OrganizationID string
	AccountPbkdf2  Account
	AccountArgon2  Account
)

func init() {
	data, err := os.ReadFile("fixtures/test-pbkdf_GET_api_sync.json")
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return
	}

	var resp webapi.SyncResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
		return
	}

	OrganizationID = resp.Profile.Organizations[0].Id
	AccountPbkdf2 = Account{
		AccountUUID: resp.Profile.Id,
		Email:       resp.Profile.Email,
		VaultFormat: "API",
		KdfConfig: models.KdfConfiguration{
			KdfType:       models.KdfTypePBKDF2_SHA256,
			KdfIterations: 1000,
		},
		ProtectedSymmetricKey:  resp.Profile.Key,
		ProtectedRSAPrivateKey: resp.Profile.PrivateKey,
	}

	data, err = os.ReadFile("fixtures/test-argon2_GET_api_sync.json")
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return
	}

	err = json.Unmarshal(data, &resp)
	if err != nil {
		log.Printf("Error unmarshalling data: %v", err)
		return
	}

	AccountArgon2 = Account{
		AccountUUID: resp.Profile.Id,
		Email:       resp.Profile.Email,
		VaultFormat: "API",
		KdfConfig: models.KdfConfiguration{
			KdfType:        models.KdfTypeArgon2,
			KdfIterations:  3,
			KdfMemory:      64,
			KdfParallelism: 4,
		},
		ProtectedSymmetricKey:  resp.Profile.Key,
		ProtectedRSAPrivateKey: resp.Profile.PrivateKey,
	}
}
