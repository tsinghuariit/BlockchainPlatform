## Setting up the development environment

### Overview

Through the v0.6 release, the development environment utilized Vagrant running
an Ubuntu image, which in turn launched Docker containers as a means of
ensuring a consistent experience for developers who might be working with
varying platforms, such as MacOSX, Windows, Linux, or whatever. Advances in
Docker have enabled native support on the most popular development platforms:
MacOSX and Windows. Hence, we have reworked our build to take full advantage
of these advances. While we still maintain a Vagrant based approach that can
be used for older versions of MacOSX and Windows that Docker does not support,
we strongly encourage that the non-Vagrant development setup be used.

Note that while the Vagrant-based development setup could not be used in a
cloud context, the Docker-based build does support cloud platforms such as AWS,
Azure, Google and IBM to name a few. Please follow the instructions for Ubuntu
builds, below.

### Prerequisites
* [Git client](https://git-scm.com/downloads)
* [Go](https://golang.org/) - 1.7 or later (for releases before v1.0, 1.6 or later)
* For MacOSX, [Xcode](https://itunes.apple.com/us/app/xcode/id497799835?mt=12)
must be installed
* [Docker](https://www.docker.com/products/overview) - 1.12 or later
* [Pip](https://pip.pypa.io/en/stable/installing/)
* (MacOSX) you may need to install gnutar, as MacOSX comes with bsdtar as the
default, but the build uses some gnutar flags. You can use Homebrew to install
it as follows:

```
brew install gnu-tar --with-default-names
```

* (only if using Vagrant) - [Vagrant](https://www.vagrantup.com/) - 1.7.4 or
later
* (only if using Vagrant) - [VirtualBox](https://www.virtualbox.org/) - 5.0 or
later
* BIOS Enabled Virtualization - Varies based on hardware

- Note: The BIOS Enabled Virtualization may be within the CPU or Security
settings of the BIOS

### `pip`, `behave` and `docker-compose`

```
pip install --upgrade pip
pip install behave nose docker-compose
pip install -I flask==0.10.1 python-dateutil==2.2 pytz==2014.3 pyyaml==3.10 couchdb==1.0 flask-cors==2.0.1 requests==2.4.3 pyOpenSSL==16.2.0 sha3==0.2.1
```

### Steps

#### Set your GOPATH
Make sure you have properly setup your Host's [GOPATH environment variable](https://github.com/golang/go/wiki/GOPATH). This allows for both
building within the Host and the VM.

#### Note to Windows users

If you are running Windows, before running any `git clone` commands, run the
following command.
```
git config --get core.autocrlf
```
If `core.autocrlf` is set to `true`, you must set it to `false` by running
```
git config --global core.autocrlf false
```
If you continue with `core.autocrlf` set to `true`, the `vagrant up` command
will fail with the error:

`./setup.sh: /bin/bash^M: bad interpreter: No such file or directory`

#### Cloning the Fabric project

Since the Fabric project is a `Go` project, you'll need to clone the Fabric
repo to your $GOPATH/src directory. If your $GOPATH has multiple path
components, then you will want to use the first one. There's a little bit of
setup needed:

```
cd $GOPATH/src
mkdir -p github.com/hyperledger
cd github.com/hyperledger
```

Recall that we are using `Gerrit` for source control, which has its own internal
git repositories. Hence, we will need to clone from [Gerrit](../Gerrit/gerrit.md#Working-with-a-local-clone-of-the-repository). For
brevity, the command is as follows:
```
git clone ssh://LFID@gerrit.hyperledger.org:29418/fabric && scp -p -P 29418 LFID@gerrit.hyperledger.org:hooks/commit-msg fabric/.git/hooks/
```
**Note:** Of course, you would want to replace `LFID` with your own
[Linux Foundation ID](../Gerrit/lf-account.md).

#### Boostrapping the VM using Vagrant

If you are planning on using the Vagrant developer environment, the following
steps apply. **Again, we recommend against its use except for developers that
are limited to older versions of MacOSX and Windows that are not supported by
Docker for Mac or Windows.**

```
cd $GOPATH/src/github.com/hyperledger/fabric/devenv
vagrant up
```

Go get coffee... this will take a few minutes. Once complete, you should be able
to `ssh` into the Vagrant VM just created.

```
vagrant ssh
```

Once inside the VM, you can find the peer project under
`$GOPATH/src/github.com/hyperledger/fabric`. It is also mounted as
`/hyperledger`.

### Building the fabric

Once you have all the dependencies installed, and have cloned the repository,
you can proceed to [build and test](build.md) the fabric.

### Notes

**NOTE:** Any time you change any of the files in your local fabric directory
(under `$GOPATH/src/github.com/hyperledger/fabric`), the update will be
instantly available within the VM fabric directory.

**NOTE:** If you intend to run the development environment behind an HTTP Proxy,
you need to configure the guest so that the provisioning process may complete.
You can achieve this via the *vagrant-proxyconf* plugin. Install with
`vagrant plugin install vagrant-proxyconf` and then set the VAGRANT_HTTP_PROXY
and VAGRANT_HTTPS_PROXY environment variables *before* you execute `vagrant up`.
More details are available here: https://github.com/tmatilai/vagrant-proxyconf/

**NOTE:** The first time you run this command it may take quite a while to
complete (it could take 30 minutes or more depending on your environment) and
at times it may look like it's not doing anything. As long you don't get any
error messages just leave it alone, it's all good, it's just cranking.

**NOTE to Windows 10 Users:** There is a known problem with vagrant on Windows
10 (see [mitchellh/vagrant#6754](https://github.com/mitchellh/vagrant/issues/6754)).
If the `vagrant up` command fails it may be because you do not have the Microsoft
Visual C++ Redistributable package installed. You can download the missing package at
the following address: http://www.microsoft.com/en-us/download/details.aspx?id=8328
