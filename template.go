package preseeder

import (
	"fmt"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"regexp"
	"text/template"
)

const defaultPreseed = `
#### Contents of the preconfiguration file (for squeeze)
### Localization
# Preseeding only locale sets language, country and locale.
d-i debian-installer/locale string {{ .Locale }}
d-i debian-installer/country string {{ .Country }}
d-i localechooser/languagelist string {{ .Language }}

# Keyboard selection.
# Disable automatic (interactive) keymap detection.
d-i console-setup/ask_detect boolean false
d-i console-setup/layoutcode string us 
d-i keyboard-configuration/layoutcode string us
d-i keyboard-configuration/xkb-keymap string us
d-i keymap string us


### Network configuration
# netcfg will choose an interface that has link if possible. This makes it
# skip displaying a list if there is more than one interface.
d-i netcfg/choose_interface select eth0

# If you have a slow dhcp server and the installer times out waiting for
# it, this might be useful.
#d-i netcfg/dhcp_timeout string 60

# Disable that annoying WEP key dialog.
d-i netcfg/wireless_wep string

# All other netcfg options (for setting a static IP or setting a
# hostname) don't seem to work, so they've been left out.


### Mirror settings
# If you select ftp, the mirror/country string does not need to be set.
d-i mirror/protocol string http
d-i mirror/country string manual
d-i mirror/http/hostname string {{ .Mirror.Hostname }}
d-i mirror/http/directory string {{ .Mirror.Directory }}
d-i mirror/http/proxy string {{ .Mirror.Proxy }}


### Clock and time zone setup
# Controls whether or not the hardware clock is set to UTC.
d-i clock-setup/utc boolean true

# You may set this to any valid setting for $TZ; see the contents of
# /usr/share/zoneinfo/ for valid values.
d-i time/zone string {{ .Timezone }}

# Controls whether to use NTP to set the clock during the install
d-i clock-setup/ntp boolean true
# NTP server to use. The default is almost always fine here.
#d-i clock-setup/ntp-server string ntp.example.com


### Partitioning
{{ .Partman }}


### Base system installation
# Configure APT to not install recommended packages by default. Use of this
# option can result in an incomplete system and should only be used by very
# experienced users.
#d-i base-installer/install-recommends boolean false

# The kernel image (meta) package to be installed; "none" can be used if no
# kernel is to be installed.
d-i base-installer/kernel/image string linux-generic


### Account setup
# Set up normal user account. Password will be locked by late_command.
d-i passwd/user-fullname string {{ .Username }}
d-i passwd/username string {{ .Username }}
d-i passwd/user-password password {{ .Password }}
d-i passwd/user-password-again password {{ .Password }}

# Skip creation of a root account (normal user account will be able to
# use sudo). The default is false; preseed this to true if you want to set
# a root password.
d-i passwd/root-login boolean false

# Set to true if you want to encrypt the first user's home directory.
d-i user-setup/encrypt-home boolean false


### Apt setup
# You can choose to install non-free and contrib software.
d-i apt-setup/non-free boolean true
d-i apt-setup/contrib boolean true


### Package selection
tasksel tasksel/first multiselect


### Boot loader installation
# This is fairly safe to set, it makes grub install automatically to the MBR
# if no other operating system is detected on the machine.
d-i grub-installer/only_debian boolean true

# This one makes grub-installer install to the MBR if it also finds some other
# OS, which is less safe as it might not be able to boot that other OS.
d-i grub-installer/with_other_os boolean true


### Finishing up the installation
# Avoid that last message about the install being complete.
d-i finish-install/reboot_in_progress note


#### Advanced options
### Running custom commands during the installation
# d-i preseeding is inherently not secure. Nothing in the installer checks
# for attempts at buffer overflows or other exploits of the values of a
# preconfiguration file like this one. Only use preconfiguration files from
# trusted locations! To drive that home, and because it's generally useful,
# here's a way to run any shell command you'd like inside the installer,
# automatically.

# This first command is run as early as possible, just after
# preseeding is read.
#d-i preseed/early_command string anna-install some-udeb
# This command is run immediately before the partitioner starts. It may be
# useful to apply dynamic partitioner preseeding that depends on the state
# of the disks (which may not be visible when preseed/early_command runs).
#d-i partman/early_command \
#       string debconf-set partman-auto/disk "$(list-devices disk | head -n1)"

# This command is run just before the install finishes, but when there is
# still a usable /target directory. You can chroot to /target and use it
# directly, or use the apt-install and in-target commands to easily install
# packages and run commands in the target system.
d-i preseed/late_command string wget -O /target/root/late_command \
    http://{{.PreseedHost}}/preseed/late_command && \
    in-target chmod +x /root/late_command && \
    in-target /root/late_command

{{.Extra}}
`

const lateCommandScript = `
#!/bin/bash
#

set -e

exec 1>/var/log/installer-late-command.log
exec 2>&1

{{ if .AuthorizedKeys }}
mkdir -p /root/.ssh/
chmod 700 /root/.ssh
cat > /root/.ssh/authorized_keys <<EOF
{{.AuthorizedKeys}}
EOF

{{ if .SSHServer }}
usermod -L {{.Username}}  # Disable password login
apt-get install -y {{.SSHServer}}
{{ end }}

{{.LateCommand}}
`

/*
 * Types
 */

type MirrorContext struct {
	Hostname  string
	Directory string
	Proxy     string
}

type PreseedContext struct {
	Locale      string
	Country     string
	Language    string
	Timezone    string
	Mirror      MirrorContext
	Username    string
	Password    string
	Partman     string
	PreseedHost string
	SSHServer   string
	Extra       string
}

type lateCommandContext struct {
	AuthorizedKeys string
	LateCommand    string
	*PreseedContext
}


/*
 * Globals
 */

var LocaleRegex *regexp.Regexp

func init() {
	LocaleRegex = regexp.MustCompile(`^([a-z]+)_([A-Z]+)(?:\.(.*))$`)
}

/*
 * Functions
 */

func validatePreseedContext(ctx *PreseedContext) error {
	submatch := LocaleRegex.FindStringSubmatch(ctx.Locale)
	if len(submatch) < 2 {
		return fmt.Errorf("failed to parse locale %v",
			ctx.Locale)
	}
	if ctx.Country == "" {
		ctx.Country = submatch[2]
	}
	if ctx.Language == "" {
		ctx.Language = submatch[1]
	}

	if ctx.Timezone == "" {
		ctx.Timezone = "UTC"
	}

	if ctx.Username == "" {
		return fmt.Errorf("username must be defined")
	}

	if ctx.Password == "" {
		return fmt.Errorf("password must be defined")
	}

	return nil
}

func ParseYaml(yamlPath string) (*PreseedContext, error) {
	data, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}

	ctx := &PreseedContext{}
	err = yaml.Unmarshal(data, ctx)
	if err != nil {
		return nil, err
	}

	err = validatePreseedContext(ctx)
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func parseTemplateString(name string, content string) *template.Template {
	t := template.New("preseed")
	_, err := t.Parse(content)
	if err != nil {
		log.Fatalf("failure parsing %s: %v", name, err)
	}
	return t
}

func GetDefaultPreseed() string {
	return defaultPreseed
}

func GetLateCommandScript() string {
	return lateCommandScript
}
