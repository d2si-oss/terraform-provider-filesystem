package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"syscall"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccFilesystemFile(t *testing.T) {
	const (
		fileCreateResource = `
resource "filesystem_file" "test" {
  path = "/tmp/testfile"
  content = "blah"
  mode = "0600"
}
`

		fileUpdateModeResource = `
resource "filesystem_file" "test" {
  path = "/tmp/testfile"
  content = "blah"
  mode = "0644"
}
`

		fileUpdateContentResource = `
resource "filesystem_file" "test" {
  path = "/tmp/testfile"
  content = "yay"
  mode = "0644"
}
`
	)

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{"filesystem": Provider()},
		Steps: []resource.TestStep{
			resource.TestStep{
				Check:  resource.ComposeAggregateTestCheckFunc(testFilesystemFileCreate),
				Config: fileCreateResource,
			},
			resource.TestStep{
				Check:  resource.ComposeAggregateTestCheckFunc(testFilesystemFileUpdateMode),
				Config: fileUpdateModeResource,
			},
			resource.TestStep{
				Check:  resource.ComposeAggregateTestCheckFunc(testFilesystemFileUpdateContent),
				Config: fileUpdateContentResource,
			},
		},
		CheckDestroy: testFilesystemFileDelete,
	})
}

func testFilesystemFileCreate(state *terraform.State) error {
	rs, ok := state.RootModule().Resources["filesystem_file.test"]
	if !ok {
		return fmt.Errorf("Not found: %s", "filesystem_file.test")
	}

	fileInfo, err := os.Stat(rs.Primary.Attributes["path"])
	if err != nil {
		return err
	}

	u, err := user.LookupId(fmt.Sprintf("%d", fileInfo.Sys().(*syscall.Stat_t).Uid))
	if err != nil {
		return fmt.Errorf("unable to lookup file owner user information: %s", err)
	}
	currentUsername, _ := getCurrentUsername()
	if u.Username != currentUsername {
		return fmt.Errorf("test file username (%q) different from expected username (%q)",
			u.Username,
			currentUsername)
	}

	g, err := user.LookupGroupId(fmt.Sprintf("%d", fileInfo.Sys().(*syscall.Stat_t).Gid))
	if err != nil {
		return fmt.Errorf("unable to lookup file owner group information: %s", err)
	}
	currentUserGroupname, _ := getCurrentUserGroupname()
	if g.Name != currentUserGroupname {
		return fmt.Errorf("test file groupname (%q) different from expected groupname (%q)",
			g.Name,
			currentUserGroupname)
	}

	if fileInfo.Mode() != os.FileMode(0600) {
		return fmt.Errorf("test file mode (%#o) different from expected mode (%#o)", fileInfo.Mode(), 0600)
	}

	fileContent, err := ioutil.ReadFile(rs.Primary.Attributes["path"])
	if err != nil {
		return err
	}
	if hash(string(fileContent)) != hash("blah") {
		return fmt.Errorf("test file content hash (%q) different from expected hash (%q)",
			hash(string(fileContent)),
			hash("blah"))
	}

	return nil
}

func testFilesystemFileUpdateMode(state *terraform.State) error {
	rs, ok := state.RootModule().Resources["filesystem_file.test"]
	if !ok {
		return fmt.Errorf("Not found: %s", "filesystem_file.test")
	}

	fileInfo, err := os.Stat(rs.Primary.Attributes["path"])
	if err != nil {
		return err
	}

	if fileInfo.Mode() != os.FileMode(0644) {
		return fmt.Errorf("test file mode (%#o) different from expected mode (%#o)", fileInfo.Mode(), 0644)
	}

	return nil
}

func testFilesystemFileUpdateContent(state *terraform.State) error {
	rs, ok := state.RootModule().Resources["filesystem_file.test"]
	if !ok {
		return fmt.Errorf("Not found: %s", "filesystem_file.test")
	}

	fileContent, err := ioutil.ReadFile(rs.Primary.Attributes["path"])
	if err != nil {
		return err
	}
	if hash(string(fileContent)) != hash("yay") {
		return fmt.Errorf("test file content hash (%q) different from expected hash (%q)",
			hash(string(fileContent)),
			hash("yay"))
	}

	return nil
}

func testFilesystemFileDelete(state *terraform.State) error {
	rs, ok := state.RootModule().Resources["filesystem_file.test"]
	if !ok {
		return fmt.Errorf("Not found: %s", "filesystem_file.test")
	}

	if _, err := os.Stat(rs.Primary.Attributes["path"]); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return fmt.Errorf("test file not deleted properly")
}
