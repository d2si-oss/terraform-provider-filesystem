package filesystem

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDirectory() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"path": {
				Type:        schema.TypeString,
				Description: "Path to the directory to be created",
				Required:    true,
				ForceNew:    true,
			},
			"user": {
				Type:        schema.TypeString,
				Description: "Directory owner user name (default: current user)",
				Optional:    true,
				ForceNew:    false,
				DefaultFunc: getCurrentUsername,
			},
			"group": {
				Type:        schema.TypeString,
				Description: "Directory owner group name (default: current user group)",
				Optional:    true,
				ForceNew:    false,
				DefaultFunc: getCurrentUserGroupname,
			},
			"mode": {
				Type:        schema.TypeString,
				Description: "Permissions to apply to directory (in octal representation, e.g. 0755)",
				Optional:    true,
				Default:     "0755",
				ForceNew:    false,
				ValidateFunc: func(i interface{}, k string) (ws []string, errors []error) {
					if _, err := strconv.ParseUint(i.(string), 8, 32); err != nil {
						errors = append(errors, fmt.Errorf("%q: invalid value", k))
					}
					return
				},
				StateFunc: func(v interface{}) string {
					// We serialize the permissions including 'directory mode' (e.g. `020000000755`) or else
					// the internal format will always be found different from the state format (`0755`)
					permBits, _ := strconv.ParseUint(v.(string), 8, 32)
					return fmt.Sprintf("%#o", os.ModeDir|os.FileMode(permBits))
				},
			},
			"create_parents": {
				Type:        schema.TypeBool,
				Description: "Create parent directories as needed",
				Optional:    true,
				Default:     false,
				ForceNew:    false,
			},
		},

		Create: resourceFilesystemDirectoryCreate,
		Read:   resourceFilesystemDirectoryRead,
		Update: resourceFilesystemDirectoryUpdate,
		Delete: resourceFilesystemDirectoryDelete,
	}
}

func resourceFilesystemDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemDirectoryCreate()")

	dirMode, _ := strconv.ParseUint(d.Get("mode").(string), 8, 32)

	if d.Get("create_parents").(bool) {
		if err := os.MkdirAll(d.Get("path").(string), os.FileMode(dirMode)); err != nil {
			return err
		}
	} else {
		if err := os.Mkdir(d.Get("path").(string), os.FileMode(dirMode)); err != nil {
			return err
		}
	}

	dir, err := os.OpenFile(d.Get("path").(string), os.O_RDONLY, os.FileMode(dirMode))
	if err != nil {
		return err
	}
	defer dir.Close()

	dirInfo, err := dir.Stat()
	if err != nil {
		return err
	}
	d.Set("mode", fmt.Sprintf("%#o", dirInfo.Mode()))

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

	if err := dir.Chown(uid, gid); err != nil {
		return fmt.Errorf("unable to change file user/group: %s", err)
	}

	d.SetId(hash(dir.Name()))

	return nil
}

func resourceFilesystemDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemDirectoryRead()")

	dirInfo, err := os.Stat(d.Get("path").(string))
	if err != nil {
		if os.IsNotExist(err) {
			d.SetId("")
			return nil
		}

		return err
	}
	d.Set("mode", fmt.Sprintf("%#o", dirInfo.Mode()))

	u, err := user.LookupId(fmt.Sprintf("%d", dirInfo.Sys().(*syscall.Stat_t).Uid))
	if err != nil {
		return fmt.Errorf("unable to lookup directory owner user information: %s", err)
	}
	d.Set("user", u.Username)

	g, err := user.LookupGroupId(fmt.Sprintf("%d", dirInfo.Sys().(*syscall.Stat_t).Gid))
	if err != nil {
		return fmt.Errorf("unable to lookup directory owner group information: %s", err)
	}
	d.Set("group", g.Name)

	return nil
}

func resourceFilesystemDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemDirectoryUpdate()")

	dir, err := os.OpenFile(d.Get("path").(string), os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer dir.Close()

	if d.HasChange("mode") {
		dirMode, _ := strconv.ParseUint(d.Get("mode").(string), 8, 32)

		if err := dir.Chmod(os.FileMode(dirMode)); err != nil {
			return err
		}
	}

	if d.HasChange("user") || d.HasChange("group") {
		u, err := user.Lookup(d.Get("user").(string))
		if err != nil {
			return fmt.Errorf("unable to lookup directory owner user information: %s", err)
		}
		uid, _ := strconv.Atoi(u.Uid)

		g, err := user.LookupGroup(d.Get("group").(string))
		if err != nil {
			return fmt.Errorf("unable to lookup directory owner group information: %s", err)
		}
		gid, _ := strconv.Atoi(g.Gid)

		if err := dir.Chown(uid, gid); err != nil {
			return fmt.Errorf("unable to change directory user/group: %s", err)
		}
	}

	return nil
}

func resourceFilesystemDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	p := meta.(filesystemProvider)

	p.log.Debug("calling resourceFilesystemDirectoryDelete()")

	return os.Remove(d.Get("path").(string))
}
