package crypto

import (
	"crypto/x509"
	"encoding/base64"
	"testing"

	"github.com/maxlaverse/terraform-provider-bitwarden/internal/bitwarden/crypto/symmetrickey"
	"github.com/stretchr/testify/assert"
)

const (
	testEncryptedEncryptionKey = "2.A/89QpHf5lcmfJhoEHKbLw==|gUk+VUVzeaArKHGTz4O8ds0EoIiugPnTiZ59li6uy/k7lnAhantnrZtx6xVrEeNWjzaPuoWVUJ5rycESsLPYqn1eUak5OWO43tdFvD0beic=|CvPdfaaIbbsBr3Tnius/n69Rg60v/AoBQTfecWgFsv4="
	testEncryptedPrivateKey    = "2.A/89QpHf5lcmfJhoEHKbLw==|AZKMjtMjymF1xd3T9dJpctyq9b4ANPc3um6XPQY2CPWltR1veGav/HFGZRsTsK3gBz8jVvvR7rISwhs2vXyN5bJwS/TzKJGh22rqgIvqQYd+OQ7Bnpc9TOyEdO2JmAt3Y2zklHkRgYEcdZKgFTYQAa59wiqZBnduZgGoeEyyK/+q+0blbgT40gu9Xpu9LgqhK2ZrKMfM+o1ckhmanLVpqtNsAKTFWc3o87SY9WR3csO+9UexTJ8WKsdBL0VFILsdjcl2LaGtJrg4/dizGf4blkrSaTVgZPNy4FvJd8LH/wnyaqARJQHNwCyxId4o1ZLovhBH6Sm6tGTLLU5FW5pCR16smt0ZTlpKckj0rQ4AIG81nVZyfri275wiov5+64uGoNu32xy8GOZVlTT9YxLB52t8OS3012boI3DT2PQeU8YfzkkmHUCOCNeG0F2nTptbV59k1TvAJ3FwI+5Zsx25ygGzu/FXWpBmMjQyrb2n9B4ftF3gzYaCKnAtf/efOJVu6WmQ8dl27peCurkQpFiTG7p0FIvRmnVyRj3DKPG+uonB9Z07jkg7dLHOVFCFYAz5VwBfvPXaES6/DH3V/biBq7zt6yvy0B2CRZcUM22Q8xuboeUARRiy+dsmndSxaQmIoreuzWfPAcQ0/bxGzvL/8oVzpUu3YfzNrRjWC2ptQPRCHC94CId4w/MDOMvJS4w1BgMcDZLyWbjyxAmb69Bn+NsVGoSeCR9eXiHk8tZFUReS1s6Zixhoz0z/oVDacTaHEP1qEcFZiewfPdiVSXOPlgoY/OdVfBOJQPjdKR73EsbOpagc7a+4GTRLmFoNNdX5g5t0pS9FL+9FAoUXyu0qCBioaRw5N8TSA3yQ3smTD2t4237mqfKK0yGQ/Uq8WNJmyC1HYgZMq54Y4pu5XgA9nPM5AdCBjN+JRIiaWsFH56nyKiR93zLAhffUxB4AKySr8hpawIYoga/v1YsGAFV7rDI05UePYiuybiTajseJBKECCsgnFaUdNSsJP4ZAlHpl/IUQ1f80+0A+iFGeGJ2U0Gx6KFupqnWnc5nlKhU4wMZLqi+vMJUnGvNAUwghjxIYvFOg44yEXRh4q07NpmLFb3xLKmLB6/SigjdO7I7oVD4nW/yqfrdOdJdgyuzEQO5/6pkNBKX0Ak7M9Te6Zk8F8hAFd5nL5PK2vodTdr4qxZuSFbF5cK+FrkiWbQzR58c71yUyR+QA2yoN2iKtmgdQXIPvA2ohRNqJD2VG5CUS465qHNkSYkHec411r0+ORpAU3z7q0/iH7UBghP8EEGZKC0VI+cwt4NyRb6NIhSfqs71e/inqltq5onq1nEfb8Akt/ZhpxdIKwLom5raiVNXAM8CG0Jkoua7IqE+dsoUxBxBboU3A0vXo7EtLkBTfvj55h1//kWeq4NeWJM59oJswWbKoWWosLZT0xh122ysfC5B/zhpum8jKTUF2No1tlWnhB5I6pBywo/rXZ+T4owqJS4kpKm8l5vMGcWOM4FVXBPlkf0r0A8wyp3/XZfZLOoZbbNjtrGJIBbVtRDglka/anOUtnVxHPXZPHlvOWm2Ql9J+YS3fubINXVfjg2HbJKNnk3qT53fMfrc4jnMvJ04ggu91Rdu+9PN3iBkexMMgzaY=|rN+kIkJLhAkHozjGS8g/jKB02unQ7ZNreM37Px2QPnU="
)

var (
	testEncryptionKey = []byte{40, 201, 91, 4, 62, 57, 230, 98, 146, 113, 111, 129, 180, 230, 116, 91, 110, 163, 34, 47, 127, 131, 59, 252, 7, 101, 153, 48, 185, 209, 19, 45, 227, 232, 133, 165, 156, 157, 9, 202, 36, 235, 96, 151, 31, 27, 38, 238, 213, 219, 189, 229, 182, 208, 39, 208, 53, 69, 204, 22, 157, 76, 151, 209}
	testPreloginKey   = []byte{132, 18, 136, 110, 143, 171, 79, 198, 50, 219, 96, 87, 26, 192, 36, 108, 52, 152, 189, 158, 177, 251, 131, 64, 138, 139, 216, 149, 251, 177, 200, 131}
)

func TestDecryptEncryptionKey(t *testing.T) {
	preloginKey, err := symmetrickey.NewFromRawBytes(testPreloginKey)
	assert.NoError(t, err)

	encryptionKey, err := DecryptEncryptionKey(testEncryptedEncryptionKey, *preloginKey)
	assert.NoError(t, err)
	assert.Equal(t, testEncryptionKey, encryptionKey.Key)
}

func TestDecryptPrivateKey(t *testing.T) {
	encryptionKey, err := symmetrickey.NewFromRawBytes(testEncryptionKey)
	assert.NoError(t, err)

	privateKey, err := DecryptPrivateKey(testEncryptedPrivateKey, *encryptionKey)
	assert.NoError(t, err)

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	assert.NoError(t, err)

	base64PublicKey := base64.StdEncoding.EncodeToString(publicKeyBytes)
	assert.Equal(t, "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzfZx4rRpKBVnhiqZe5IH5mRvHjY1iTrZOpooma8PtOIoIdtSRY5YdeX4Hben09C8jZODgyPtxVbWZv/YBS9okE6gPsqugDMQ5M+t7hp3ye9art7CkfvIDjGHZMrANQCYB/tPWkda7jaaAIBkCIPM4+vZ7afBN3Mq/BX7hotSaGlPPP7DCkzbKK/f5U/F/dA8UTZFXtST9ivRWWI8bHdjNwe6Zm2wGUT29zcDmkFq5FqvtY5AuQ6yhuOjXwS1vLP1ckXSJePz0TJNDITW5UmSRI/tesjvnbsq+D/NcerrOvuF0xzKkXlm/lMYq2n3EgQ7neWCCQCrKiQcY9BdhsFEqwIDAQAB", base64PublicKey)
}
