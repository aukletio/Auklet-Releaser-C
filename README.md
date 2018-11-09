# Auklet for C

<a href="https://www.apache.org/licenses/LICENSE-2.0" alt="Apache page link -- Apache 2.0 License"><img src="https://img.shields.io/pypi/l/auklet.svg" /></a>
<a href="https://codeclimate.com/repos/5a96cefc514d3a60340008cb/maintainability"><img src="https://api.codeclimate.com/v1/badges/fdcc057ce9f2d33d7ade/maintainability" /></a>

This is the official C releaser for Auklet. It officially supports C
and C++, and runs on most POSIX-based operating systems (Debian, 
Ubuntu Core, Raspbian, QNX, etc).

## Features

[auklet_site]: https://app.auklet.io
[auklet_client]: https://github.com/aukletio/Auklet-Client-C
[auklet_agent]: https://github.com/aukletio/Auklet-Agent-C
[mail_auklet]: mailto:hello@auklet.io
[latest_releaser]: https://s3.amazonaws.com/auklet/releaser/latest/auklet-releaser-linux-amd64-latest

- Automatic crash reporting
- Automatic Function performance issue reporting
- Location, system architecture, and system metrics identification for all 
issues
- Ability to define data usage restriction

## Prerequisites

Before creating a new release, the Auklet library, **libauklet.a**  needs to 
be integrated with the application. See the README for the 
[Auklet Agent][auklet_agent] for integration instructions.

## Quickstart

### Getting Started

1. Follow the [C/C++ Agent Quickstart Guide][auklet_agent] to integrate the 
   C/C++ agent that will monitor performance issues in your app and securely 
   transmit the information to the client

1. Set the following environment variables in the environment you will be 
creating releases from (CI environment, build server, local system, etc)

    - AUKLET_APP_ID
    - AUKLET_API_KEY

1. Add the following commands to your build/CI environment
 
        curl https://s3.amazonaws.com/auklet/releaser/latest/auklet-releaser-linux-amd64-latest > release
        chmod +x release
    
1. If you do not already have a debug version of your application, you'll 
need to create a -dbg file. To create one for an application called, for 
example, "Application" create an executable in the same directory called 
Application-dbg. Application-dbg does not need to actually contain debug info.

        cp Application{,-dbg}
        
1. Then you can create a release.

        release <InsertYourApplication>
        
Your code is almost ready to be analyzed by Auklet! Check out the README on the 
[Auklet Client's][auklet_client] repository for instructions of how to run 
your application with Auklet. 
 
 ## Advanced Settings
 
 ### Releasing a Stripped Application
 
If you want to release a stripped executable (one without debug info), 
copy the debuggable executable before running `strip`

    cp Application{,-dbg}
    strip <InsertYourApplication>