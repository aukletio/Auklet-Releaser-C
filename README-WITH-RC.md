# Changelog

### [0.9.0-rc.1](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.9.0-rc.1)

**Implemented enhancements:**

- 1.0.0 Documentation [#44](https://github.com/ESG-USA/Auklet-Releaser-C/pull/44) ([nchoch](https://github.com/nchoch))

**DevOps changes:**

- Remove autobuild script [#46](https://github.com/ESG-USA/Auklet-Releaser-C/pull/46) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.8.1](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.8.1)

### [0.8.1-rc.1](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.8.1-rc.1)

**Fixed bugs:**

- cmd/release: various schema adjustments [#41](https://github.com/ESG-USA/Auklet-Releaser-C/pull/41) ([kdsch](https://github.com/kdsch))

## [0.8.0](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.8.0)

### [0.8.0-rc.2](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.8.0-rc.2)

**Fixed bugs:**

- cmd/release: omit version field if empty [#39](https://github.com/ESG-USA/Auklet-Releaser-C/pull/39) ([kdsch](https://github.com/kdsch))

### [0.8.0-rc.1](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.8.0-rc.1)

**Implemented enhancements:**

- APM-1593 C/C++ Releaser App Versioning Update [#36](https://github.com/ESG-USA/Auklet-Releaser-C/pull/36) ([kdsch](https://github.com/kdsch))

**DevOps changes:**

- Add gofmt hook [#35](https://github.com/ESG-USA/Auklet-Releaser-C/pull/35) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Generalize gathering of core Golang licenses [#33](https://github.com/ESG-USA/Auklet-Releaser-C/pull/33) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Fix some missing license texts [#32](https://github.com/ESG-USA/Auklet-Releaser-C/pull/32) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Fix CircleCI Docker image [#31](https://github.com/ESG-USA/Auklet-Releaser-C/pull/31) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.7.0](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.7.0)

### [0.7.0-rc.1](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.7.0-rc.1)

**Implemented enhancements:**

- License under Apache 2.0 / harvest dependency licenses [#27](https://github.com/ESG-USA/Auklet-Releaser-C/pull/27) ([rjenkinsjr](https://github.com/rjenkinsjr))

**DevOps changes:**

- Push prod branch to aukletio [#28](https://github.com/ESG-USA/Auklet-Releaser-C/pull/28) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Add WhiteSource integration [#26](https://github.com/ESG-USA/Auklet-Releaser-C/pull/26) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Fix prod PR update script [#25](https://github.com/ESG-USA/Auklet-Releaser-C/pull/25) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Fix changelog generation syntax [#24](https://github.com/ESG-USA/Auklet-Releaser-C/pull/24) ([rjenkinsjr](https://github.com/rjenkinsjr))
- TS-419: Stop using GitHub API for gathering commit lists [#23](https://github.com/ESG-USA/Auklet-Releaser-C/pull/23) ([rjenkinsjr](https://github.com/rjenkinsjr))
- TS-417: update prod release PR after QA release finishes [#22](https://github.com/ESG-USA/Auklet-Releaser-C/pull/22) ([rjenkinsjr](https://github.com/rjenkinsjr))
- APM-1329: Fix GitHub API abuse rate limits [#21](https://github.com/ESG-USA/Auklet-Releaser-C/pull/21) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Stop building for Windows [#19](https://github.com/ESG-USA/Auklet-Releaser-C/pull/19) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.6.0](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.6.0)

### [0.6.0-rc.1](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.6.0-rc.1)

**Implemented enhancements:**

- APM-1233 Add file name to release object [#13](https://github.com/ESG-USA/Auklet-Releaser-C/pull/13) ([kdsch](https://github.com/kdsch))

## [0.5.0](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.5.0)

### [0.5.0-rc.1](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.5.0-rc.1)

**Implemented enhancements:**

- APM-1192 Change the C Agent Releaser's JSON [#12](https://github.com/ESG-USA/Auklet-Releaser-C/pull/12) ([kdsch](https://github.com/kdsch))

**DevOps changes:**

- APM-1177: fix changelog generation [#10](https://github.com/ESG-USA/Auklet-Releaser-C/pull/10) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.4.0](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.4.0)

### [0.4.0-rc.3](https://github.com/ESG-USA/Auklet-Releaser-C/tree/0.4.0-rc.3)

**Implemented enhancements:**

- APM-1125 Reorganize/distribute "docs" in Auklet-Agent-C repo [#9](https://github.com/ESG-USA/Auklet-Releaser-C/pull/9) ([kdsch](https://github.com/kdsch))
- APM-1134: Hardcode BASE_URL when not built locally [#8](https://github.com/ESG-USA/Auklet-Releaser-C/pull/8) ([rjenkinsjr](https://github.com/rjenkinsjr))
- APM-1090: Separate Auklet-Profiler-C into separate repos [#2](https://github.com/ESG-USA/Auklet-Releaser-C/pull/2) ([rjenkinsjr](https://github.com/rjenkinsjr))

**Fixed bugs:**

- APM-1084 Releaser overrides default envars with empty values [#3](https://github.com/ESG-USA/Auklet-Releaser-C/pull/3) ([kdsch](https://github.com/kdsch))

**DevOps changes:**

- TS-409 mini-fix 2 [#6](https://github.com/ESG-USA/Auklet-Releaser-C/pull/6) ([rjenkinsjr](https://github.com/rjenkinsjr))
- TS-409 mini-fix [#5](https://github.com/ESG-USA/Auklet-Releaser-C/pull/5) ([rjenkinsjr](https://github.com/rjenkinsjr))
- TS-409: Do not consider PRs not merged to HEAD [#4](https://github.com/ESG-USA/Auklet-Releaser-C/pull/4) ([rjenkinsjr](https://github.com/rjenkinsjr))
