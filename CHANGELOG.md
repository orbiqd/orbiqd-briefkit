# Changelog

## [0.8.0](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.7.0...v0.8.0) (2026-04-06)


### Features

* **workspace:** support execution workspaces from local directory via URI ([#58](https://github.com/orbiqd/orbiqd-briefkit/issues/58)) ([5bfd4ea](https://github.com/orbiqd/orbiqd-briefkit/commit/5bfd4ea73c06aa825ce3d05186b9c684694443b4))
* **workspace:** workspace URI support with dir:// and git+https/ssh providers ([#59](https://github.com/orbiqd/orbiqd-briefkit/issues/59)) ([1b15c65](https://github.com/orbiqd/orbiqd-briefkit/commit/1b15c6524754afdbf5166d2b8646e21213d100d3))


### Bug Fixes

* **test:** block after sending signal to prevent race in mock ([#51](https://github.com/orbiqd/orbiqd-briefkit/issues/51)) ([23a5c69](https://github.com/orbiqd/orbiqd-briefkit/commit/23a5c69a4584baa2e9d7ef1ab6073cc122db67be))

## [0.7.0](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.6.1...v0.7.0) (2026-04-05)


### Features

* **runtime:** support per-call reasoning effort override ([#48](https://github.com/orbiqd/orbiqd-briefkit/issues/48)) ([1d91c4b](https://github.com/orbiqd/orbiqd-briefkit/commit/1d91c4b473b7cfd8b07f141e990c97182443becd))

## [0.6.1](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.6.0...v0.6.1) (2026-04-05)


### Bug Fixes

* **runtime/claude:** align model validation with codex and gemini runtimes ([#46](https://github.com/orbiqd/orbiqd-briefkit/issues/46)) ([b69a288](https://github.com/orbiqd/orbiqd-briefkit/commit/b69a288f4526f725359a4668673f91b6782b662e))

## [0.6.0](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.5.2...v0.6.0) (2026-03-31)


### Features

* **config:** support per-agent execution timeout in agent config ([#43](https://github.com/orbiqd/orbiqd-briefkit/issues/43)) ([ee25998](https://github.com/orbiqd/orbiqd-briefkit/commit/ee25998468c543678ff697c6500b6203cae34536))

## [0.5.2](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.5.1...v0.5.2) (2026-02-04)


### Documentation

* restructure README for end users ([#38](https://github.com/orbiqd/orbiqd-briefkit/issues/38)) ([41c1e64](https://github.com/orbiqd/orbiqd-briefkit/commit/41c1e644e9fe8f8869a1a947ba5fb2bb8be95939))

## [0.5.1](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.5.0...v0.5.1) (2026-02-04)


### Bug Fixes

* **runtime:** handle updated Claude CLI MCP server not found message ([#39](https://github.com/orbiqd/orbiqd-briefkit/issues/39)) ([c153266](https://github.com/orbiqd/orbiqd-briefkit/commit/c1532663474d598df23aab79ef4da693a52944c5))

## [0.5.0](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.4.0...v0.5.0) (2026-02-03)


### Features

* **api:** public client interface ([#34](https://github.com/orbiqd/orbiqd-briefkit/issues/34)) ([704c6b3](https://github.com/orbiqd/orbiqd-briefkit/commit/704c6b323b5e7f4f989ef554665d40e2f4db7486))


### Code Refactoring

* **cli:** rename exec command to ask ([#31](https://github.com/orbiqd/orbiqd-briefkit/issues/31)) ([fd276b7](https://github.com/orbiqd/orbiqd-briefkit/commit/fd276b7efb90f8db3a2b4af048cf02d133106f76))

## [0.4.0](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.3.1...v0.4.0) (2026-01-31)


### Features

* **test:** Implement comprehensive Claude Code CLI mock ([#23](https://github.com/orbiqd/orbiqd-briefkit/issues/23)) ([ab9a115](https://github.com/orbiqd/orbiqd-briefkit/commit/ab9a115a8ee95e98d1a0ea65d985f02d45f4659f))
* **test:** Implement comprehensive Codex CLI mock ([#24](https://github.com/orbiqd/orbiqd-briefkit/issues/24)) ([4b405a3](https://github.com/orbiqd/orbiqd-briefkit/commit/4b405a384bc01d6f6429a73835965fe61539f9f2))


### Bug Fixes

* **ci:** use GitHub App token in release-please to trigger release workflow ([eba5265](https://github.com/orbiqd/orbiqd-briefkit/commit/eba526564db1d489374f0a5dc6cf4cf63fcb68bd))
* **cli:** preserve symlink path in ResolveExecutable ([#30](https://github.com/orbiqd/orbiqd-briefkit/issues/30)) ([abf1e26](https://github.com/orbiqd/orbiqd-briefkit/commit/abf1e26c2e7f2e44ecad70b0848f177b75f16ff8))


### Code Refactoring

* **gemini:** argument handling ([#27](https://github.com/orbiqd/orbiqd-briefkit/issues/27)) ([ad614c6](https://github.com/orbiqd/orbiqd-briefkit/commit/ad614c632d0d471ca457f98701082f0d10d80c23))

## [0.3.1](https://github.com/orbiqd/orbiqd-briefkit/compare/v0.3.0...v0.3.1) (2026-01-27)


### Bug Fixes

* **ci:** match release workflow trigger to release-please tag format ([5f0b124](https://github.com/orbiqd/orbiqd-briefkit/commit/5f0b124c4afeadbfebed1ede75aed1ff3ad5f113))
* **ci:** use simple version tags without component prefix ([7ea1a66](https://github.com/orbiqd/orbiqd-briefkit/commit/7ea1a66f1cb7e403bb6fee1e2a1c15aa5051c506))
* **codex:** prevent MCP tool timeouts at 60s ([#21](https://github.com/orbiqd/orbiqd-briefkit/issues/21)) ([09b04d7](https://github.com/orbiqd/orbiqd-briefkit/commit/09b04d7cd68de96bff5d56f7218d2a41ec3299af))

## [0.3.0](https://github.com/orbiqd/orbiqd-briefkit/compare/orbiqd-briefkit-v0.2.0...orbiqd-briefkit-v0.3.0) (2026-01-27)


### Features

* add Homebrew tap with GitHub App authentication ([#17](https://github.com/orbiqd/orbiqd-briefkit/issues/17)) ([7b939c9](https://github.com/orbiqd/orbiqd-briefkit/commit/7b939c9cd7689cc32ceef89473011992f018fb76)), closes [#5](https://github.com/orbiqd/orbiqd-briefkit/issues/5)

## [0.2.0](https://github.com/orbiqd/orbiqd-briefkit/compare/orbiqd-briefkit-v0.1.0...orbiqd-briefkit-v0.2.0) (2026-01-27)


### Features

* automate releases with release-please and GoReleaser ([#9](https://github.com/orbiqd/orbiqd-briefkit/issues/9)) ([5165c07](https://github.com/orbiqd/orbiqd-briefkit/commit/5165c07ccad7bd96a2e5929b12c53d3687c26c4e))


### Code Refactoring

* **runtime:** remove dead feature toggles ([#8](https://github.com/orbiqd/orbiqd-briefkit/issues/8)) ([63230a0](https://github.com/orbiqd/orbiqd-briefkit/commit/63230a03464f222338fd18f095c3757946cf549e))
