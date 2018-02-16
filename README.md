# Terraform "filesystem" provider

The *filesystem* Terraform provider manages local files and directories.

⚠️ This provider has been developed for educational purposes only, **it is not suitable for production use**.

## Installation

Note: the [Go](https://golang.org/) language compiler is required for building the provider.

Clone the repository into your `$GOPATH`:

```
$ mkdir -p $GOPATH/src/github.com/d2si-oss
$ git clone https://github.com/d2si-oss/terraform-provider-filesystem $GOPATH/src/github.com/d2si-oss/terraform-provider-filesystem
```

Enter the provider directory and build the provider

```
$ cd $GOPATH/src/github.com/d2si-oss/terraform-provider-filesystem
$ make build
```

Initialize the stack using the built provider

```
$ cd /path/to/terraform/stack
$ terraform init -plugin-dir=$GOPATH/bin
$ terraform plan
```

You should be all set!

## Configuration

### Resource "directory"

* `path` (required – type string): Path to the directory to be created
* `user` (optional – type string, default to current user): Directory owner user name
* `group` (optional – type string, default to current primary group): Directory owner group name
* `mode` (optional – type string, default `"0755"`): Permissions to apply to directory (in octal representation, e.g. 0755)
* `create_parents` (optional – type bool, default `false`): Create parent directories as needed

### Resource "file"

* `path` (required – type string): Path to the file to be created
* `user` (optional – type string, default to current user): File owner user name
* `group` (optional – type string, default to current primary group): File owner group name
* `mode` (optional – type string, default `"0644"`): Permissions to apply to file (in octal representation, e.g. 0644)
* `content` (optional – type string, default `""`): File content

## Example Usage

Using the following Terraform configuration:

```
$ cat test.tf
provider "filesystem" {
  # debug = false
}

resource "filesystem_directory" "test" {
  path = "/tmp/test/dir"
  user = "marc"
  group = "admin"
  mode = "0750"
  create_parents = true
}

resource "filesystem_file" "test" {
  path = "${filesystem_directory.test.path}/file"
  user = "marc"
  group = "admin"
  mode = "0640"
  content = <<EOF

                              /\__/\
                             /`    '\
                           === 0  0 ===
                             \  --  /
                            /        \
                           /          \
                           |           |
                           \  ||  ||  /
                            \_oo__oo_/#######o

EOF
}
```

Let's start by initializing the provider:

```
 $ terraform init

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.

$ terraform apply -auto-approve
filesystem_directory.test: Creating...
  create_parents: "" => "true"
  group:          "" => "admin"
  mode:           "" => "020000000750"
  path:           "" => "/tmp/test/dir"
  user:           "" => "marc"
filesystem_directory.test: Creation complete after 0s (ID: f5aee067c62015c4883439792bf8d42139f0ee72f2802ee0d55af7bb6a676f21)
filesystem_file.test: Creating...
  content: "" => "491b3d14819e554aaa2c950d459c3e55b99690c2b132d9c681722135458416cf"
  group:   "" => "admin"
  mode:    "" => "0640"
  path:    "" => "/tmp/test/dir/file"
  user:    "" => "marc"
filesystem_file.test: Creation complete after 0s (ID: ea77e763acccecfd63e89cef535a799dfee4f9a67e606f512a03f568b5e419d2)

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
```

Check that resources have been created according to our specifications:

```
$ tree /tmp/test/
/tmp/test/
└── dir
    └── file

1 directory, 1 file

$ stat /tmp/test/dir/ /tmp/test/dir/file
  File: /tmp/test/dir/
  Size: 96        	Blocks: 0          IO Block: 4194304 directory
Device: 1000004h/16777220d	Inode: 7489559     Links: 3
Access: (0750/drwxr-x---)  Uid: (  501/    marc)   Gid: (   80/   admin)
Access: 2018-02-08 15:19:16.494977804 +0100
Modify: 2018-02-08 15:19:03.154353901 +0100
Change: 2018-02-08 15:19:03.154353901 +0100
 Birth: 2018-02-08 15:19:03.148982056 +0100
  File: /tmp/test/dir/file
  Size: 362       	Blocks: 8          IO Block: 4194304 regular file
