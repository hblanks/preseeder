# A 5 minute guide for booting by tftp

## Part 1: DHCP

You'll need to configure at least two options:

1. next-server: the IP address of your TFTP server.

2. filename: `/pxelinux.0` or, if you've so nested the file inside
   your tftp server root, `/tftpboot/pxelinux.0`


## Part 2: tftp

Download the netboot.tar.gz of your choice. Here's Ubuntu 14.04 LTS:

    curl -O http://archive.ubuntu.com/ubuntu/dists/trusty-updates/main/installer-amd64/current/images/netboot/netboot.tar.gz

Make sure to verify your archive! Here's how you'd do it for the Ubuntu
one, assuming you already have the Ubuntu release signing key:

    curl -O http://archive.ubuntu.com/ubuntu/dists/trusty-updates/main/installer-amd64/current/images/SHA256SUMS.gpg
    curl -O http://archive.ubuntu.com/ubuntu/dists/trusty-updates/main/installer-amd64/current/images/SHA256SUMS
    gpg --verify SHA256SUMS.gpg SHA256SUMS
    grep netboot.tar.gz SHA256SUMS | sed s+./netboot/++ > NETBOOT_SHA
    shasum -a 256 -c NETBOOT_SHA

Extract your archive, replace pxelinux.cfg/default, and move it to the
tftp root:

    mkdir netboot
    tar -C netboot -xzf netboot.tar.gz
    cat > netboot/pxelinux.cfg/default <<EOF
    default auto
    label auto
    kernel ubuntu-installer/amd64/linux
    append auto=true priority=critical initrd=ubuntu-installer/amd64/initrd.gz ramdisk_size=14984 root=/dev/rd/0 rw -- nomodeset url=http://192.168.2.5:8080/preseed
    prompt 0
    timeout 0
    EOF
    
    if [ -f /private/tftpboot ] # OS X
    then
        sudo cp -r netboot/ /private/tftpboot
    else
        sudo cp -r netboot/ /srv/tftp
    fi

Finally, (install and) start the tftp server.

OS X:

    sudo launchctl load -F /System/Library/LaunchDaemons/tftp.plist
    sudo launchctl start com.apple.tftpd

Linux:

    apt-get install -y tftpd-hpa

