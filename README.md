# Auklet for C

<a href="https://www.apache.org/licenses/LICENSE-2.0" alt="Apache page link -- Apache 2.0 License"><img src="https://img.shields.io/pypi/l/auklet.svg" /></a>
<a href="https://codeclimate.com/repos/5a96cefc514d3a60340008cb/maintainability"><img src="https://api.codeclimate.com/v1/badges/fdcc057ce9f2d33d7ade/maintainability" /></a>
<a href="https://codeclimate.com/repos/5a96cefc514d3a60340008cb/test_coverage"><img src="https://api.codeclimate.com/v1/badges/fdcc057ce9f2d33d7ade/test_coverage" /></a>

This is the official C releaser for Auklet. It officially supports C
and C++, and runs on most POSIX-based operating systems (Debian, 
Ubuntu Core, Raspbian, QNX, etc).

## Features

[auklet_site]: https://app.auklet.io
[auklet_client]: https://github.com/aukletio/Auklet-Client-C
[auklet_agent]: https://github.com/aukletio/Auklet-Agent-C
[mail_auklet]: mailto:hello@auklet.io
[latest_releaser]: https://s3.amazonaws.com/auklet/releaser/latest/auklet-releaser-linux-amd64-latest

- Automatic report of unhandled exceptions
- Automatic Function performance issue reporting
- Location, system architecture, and system metrics identification for all 
issues
- Ability to define data usage restriction

## Compliance

Auklet is an edge first application performance monitor; therefore, starting 
with version 1.0.0 we maintain the following compliance levels:

- Automotive Safety Integrity Level B (ASIL B)

If there are additional compliances that your industry requires please 
contact the team at [hello@auklet.io][mail_auklet].


## Prerequisites

Before an application is released to Auklet, the Auklet library, **libauklet.a** 
needs to be integrated with the application. See the README for the [Auklet 
Agent][auklet_agent] for integration instructions.

## Quickstart

### Getting Started

1. Follow the [C/C++ Agent Quickstart Guide][auklet_agent] to integrate the 
   C/C++ agent that will monitor performance issues in your app and securely 
   transmit the information to the client

1. Create an Auklet configuration with the following environment variables 
provided by your application on the [Auklet website][auklet_site]. 

    - AUKLET_APP_ID
    - AUKLET_API_KEY

1. Download the [latest releaser][latest_releaser] to your 64-bit work 
environment and set its permissions to allow execution
 
        chmod +x release
    
1. If you do not already have a debug version of your application, you'll 
need to create a -dbg file. To create one for an application called, for 
example, "Application" create an executable in the same directory called 
Application-dbg. Application-dbg does not need to actually contain debug info.

        cp Application{,-dbg}
 
1. Initialize the application's environment variables
        
1. Then you can create a release. For an application named "Application" that
 would look like

        release Application
        

Your code is almost ready to be analyzed by Auklet! Check out the README on the 
[Auklet Client's][auklet_client] repository for
 instructions of how to run your application with Auklet. 
 
 ## Advanced Settings
 
 ### Releasing a Stripped Application
 
If you want to release a stripped executable (one without debug info), 
copy the debuggable executable before running `strip`

    cp Application{,-dbg}
    strip Application