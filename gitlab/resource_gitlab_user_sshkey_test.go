package gitlab

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/crypto/ssh"
)

func TestAccGitlabUserSSHKey_basic(t *testing.T) {
	var key gitlab.SSHKey
	rInt := acctest.RandInt()
	gitlabURL, _ := url.Parse(os.Getenv("GITLAB_BASE_URL"))

	pubKey, _, _ := testMakeSSHKeyPair()
	// gitlab is rewriting the last part of the key with the username and gitlab hostname
	gitlabPubKey := fmt.Sprintf("%s foo %d (%s)", strings.TrimSpace(pubKey), rInt, gitlabURL.Host)

	pubKey2, _, _ := testMakeSSHKeyPair()
	gitlabPubKey2 := fmt.Sprintf("%s foo %d (%s)", strings.TrimSpace(pubKey2), rInt, gitlabURL.Host)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			// Create a user + sshkey
			{
				Config: testAccGitlabUserSSHKeyConfig(rInt, gitlabPubKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabUserSSHKeyExists("gitlab_user_sshkey.foo_key", &key),
					testAccCheckGitlabUserSSHKeyAttributes(&key, &testAccGitlabUserSSHKeyExpectedAttributes{
						Title: fmt.Sprintf("foo-key %d", rInt),
						Key:   gitlabPubKey,
					}),
				),
			},
			// Update the key and title
			{
				Config: testAccGitlabUserSSHKeyUpdateConfig(rInt, gitlabPubKey2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGitlabUserSSHKeyExists("gitlab_user_sshkey.foo_key", &key),
					testAccCheckGitlabUserSSHKeyAttributes(&key, &testAccGitlabUserSSHKeyExpectedAttributes{
						Title: fmt.Sprintf("key %d", rInt),
						Key:   gitlabPubKey2,
					}),
				),
			},
			{
				ResourceName:      "gitlab_user_sshkey.foo_key",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getSSHKeyImportID("gitlab_user_sshkey.foo_key"),
			},
		},
	})
}

func getSSHKeyImportID(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Resource not found: %s", resourceName)
		}

		id := rs.Primary.ID
		if id == "" {
			return "", fmt.Errorf("No SSH key ID is set")
		}
		userID := rs.Primary.Attributes["user_id"]
		if userID == "" {
			return "", fmt.Errorf("No user ID is set")
		}

		return fmt.Sprintf("%s:%s", userID, id), nil
	}
}

func testMakeSSHKeyPair() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", "", err
	}

	// generate and write private key as PEM
	var privKeyBuf strings.Builder

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	if err := pem.Encode(&privKeyBuf, privateKeyPEM); err != nil {
		return "", "", err
	}

	// generate and write public key
	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return "", "", err
	}

	var pubKeyBuf strings.Builder
	pubKeyBuf.Write(ssh.MarshalAuthorizedKey(pub))

	return pubKeyBuf.String(), privKeyBuf.String(), nil
}

func testAccCheckGitlabUserSSHKeyExists(n string, key *gitlab.SSHKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not Found: %s", n)
		}

		keyID := rs.Primary.ID
		if keyID == "" {
			return fmt.Errorf("No key ID is set")
		}

		conn := testAccProvider.Meta().(*gitlab.Client)

		id, _ := strconv.Atoi(keyID)
		userID, _ := strconv.Atoi(rs.Primary.Attributes["user_id"])

		keys, _, err := conn.Users.ListSSHKeysForUser(userID, &gitlab.ListSSHKeysForUserOptions{})
		if err != nil {
			return err
		}

		var gotKey *gitlab.SSHKey

		for _, k := range keys {
			if k.ID == id {
				gotKey = k
				break
			}
		}
		if gotKey == nil {
			return fmt.Errorf("Could not find sshkey %d for user %d", id, userID)
		}

		*key = *gotKey
		return nil
	}
}

type testAccGitlabUserSSHKeyExpectedAttributes struct {
	Title     string
	Key       string
	CreatedAt string
}

func testAccCheckGitlabUserSSHKeyAttributes(key *gitlab.SSHKey, want *testAccGitlabUserSSHKeyExpectedAttributes) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if key.Title != want.Title {
			return fmt.Errorf("got title %q; want %q", key.Title, want.Title)
		}

		k := strings.Join(strings.Split(key.Key, " ")[:2], " ")
		wk := strings.Join(strings.Split(want.Key, " ")[:2], " ")

		if k != wk {
			return fmt.Errorf("got key %q; want %q", k, wk)
		}

		return nil
	}
}

func testAccGitlabUserSSHKeyConfig(rInt int, pubKey string) string {
	return fmt.Sprintf(`
resource "gitlab_user" "foo" {
  name             = "foo %d"
  username         = "listest%d"
  password         = "test%dtt"
  email            = "listest%d@ssss.com"
  is_admin         = false
  projects_limit   = 0
  can_create_group = false
  is_external      = false
}

resource "gitlab_user_sshkey" "foo_key" {
  title = "foo-key %d"
  key = "%s"
  user_id = gitlab_user.foo.id
}
  `, rInt, rInt, rInt, rInt, rInt, pubKey)
}

func testAccGitlabUserSSHKeyUpdateConfig(rInt int, pubKey string) string {
	return fmt.Sprintf(`
resource "gitlab_user" "foo" {
  name             = "foo %d"
  username         = "listest%d"
  password         = "test%dtt"
  email            = "listest%d@ssss.com"
  is_admin         = false
  projects_limit   = 0
  can_create_group = false
  is_external      = false
}

resource "gitlab_user_sshkey" "foo_key" {
  title = "key %d"
  key = "%s"
  user_id = gitlab_user.foo.id
}
  `, rInt, rInt, rInt, rInt, rInt, pubKey)
}