Device: 1000004h/16777220d	Inode: 7489560     Links: 1
Access: (0640/-rw-r-----)  Uid: (  501/    marc)   Gid: (   80/   admin)
Access: 2018-02-08 15:19:03.154332797 +0100
Modify: 2018-02-08 15:19:03.154427021 +0100
Change: 2018-02-08 15:19:03.154489856 +0100
 Birth: 2018-02-08 15:19:03.154332797 +0100

$ cat /tmp/test/dir/file

                              /\__/\
                             /`    '\
                           === 0  0 ===
                             \  --  /
                            /        \
                           /          \
                           |           |
                           \  ||  ||  /
                            \_oo__oo_/#######o

```

*"Introduce a little anarchy..."*

```
$ echo meow > /tmp/test/dir/file
$ cat /tmp/test/dir/file
meow
```

We now restore our file to the desired state:

```
$ terraform plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

filesystem_directory.test: Refreshing state... (ID: f5aee067c62015c4883439792bf8d42139f0ee72f2802ee0d55af7bb6a676f21)
filesystem_file.test: Refreshing state... (ID: ea77e763acccecfd63e89cef535a799dfee4f9a67e606f512a03f568b5e419d2)

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  ~ update in-place

Terraform will perform the following actions:

  ~ filesystem_file.test
      content: "b0f0d8ff8cc965a7b70b07e0c6b4c028f132597196ae9c70c620cb9e41344106" => "491b3d14819e554aaa2c950d459c3e55b99690c2b132d9c681722135458416cf"


Plan: 0 to add, 1 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.

$ terraform apply -auto-approve
filesystem_directory.test: Refreshing state... (ID: f5aee067c62015c4883439792bf8d42139f0ee72f2802ee0d55af7bb6a676f21)
filesystem_file.test: Refreshing state... (ID: ea77e763acccecfd63e89cef535a799dfee4f9a67e606f512a03f568b5e419d2)
filesystem_file.test: Modifying... (ID: ea77e763acccecfd63e89cef535a799dfee4f9a67e606f512a03f568b5e419d2)
  content: "b0f0d8ff8cc965a7b70b07e0c6b4c028f132597196ae9c70c620cb9e41344106" => "491b3d14819e554aaa2c950d459c3e55b99690c2b132d9c681722135458416cf"
filesystem_file.test: Modifications complete after 0s (ID: ea77e763acccecfd63e89cef535a799dfee4f9a67e606f512a03f568b5e419d2)

Apply complete! Resources: 0 added, 1 changed, 0 destroyed.

$ cat /tmp/test/dir/file

                              /\__/\
                             /`    '\
                           === 0  0 ===
                             \  --  /
                            /        \
                           /          \
                           |           |
                           \  ||  ||  /
                            \_oo__oo_/#######o

```

We're done, let's clean up:

```
$ terraform destroy -force
filesystem_directory.test: Refreshing state... (ID: f5aee067c62015c4883439792bf8d42139f0ee72f2802ee0d55af7bb6a676f21)
filesystem_file.test: Refreshing state... (ID: ea77e763acccecfd63e89cef535a799dfee4f9a67e606f512a03f568b5e419d2)
filesystem_file.test: Destroying... (ID: ea77e763acccecfd63e89cef535a799dfee4f9a67e606f512a03f568b5e419d2)
filesystem_file.test: Destruction complete after 0s
filesystem_directory.test: Destroying... (ID: f5aee067c62015c4883439792bf8d42139f0ee72f2802ee0d55af7bb6a676f21)
filesystem_directory.test: Destruction complete after 0s

Destroy complete! Resources: 2 destroyed.

$ ls /tmp/test/
$
```
