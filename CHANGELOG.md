# Changelog

## 0.1.0 (2023-02-24)


### Features

* add remediation strategy in api ([e4ec67d](https://github.com/padok-team/burrito/commit/e4ec67d26fffd9c9ca847c7ad90110371daa3196))
* allow ssh git clone ([#25](https://github.com/padok-team/burrito/issues/25)) ([4158c89](https://github.com/padok-team/burrito/commit/4158c898b72f1dd071d8bc7d69351f0b7a8f6dbf))
* **cache:** implement not found generic error and use it to fail fast in case cache has an issue ([#22](https://github.com/padok-team/burrito/issues/22)) ([0dcac6b](https://github.com/padok-team/burrito/commit/0dcac6b5c6fab575ec1f2926ecdc0b3d59489046))
* **controller:** expose all the current manager options as flags in cmd ([#44](https://github.com/padok-team/burrito/issues/44)) ([72b9cb7](https://github.com/padok-team/burrito/commit/72b9cb7b220779495b3362443c2781267a14668f))
* **controller:** remove cache dependency and put lock on an annotation ([84bc075](https://github.com/padok-team/burrito/commit/84bc075115bf812fa92f782ff975f97f33812b34))
* handle remediation strategy and concerning branch commit comparison ([#41](https://github.com/padok-team/burrito/issues/41)) ([d52b467](https://github.com/padok-team/burrito/commit/d52b46786983d5492585590339d9f35bb64202eb))
* improve output of kubectl get on CRDs ([#67](https://github.com/padok-team/burrito/issues/67)) ([81049d7](https://github.com/padok-team/burrito/commit/81049d79c822554c1be253abda7117fe7453d49b))
* **init:** initialize operator and both controllers ([3d5f8eb](https://github.com/padok-team/burrito/commit/3d5f8ebe9053920b4a456defd83c3b2e42b31233))
* introduce 2 new status fields and show them in kubectl get command ([#52](https://github.com/padok-team/burrito/issues/52)) ([d0867ab](https://github.com/padok-team/burrito/commit/d0867ab94b06aec9a29ad088790abb3ee6ec4890))
* make drift check and apply with custom runner code ([#2](https://github.com/padok-team/burrito/issues/2)) ([544259d](https://github.com/padok-team/burrito/commit/544259dec499117e4ce890eb32b1756d6706d362))
* **multiple:** moving some cache keys to annotations base, add kubernetes client to runner ([17b3d15](https://github.com/padok-team/burrito/commit/17b3d15b052c9cf647245b081264dd2fea099e85))
* no apply on empty plan ([#21](https://github.com/padok-team/burrito/issues/21)) ([ef767bd](https://github.com/padok-team/burrito/commit/ef767bd569f872702037144eca39b4ae75e7129f))
* **runner:** enhance runner with a kubernetes client to check layer resources ([3672cf7](https://github.com/padok-team/burrito/commit/3672cf77ea95084fc5d85f15bb16594032b21656))
* **timers:** introduce timers configuration (driftDetection, waitAction, onError) ([#20](https://github.com/padok-team/burrito/issues/20)) ([7a3b1ef](https://github.com/padok-team/burrito/commit/7a3b1efb232fcc61b95933a44f099e2d05d29fdb))
* use common remediation strategy ([8985b89](https://github.com/padok-team/burrito/commit/8985b890254c0dfbf1925836f6e8480fc1e9b79b))
* **version:** inject version at build time to runner are ran in the same version as controller ([#45](https://github.com/padok-team/burrito/issues/45)) ([950d729](https://github.com/padok-team/burrito/commit/950d7291859197821f5a1d899d5b3402fe937000))


### Bug Fixes

* check for branch inside webhook ([#43](https://github.com/padok-team/burrito/issues/43)) ([6097d33](https://github.com/padok-team/burrito/commit/6097d33f898fb78d32383d2e09fa6ae9584092a4))
* **layer:** usage of printcolumn wasn't understood correctly ([a965dca](https://github.com/padok-team/burrito/commit/a965dcab4c278440790e2c9c9510653e74b12315))
* **state:** index out of range crash ([cf55b40](https://github.com/padok-team/burrito/commit/cf55b40e885832310b932e41e0468fbf0f67f374))


### Miscellaneous Chores

* release 0.1.0 ([f6eeb62](https://github.com/padok-team/burrito/commit/f6eeb62f94382aefab0fe2870dab7f738e5abac5))
* release v0.1.0 ([81661b3](https://github.com/padok-team/burrito/commit/81661b326720e6fbaecbcce0b1b6b73b7f649258))
