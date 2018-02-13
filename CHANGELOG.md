# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/) and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2018-02-13

In this release, we added breaking changes by [#18](https://github.com/mercari/go-httpdoc/pull/18). Now user can set custom asset function to each test cases. Since this added new field named `AssertFunc` to `TestCase` struct, the code which uses it without specifying field name will be broken. To migrate to new version easily, we add `NewTestCase` function. Check [#18](https://github.com/mercari/go-httpdoc/pull/18) and see how our example migrate to new `TestCase` by it.

### Added 

- Add MkdirAll if not exist output directory [#14](https://github.com/mercari/go-httpdoc/pull/14)

### Changed

- Allow to provide custom assert function to `TestCase` [#18](https://github.com/mercari/go-httpdoc/pull/18)

### Removed

- Stop support Go 1.6.x [#20](https://github.com/mercari/go-httpdoc/pull/20)


## [0.1.0] - 2017-06-01

Initial release. 

### Added 

- Fundamental features
