# Change Log

## Upcoming Changes
### [0.2.0-rc.1](https://github.com/ESG-USA/Auklet-Profiler-C/tree/0.2.0-rc.1) (Mon Jan 29 20:25:19 2018 UTC)
**Implemented enhancements:**

- wrap.go: Add app\_id to profile struct [\#49](https://github.com/ESG-USA/Auklet-Profiler-C/pull/49) ([kdsch](https://github.com/kdsch))
- APM-947 Allow child to run while Kafka producer initializes [\#42](https://github.com/ESG-USA/Auklet-Profiler-C/pull/42) ([kdsch](https://github.com/kdsch))

**Fixed bugs:**

- APM-950 instrument: Remove all calls to exit\(\) [\#45](https://github.com/ESG-USA/Auklet-Profiler-C/pull/45) ([kdsch](https://github.com/kdsch))

**Merged pull requests:**

- Streamline Customer-Facing Documentation [\#23](https://github.com/ESG-USA/Auklet-Profiler-C/pull/23) ([kdsch](https://github.com/kdsch))

## [0.1.0](https://github.com/ESG-USA/Auklet-Profiler-C/tree/0.1.0) (Tue Jan 16 21:31:32 2018 UTC)
**Merged pull requests:**

- Docker containers for release and wrapper [\#48](https://github.com/ESG-USA/Auklet-Profiler-C/pull/48) ([shogun656](https://github.com/shogun656))

### [0.1.0-rc.2](https://github.com/ESG-USA/Auklet-Profiler-C/tree/0.1.0-rc.2) (Tue Jan  9 20:27:42 2018 UTC)
**Implemented enhancements:**

- APM-936: Populate public-facing S3 bucket with Auklet profiler binaries [\#44](https://github.com/ESG-USA/Auklet-Profiler-C/pull/44) ([rjenkinsjr](https://github.com/rjenkinsjr))
- APM-941: Add dependency management to Go [\#41](https://github.com/ESG-USA/Auklet-Profiler-C/pull/41) ([rjenkinsjr](https://github.com/rjenkinsjr))

**Merged pull requests:**

- wrap.go: Use ipify to get public IP address [\#43](https://github.com/ESG-USA/Auklet-Profiler-C/pull/43) ([kdsch](https://github.com/kdsch))
- Fix Compile-Time Errors [\#40](https://github.com/ESG-USA/Auklet-Profiler-C/pull/40) ([kdsch](https://github.com/kdsch))
- wrap.go: Add millisecond timestamp to profiles [\#39](https://github.com/ESG-USA/Auklet-Profiler-C/pull/39) ([kdsch](https://github.com/kdsch))
- Use combined realtime and virtual time profile emission [\#38](https://github.com/ESG-USA/Auklet-Profiler-C/pull/38) ([kdsch](https://github.com/kdsch))
- lib.c: Use efficient JSON marshaler [\#34](https://github.com/ESG-USA/Auklet-Profiler-C/pull/34) ([kdsch](https://github.com/kdsch))
- Decouple lib.c and wrap.go [\#33](https://github.com/ESG-USA/Auklet-Profiler-C/pull/33) ([kdsch](https://github.com/kdsch))
- APM-862 Retrieve Kafka certs from Auklet API [\#27](https://github.com/ESG-USA/Auklet-Profiler-C/pull/27) ([kdsch](https://github.com/kdsch))

### [0.1.0-rc.1](https://github.com/ESG-USA/Auklet-Profiler-C/tree/0.1.0-rc.1) (Mon Dec 11 21:37:09 2017 UTC)
**Implemented enhancements:**

- Post Release Objects to API [\#2](https://github.com/ESG-USA/Auklet-Profiler-C/pull/2) ([kdsch](https://github.com/kdsch))

**Merged pull requests:**

- wrap.go: Add default value for BASE\_URL envar [\#35](https://github.com/ESG-USA/Auklet-Profiler-C/pull/35) ([kdsch](https://github.com/kdsch))
- lib.c: Add missing underscore in stack\_trace [\#32](https://github.com/ESG-USA/Auklet-Profiler-C/pull/32) ([kdsch](https://github.com/kdsch))
- Revert "wrap.go: Use long form of timezone name" [\#30](https://github.com/ESG-USA/Auklet-Profiler-C/pull/30) ([MZein1292](https://github.com/MZein1292))
- wrap.go: Use long form of timezone name [\#29](https://github.com/ESG-USA/Auklet-Profiler-C/pull/29) ([kdsch](https://github.com/kdsch))
- APM-464: DevOps improvements [\#26](https://github.com/ESG-USA/Auklet-Profiler-C/pull/26) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Post device object to device endpoint [\#25](https://github.com/ESG-USA/Auklet-Profiler-C/pull/25) ([npalaska](https://github.com/npalaska))
- lib\_test.c: Add sample count sanity check [\#24](https://github.com/ESG-USA/Auklet-Profiler-C/pull/24) ([kdsch](https://github.com/kdsch))
- Emit stacktrace on error signal [\#22](https://github.com/ESG-USA/Auklet-Profiler-C/pull/22) ([kdsch](https://github.com/kdsch))
- QA Release: Zero Value JSON field Bug Fix, Added Tests, Exit Satuts [\#21](https://github.com/ESG-USA/Auklet-Profiler-C/pull/21) ([MZein1292](https://github.com/MZein1292))
- get cpu, memory and network data for events json [\#19](https://github.com/ESG-USA/Auklet-Profiler-C/pull/19) ([npalaska](https://github.com/npalaska))
- Remove exit\(\) calls in profiler runtime [\#18](https://github.com/ESG-USA/Auklet-Profiler-C/pull/18) ([kdsch](https://github.com/kdsch))
- Fix zero-valued JSON field bug [\#17](https://github.com/ESG-USA/Auklet-Profiler-C/pull/17) ([kdsch](https://github.com/kdsch))
- Use package github.com/satori/go.uuid for UUIDs [\#15](https://github.com/ESG-USA/Auklet-Profiler-C/pull/15) ([kdsch](https://github.com/kdsch))
- Use environment variables for configuration [\#14](https://github.com/ESG-USA/Auklet-Profiler-C/pull/14) ([kdsch](https://github.com/kdsch))
- lib\_test: Fail slowly, print test results [\#13](https://github.com/ESG-USA/Auklet-Profiler-C/pull/13) ([kdsch](https://github.com/kdsch))
- Improve Readme [\#12](https://github.com/ESG-USA/Auklet-Profiler-C/pull/12) ([kdsch](https://github.com/kdsch))
- Use Git object hashes instead of paths [\#11](https://github.com/ESG-USA/Auklet-Profiler-C/pull/11) ([kdsch](https://github.com/kdsch))
- Events, unit tests, thread support, configuration [\#10](https://github.com/ESG-USA/Auklet-Profiler-C/pull/10) ([kdsch](https://github.com/kdsch))
- Production Release: Wrapper Implementation, Release Tool Implementation [\#9](https://github.com/ESG-USA/Auklet-Profiler-C/pull/9) ([MZein1292](https://github.com/MZein1292))
- QA Release: Release Tool, Wrapper, CircleCI addition [\#8](https://github.com/ESG-USA/Auklet-Profiler-C/pull/8) ([MZein1292](https://github.com/MZein1292))
- Add Docs [\#6](https://github.com/ESG-USA/Auklet-Profiler-C/pull/6) ([kdsch](https://github.com/kdsch))
- CircleCI/Code Climate support [\#4](https://github.com/ESG-USA/Auklet-Profiler-C/pull/4) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Add wrapper, instrument [\#1](https://github.com/ESG-USA/Auklet-Profiler-C/pull/1) ([kdsch](https://github.com/kdsch))

\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*