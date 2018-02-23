package filesystem

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceFile() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"path": {
				Type:        schema.TypeString,
				Description: "Path to the file to be created",
				Required:    true,
				ForceNew:    true,
			},
			"user": {
				Type:        schema.TypeString,
				Description: "File owner user name (default: current user)",
				Optional:    true,
				ForceNew:    false,
			},
			"group": {
				Type:        schema.TypeString,
				Description: "File owner group name",
				Optional:    true,
				ForceNew:    false,
			},
			"mode": {
				Type:        schema.TypeString,
				Description: "Permissions to apply to file (in octal representation, e.g. 0644)",
				Optional:    true,
				Default:     "0644",
				ForceNew:    false,
				ValidateFunc: func(i interface{}, k string) (ws []string, errors []error) {
					if _, err := strconv.ParseUint(i.(string), 8, 32); err != nil {
						errors = append(errors, fmt.Errorf("%q: invalid value", k))
					}
					return
				},
			},
			"content": {
				Type:        schema.TypeString,
				Description: "File content",
				Optional:    true,
				Default:     "",
				ForceNew:    false,
				StateFunc: func(v interface{}) string {
					return hash(v.(string))
				},
			},
		},

		Create: resourceFilesystemFileCreate,
		Read:   resourceFilesystemFileRead,
		Update: resourceFilesystemFileUpdate,
		Delete: resourceFilesystemFileDelete,
	}
}

func resourceFilesystemFileCreate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemFileCreate()")

	fileMode, _ := strconv.ParseUint(d.Get("mode").(string), 8, 32)
	d.Set("mode", fmt.Sprintf("%#o", os.FileMode(fileMode)))

	file, err := os.OpenFile(d.Get("path").(string), os.O_RDWR|os.O_CREATE, os.FileMode(fileMode))
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(d.Get("content").(string)); err != nil {
		return err
	}

	// If neither user or group attributes are set, fallback to current user uid/gid
	if d.Get("user").(string) == "" || d.Get("group").(string) == "" {
		currentUser, err := user.Current()
		if err != nil {
			return fmt.Errorf("unable to lookup current user name: %s", err)
		}
		d.Set("user", currentUser.Username)

		if d.Get("group").(string) == "" {
			currentGroup, err := user.LookupGroupId(currentUser.Gid)
			if err != nil {
				return fmt.Errorf("unable to lookup current user group name: %s", err)
			}
			d.Set("group", currentGroup.Name)
		}
	}

	u, err := user.Lookup(d.Get("user").(string))
	if err != nil {
		return fmt.Errorf("unable to lookup file owner user information: %s", err)
	}
	uid, _ := strconv.Atoi(u.Uid)

	g, err := user.LookupGroup(d.Get("group").(string))
	if err != nil {
		return fmt.Errorf("unable to lookup file owner group information: %s", err)
	}
	gid, _ := strconv.Atoi(g.Gid)

	if err := file.Chown(uid, gid); err != nil {
		return fmt.Errorf("unable to change file user/group: %s", err)
	}

	d.SetId(hash(file.Name()))

	return nil
}

func resourceFilesystemFileRead(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemFileRead()")

	fileInfo, err := os.Stat(d.Get("path").(string))
	if err != nil {
		if os.IsNotExist(err) {
			d.SetId("")
			return nil
		}

		return err
	}
	d.Set("mode", fmt.Sprintf("%#o", fileInfo.Mode()))

	fileContent, err := ioutil.ReadFile(d.Get("path").(string))
	if err != nil {
		return err
	}
	d.Set("content", hash(string(fileContent)))

	u, err := user.LookupId(fmt.Sprintf("%d", fileInfo.Sys().(*syscall.Stat_t).Uid))
	if err != nil {
		return fmt.Errorf("unable to lookup file owner user information: %s", err)
	}
	d.Set("user", u.Username)

	g, err := user.LookupGroupId(fmt.Sprintf("%d", fileInfo.Sys().(*syscall.Stat_t).Gid))
	if err != nil {
		return fmt.Errorf("unable to lookup file owner group information: %s", err)
	}
	d.Set("group", g.Name)

	return nil
}

func resourceFilesystemFileUpdate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemFileUpdate()")

	file, err := os.OpenFile(d.Get("path").(string), os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	if d.HasChange("mode") {
		fileMode, _ := strconv.ParseUint(d.Get("mode").(string), 8, 32)

		if err := file.Chmod(os.FileMode(fileMode)); err != nil {
			return err
		}
	}

	if d.HasChange("user") || d.HasChange("group") {
		u, err := user.Lookup(d.Get("user").(string))
		if err != nil {
			return fmt.Errorf("unable to lookup file owner user information: %s", err)
		}
		uid, _ := strconv.Atoi(u.Uid)

		g, err := user.LookupGroup(d.Get("group").(string))
		if err != nil {
			return fmt.Errorf("unable to lookup file owner group information: %s", err)
		}
		gid, _ := strconv.Atoi(g.Gid)

		if err := file.Chown(uid, gid); err != nil {
			return fmt.Errorf("unable to change file user/group: %s", err)
		}
	}

	if d.HasChange("content") {
		if err := file.Truncate(0); err != nil {
			return err
		}

		if _, err := file.WriteString(d.Get("content").(string)); err != nil {
			return err
		}
	}

	return nil
}

func resourceFilesystemFileDelete(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemFileDelete()")

	return os.Remove(d.Get("path").(string))
}
