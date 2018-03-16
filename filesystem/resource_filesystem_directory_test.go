package filesystem

import (
	"fmt"
	"os"
	"os/user"
	"regexp"
	"syscall"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccFilesystemDirectory(t *testing.T) {
	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{"filesystem": Provider()},
		Steps: []resource.TestStep{
			resource.TestStep{
				PreConfig: func() { os.Remove("/tmp/test") },
				Config: `
resource "filesystem_directory" "test" {
  path = "/tmp/test/testdir"
  mode = "0700"
}
`,
				ExpectError: regexp.MustCompile("mkdir /tmp/test/testdir: no such file or directory"),
			},
			resource.TestStep{
				Check: resource.ComposeAggregateTestCheckFunc(testFilesystemDirectoryCreateParents),
				Config: `
resource "filesystem_directory" "test" {
  path = "/tmp/test/testdir"
  mode = "0700"
  create_parents = true
}
`,
			},
			resource.TestStep{
				Check: resource.ComposeAggregateTestCheckFunc(testFilesystemDirectoryUpdateMode),
				Config: `
resource "filesystem_directory" "test" {
  path = "/tmp/test/testdir"
  mode = "0755"
}
`,
			},
		},
		CheckDestroy: testFilesystemDirectoryDelete,
	})
}

func testFilesystemDirectoryCreateParents(state *terraform.State) error {
	rs, ok := state.RootModule().Resources["filesystem_directory.test"]
	if !ok {
		return fmt.Errorf("Not found: %s", "filesystem_directory.test")
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
		return fmt.Errorf("test directory username (%q) different from expected username (%q)",
			u.Username,
			currentUsername)
	}

	g, err := user.LookupGroupId(fmt.Sprintf("%d", fileInfo.Sys().(*syscall.Stat_t).Gid))
	if err != nil {
		return fmt.Errorf("unable to lookup file owner group information: %s", err)
	}
	currentUserGroupname, _ := getCurrentUserGroupname()
	if g.Name != currentUserGroupname {
		return fmt.Errorf("test directory groupname (%q) different from expected groupname (%q)",
			g.Name,
			currentUserGroupname)
	}

	if fileInfo.Mode() != os.ModeDir|os.FileMode(0700) {
		return fmt.Errorf("test directory mode (%#o) different from expected mode (%#o)",
			fileInfo.Mode(),
			os.ModeDir|0700)
	}

	return nil
}

func testFilesystemDirectoryUpdateMode(state *terraform.State) error {
	rs, ok := state.RootModule().Resources["filesystem_directory.test"]
	if !ok {
		return fmt.Errorf("Not found: %s", "filesystem_directory.test")
	}

	fileInfo, err := os.Stat(rs.Primary.Attributes["path"])
	if err != nil {
		return err
	}

	if fileInfo.Mode() != os.ModeDir|os.FileMode(0755) {
		return fmt.Errorf("test directory mode (%#o) different from expected mode (%#o)",
			fileInfo.Mode(),
			os.ModeDir|0755)
	}

	return nil
}

func testFilesystemDirectoryDelete(state *terraform.State) error {
	rs, ok := state.RootModule().Resources["filesystem_directory.test"]
	if !ok {
		return fmt.Errorf("Not found: %s", "filesystem_directory.test")
	}

	if _, err := os.Stat(rs.Primary.Attributes["path"]); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return fmt.Errorf("test directory not deleted properly")
}
