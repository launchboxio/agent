# Changelog

All notable changes to this project will be documented in this file.

### [1.5.2](https://github.com/launchboxio/agent/compare/v1.5.1...v1.5.2) (2023-11-17)


### Bug Fixes

* Add pkg.crossplane.io/v1 to client scheme ([433c2ca](https://github.com/launchboxio/agent/commit/433c2ca800edc2601cf41822d07ee0971fa12108))

### [1.5.1](https://github.com/launchboxio/agent/compare/v1.5.0...v1.5.1) (2023-11-16)


### Bug Fixes

* Patch type casting and loading of project resource from event ([a4bfb74](https://github.com/launchboxio/agent/commit/a4bfb749e7062738db03e270d7802eb6cb0bc7de))

## [1.5.0](https://github.com/launchboxio/agent/compare/v1.4.3...v1.5.0) (2023-11-16)


### Features

* Remove vendor, process using manifest endpoint ([365eef6](https://github.com/launchboxio/agent/commit/365eef6de88418db790e41fd04d3804115f990fc))


### Bug Fixes

* Remove GOPRIVATE ([dbc993d](https://github.com/launchboxio/agent/commit/dbc993d88df3322e1ba4f55b5f8ba80b6371f336))
* Remove references ([81e5c7b](https://github.com/launchboxio/agent/commit/81e5c7b0734ac8236dc3cc4bd57fe6d06ac6e990))

### [1.4.3](https://github.com/launchboxio/agent/compare/v1.4.2...v1.4.3) (2023-11-14)


### Bug Fixes

* Use the appropriate event names for addons ([97da5a6](https://github.com/launchboxio/agent/commit/97da5a6eb3197847343d4bacdcb2ffd8a672d2b1))

### [1.4.2](https://github.com/launchboxio/agent/compare/v1.4.1...v1.4.2) (2023-11-14)


### Bug Fixes

* Type cast addon parameters ([95440d6](https://github.com/launchboxio/agent/commit/95440d606866fb0d3847d260b3b0a5ce2743a560))

### [1.4.1](https://github.com/launchboxio/agent/compare/v1.4.0...v1.4.1) (2023-11-14)


### Bug Fixes

* Adjust type casting of addon interface ([46191cf](https://github.com/launchboxio/agent/commit/46191cf297cde5ff2a4d92e7e6634cb97db827af))

## [1.4.0](https://github.com/launchboxio/agent/compare/v1.3.0...v1.4.0) (2023-11-14)


### Features

* Add addon watcher ([61c6f12](https://github.com/launchboxio/agent/commit/61c6f1253d6801e7fe7128c1a23ff8a8a6960e3e))
* Add watcher for addons ([6ca5585](https://github.com/launchboxio/agent/commit/6ca55854f3bd8ccc6e212a3f378a8c7e5be788ab))


### Bug Fixes

* Map a project payload to its configured addons ([20dbfeb](https://github.com/launchboxio/agent/commit/20dbfeb05690ba927aef69b7cb88068f7f2cf056))

## [1.3.0](https://github.com/launchboxio/agent/compare/v1.2.2...v1.3.0) (2023-11-09)


### Features

* Populate project Kubernetes version ([2528b33](https://github.com/launchboxio/agent/commit/2528b336fc78ff58695e486e704bbaeb7e450e2b))

### [1.2.2](https://github.com/launchboxio/agent/compare/v1.2.1...v1.2.2) (2023-11-08)


### Bug Fixes

* Adjust event processing ([55dd3dc](https://github.com/launchboxio/agent/commit/55dd3dce26dfc899574c90ba8f6f175a953952cb))

### [1.2.1](https://github.com/launchboxio/agent/compare/v1.2.0...v1.2.1) (2023-11-08)


### Bug Fixes

* Remove persistent prerun for testing ([c07b39a](https://github.com/launchboxio/agent/commit/c07b39afa800e7d989bdc53e888625ccc585fcaa))

## [1.2.0](https://github.com/launchboxio/agent/compare/v1.1.1...v1.2.0) (2023-11-08)


### Features

* Add build version output to agent binary ([81ef0be](https://github.com/launchboxio/agent/commit/81ef0beb59859113660b0e6098b4d413a4c40209))

### [1.1.1](https://github.com/launchboxio/agent/compare/v1.1.0...v1.1.1) (2023-11-08)


### Bug Fixes

* **lint:** Fix formatting ([f0482d8](https://github.com/launchboxio/agent/commit/f0482d8d037be291539b57ef070ac1e3f489ae07))

## [1.1.0](https://github.com/launchboxio/agent/compare/v1.0.3...v1.1.0) (2023-11-08)


### Features

* **eval:** Add evaluation to transmit cluster information ([9c1d7d8](https://github.com/launchboxio/agent/commit/9c1d7d82365165252bec56505761ab08e791cdc9))

### [1.0.3](https://github.com/launchboxio/agent/compare/v1.0.2...v1.0.3) (2023-11-08)


### Bug Fixes

* **ping:** Add cluster ID to request payload ([ca5a78d](https://github.com/launchboxio/agent/commit/ca5a78da518b5af722d72754422e273632a912fd))

### [1.0.2](https://github.com/launchboxio/agent/compare/v1.0.1...v1.0.2) (2023-11-08)


### Bug Fixes

* **config:** Use supported environment variables ([8b0ce9c](https://github.com/launchboxio/agent/commit/8b0ce9c901d98e5f315625473f190419c9a63dc1))

### [1.0.1](https://github.com/launchboxio/agent/compare/v1.0.0...v1.0.1) (2023-11-07)


### Bug Fixes

* **sdk:** Update agent to use launchbox-go-sdk ([cebb3fb](https://github.com/launchboxio/agent/commit/cebb3fb1617447bbf6f3bdcd1730afba86cd8939))

## 1.0.0 (2023-11-07)


### Bug Fixes

* **ci:** Add semantic-release ([dd468eb](https://github.com/launchboxio/agent/commit/dd468eb1c5bad18d4824bd78e69e72da7b04e6ff))
