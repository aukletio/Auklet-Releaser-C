# Intro

Auklet is a C/C++ profiler for IoT Linux apps. Like conventional benchtop C/C++
profilers, it is implemented as a library that you can link your program
against. Unlike benchtop profilers, it is meant to be run in production and to
continuously generate performance metrics. These metrics are periodically sent
to Auklet's backend, aggregated, analyzed, and presented via a web interface.
The backend has to be told in advance about a new release so that it is able to
make sense of incoming data.
