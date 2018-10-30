# Auklet for C

<a href="https://www.apache.org/licenses/LICENSE-2.0" alt="Apache page link -- Apache 2.0 License">
<img src="https://img.shields.io/pypi/l/auklet.svg" /></a>
<a href="https://codeclimate.com/repos/5a96cefc514d3a60340008cb/maintainability">
<img src="https://api.codeclimate.com/v1/badges/5a96cefc514d3a60340008cb/maintainability" /></a>
<a href="https://codeclimate.com/repos/5a96cefc514d3a60340008cb/test_coverage" alt="Test Coverage">
<img src="https://api.codeclimate.com/v1/badges/5a96cefc514d3a60340008cb/test_coverage" /></a>

Auklet is a profiler for IoT and embedded Linux apps. Like conventional 
benchtop C/C++ profilers, it is implemented as a library that you can link 
your program against. Unlike benchtop profilers, it is meant to be run in 
production and to continuously generate performance metrics. 

# Auklet Releaser

Auklet's IoT releaser (`release`) is a deploy-time command-line tool that sends
to the Auklet backend the symbol information from any program compiled with the
Auklet agent. The releaser is built to run on 64-bit POSIX and Windows systems,
and is intended for use in CI environments.

## Prerequisites

Before an application is released to Auklet, the Auklet library, **libauklet.a** 
needs to be integrated with the application. See the README for the [Auklet 
agent](https://github.com/aukletio/Auklet-Agent-C) for integration instructions.


## Setting Environment Variables

[auklet_site]: https://app.auklet.io

An Auklet configuration is defined by the following environment variables

AUKLET_APP_ID <br />
AUKLET_API_KEY

These variables are available on the [Auklet website][auklet_site] after 
creating a new application. The variables AUKLET_API_KEY and AUKLET_APP_ID 
will be different between applications, so it is suggested that they be 
defined in a separate file, .env. A set .env file should look like:

    export AUKLET_APP_ID=5171dbff-c...
    export AUKLET_API_KEY=SM49BAMCA0...
    
## Releasing to Auklet

As mentioned with the environment variables, you should create your application 
on the [Auklet website][auklet_site] prior to creating a release. After 
that is complete, follow the below instructions.

1. Download the latest releaser to your 64-bit work environment and set its 
permissions to allow execution

        curl -o release https://s3.amazonaws.com/auklet/releaser/latest/auklet-releaser-linux-amd64-latest
        chmod +x release
    
1. If you do not already have a debug version of your application, you'll 
need to create a -dbg file. To create one for an application called, for 
example, "Application" create an executable in the same directory called 
Application-dbg. Application-dbg does not need to actually contain debug info.

        cp Application{,-dbg}
        
    If you want to release a stripped executable (one without debug info), 
    copy the debuggable executable before running `strip`
    
        cp Application{,-dbg}
        strip Application
 
1. Initialize the application's environment variables

        . .env
        
1. Then you can create a release. For an application named "Application" that
 would look like

        release Application
        
## Deploying and Running

Your code is ready to be analyzed by Auklet! Check out the README on the 
[Auklet Client's](https://github.com/aukletio/Auklet-Client-C) repository for
 instructions of how to run your application with Auklet. 