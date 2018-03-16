package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os/user"

	"github.com/facette/logger"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

type filesystemProvider struct {
	log *logger.Logger
}

var providerLogFile = "terraform-provider-filesystem.log"

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"debug": {
				Type:        schema.TypeBool,
				Description: fmt.Sprintf("Enable provider debug logging (logs to file %s)", providerLogFile),
				Optional:    true,
				Default:     false,
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"filesystem_directory": resourceDirectory(),
			"filesystem_file":      resourceFile(),
		},

		ConfigureFunc: config,
	}
}

func config(d *schema.ResourceData) (interface{}, error) {
	var (
		p            filesystemProvider
		loggerConfig logger.FileConfig
		err          error
	)

	if d.Get("debug").(bool) {
		loggerConfig = logger.FileConfig{
			Level: "debug",
			Path:  providerLogFile,
		}
	}

	if p.log, err = logger.NewLogger(loggerConfig); err != nil {
		return nil, fmt.Errorf("unable to init provider debug logger: %s", err)
	}

	return p, nil
}

func getCurrentUsername() (interface{}, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to lookup current user name: %s", err)
	}
	return currentUser.Username, nil
}

func getCurrentUserGroupname() (interface{}, error) {
	currentUser, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("unable to lookup current user name: %s", err)
	}

	currentGroup, err := user.LookupGroupId(currentUser.Gid)
	if err != nil {
		return "", fmt.Errorf("unable to lookup current user group name: %s", err)
	}

	return currentGroup.Name, nil
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
