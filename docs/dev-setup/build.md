## Building the fabric

The following instructions assume that you have already set up your [development environment](devenv.md).

To build the Fabric:

```
cd $GOPATH/src/github.com/hyperledger/fabric
make dist-clean all
```

### Running the unit tests

Use the following sequence to run all unit tests

```
cd $GOPATH/src/github.com/hyperledger/fabric
make unit-test
```

To run a specific test use the `-run RE` flag where RE is a regular expression that matches the test case name. To run tests with verbose output use the `-v` flag. For example, to run the `TestGetFoo` test case, change to the directory containing the `foo_test.go` and call/excecute

```
go test -v -run=TestGetFoo
```

### Running Node.js Unit Tests

You must also run the Node.js unit tests to insure that the Node.js client SDK is not broken by your changes. To run the Node.js unit tests, follow the instructions [here](https://github.com/hyperledger/fabric-sdk-node/README.md).

### Running Behave BDD Tests

**Note:** currently, the behave tests must be run from within in the Vagrant
environment. See the devenv setup instructions if you have not already set up
your [Vagrant environment](devenv.md#Boostrapping-the-VM-using-Vagrant).

[Behave](http://pythonhosted.org/behave/) tests will setup networks of peers with different security and consensus configurations and verify that transactions run properly. To run these tests

```
cd $GOPATH/src/github.com/hyperledger/fabric
make behave
```
Some of the Behave tests run inside Docker containers. If a test fails and you want to have the logs from the Docker containers, run the tests with this option:

```
cd $GOPATH/src/github.com/hyperledger/fabric/bddtests
behave -D logs=Y
```

## Building outside of Vagrant
It is possible to build the project and run peers outside of Vagrant. Generally
speaking, one has to 'translate' the vagrant [setup file](https://github.com/hyperledger/fabric/blob/master/devenv/setup.sh)
to the platform of your choice.

### Building on Z
To make building on Z easier and faster, [this script](https://github.com/hyperledger/fabric/tree/master/devenv/setupRHELonZ.sh) is provided (which is similar to the [setup file](https://github.com/hyperledger/fabric/blob/master/devenv/setup.sh) provided for vagrant). This script has been tested only on RHEL 7.2 and has some assumptions one might want to re-visit (firewall settings, development as root user, etc.). It is however sufficient for development in a personally-assigned VM instance.

To get started, from a freshly installed OS:
```
sudo su
yum install git
mkdir -p $HOME/git/src/github.com/hyperledger
cd $HOME/git/src/github.com/hyperledger
git clone http://gerrit.hyperledger.org/r/fabric
source fabric/devenv/setupRHELonZ.sh
```
From this point, you can proceed as described above for the Vagrant development environment.

```
cd $GOPATH/src/github.com/hyperledger/fabric
make peer unit-test behave
```

### Building on Power Platform

Development and build on Power (ppc64le) systems is done outside of vagrant as outlined [here](#building-outside-of-vagrant-). For ease of setting up the dev environment on Ubuntu, invoke [this script](https://github.com/hyperledger/fabric/tree/master/devenv/setupUbuntuOnPPC64le.sh) as root. This script has been validated on Ubuntu 16.04 and assumes certain things (like, development system has OS repositories in place, firewall setting etc) and in general can be improvised further.

To get started on Power server installed with Ubuntu, first ensure you have properly setup your Host's [GOPATH environment variable](https://github.com/golang/go/wiki/GOPATH). Then, execute the following commands to build the fabric code:

```
mkdir -p $GOPATH/src/github.com/hyperledger
cd $GOPATH/src/github.com/hyperledger
git clone http://gerrit.hyperledger.org/r/fabric
sudo ./fabric/devenv/setupUbuntuOnPPC64le.sh
cd $GOPATH/src/github.com/hyperledger/fabric
make dist-clean all
```

## Configuration

Configuration utilizes the [viper](https://github.com/spf13/viper) and [cobra](https://github.com/spf13/cobra) libraries.

There is a **core.yaml** file that contains the configuration for the peer process. Many of the configuration settings can be overridden on the command line by setting ENV variables that match the configuration setting, but by prefixing with *'CORE_'*. For example, logging level manipulation through the environment is shown below:

    CORE_PEER_LOGGING_LEVEL=CRITICAL peer

## Logging

Logging utilizes the [go-logging](https://github.com/op/go-logging) library.

The available log levels in order of increasing verbosity are: *CRITICAL | ERROR | WARNING | NOTICE | INFO | DEBUG*

See [specific logging control](https://github.com/hyperledger/fabric/blob/master/docs/Setup/logging-control.md) instructions when running the peer process.
