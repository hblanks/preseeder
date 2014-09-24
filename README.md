# preseeder

`preseeder` helps automate Ubuntu & Debian installations by
templating preseed.cfg files and serving them (and other assets) to the
Debian (or Ubuntu) installer over http.

There are three reasons you might find this useful:

1. It simplifies writing a preseed.cfg.

   The arguments in a `preseed.cfg` file are not always easy to test or
   understand. `preseeder` moves the important ones into their own
   config file and lets you render the rest, either using
   preseeder's "known good" template or your own.

2. It simplifies loading a post-installation script or other assets.

   To get a fully configured installation from boot, one typically
   needs to run a lot of stuff in the so-called `late_command` in
   `preseed.cfg`. Unfortunately, it's hardly ideal to cram it all
   into the config file. `preseeder` simplifies this by letting you
   serve an arbitrarily long `late_command` (and any other static assets).
   It will also take care of automatically adding any SSH keys to
   `/root/.ssh/authorized_keys` as part of `late_command` execution.

3. For installing a small number of servers, it's fewer dependencies
   and easier to set up than a large-scale solution.

   If you need to maintain hundreds of PXE-booted servers, something
   like Cobbler is the way to go. But, if you just want to automate
   bare-metal provisioning, this may be a good way to go.  (Of course,
   for PXE, you'll still need to configure your DHCP server and set up a
   TFTP server.)


## Setup

To install, just do:

    go install github.com/hblanks/preseeder

## Usage

See `preseeder -h`. The most important thing you'll need is a
`preseed.yaml` file. See `examples/` for two examples.

Two other really useful options are `-i` to specify RSA public
keys (typically you'll just do `-i ~/.ssh/id_rsa.pub`) and `-x`
to specify a shell script to execute after booting.

Finally, it's often helpful to serve static files from your preseeder.
Yes, you can do that, too, with `-s`.


## Quickstart

### Step 1: Configure DHCP and TFTP for network booting.

See [examples/tftpboot/README.md](examples/tftpboot/README.md) for
suggestions on setting this up, if you're not already familiar. The key
is to specify a preseed URL of http://{$YOUR_IP:8080}/preseed, or
whatever port it is you're using.

If you really don't want to network boot, you could always boot
off of some other media and type the preseed URL into the boot
prompt of every server...but what's the fun in that?

### Step 2: Install and start the preseeder

Assuming you're fine with the example, you can just do:

    go install github.com/hblanks/preseeder
    preseeder -i ~/.ssh/id_rsa.pub examples/preseed.yaml

## Part 3: boot

Configuring PXE boot is BIOS-specific, but F8 usually does the trick.

You'll be able to view a JSON dump of events at http://{$YOUR_IP}:8080/.
