package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

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
		p   filesystemProvider
		err error
	)

	if d.Get("debug").(bool) {
		if p.log, err = logger.NewLogger(logger.FileConfig{
			Level: "debug",
			Path:  providerLogFile,
		}); err != nil {
			return nil, fmt.Errorf("unable to init provider debug logger: %s", err)
		}
	} else {
		if p.log, err = logger.NewLogger(); err != nil {
			return nil, fmt.Errorf("unable to init provider debug logger: %s", err)
		}
	}

	return p, nil
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
