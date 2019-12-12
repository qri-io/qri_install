# qri install

This repo is the entry point to the qri universe. Qri has a lot moving parts, and spans multiple repositories. The purpose of qri_install is to make it easy to pull everything together for the purposes of producing a full build and other related tasks.

Goals:

* performing full builds
* updating all repositories at once
* continuous builds & tests
* comprehensive documentation tests

## qri_build

qri_build enables an easy way to build qri targets. These include:

* electron frontend app
  * Qri.app
  * dmg for Mac OSX
  * TODO: windows installer
  * TODO: linux
* webapp
  * publicly accessible app.qri.io
  * standard fallback app (/ipns/webapp)
  * TODO: non-minified webapp
* qri backend: the command-line `qri`
* homebrew tap

TODO(dlong): Where do build output artifacts go to?

## Creating a changelog

1. Make sure you have "conventional-changelog" installed. If not, get it with `npm add -g conventional-changelog-cli`

2. `cd` to the project directory, run:
   `conventional-changelog -p angular -i CHANGELOG.md -s`

   This will auto generate a changelog against the previous version. CHANGELOG.md is the input file, and the `-s` flag indicates we should append to the beginning of the changelog, not save over the file
  
3. Draft a set of release notes, add these to the beginning of the CHANGELOG.md file, following the format that has already been established

4. Create a PR using the title `chore(release): release vX.X.X`

5. Get feedback and merge the PR

## Electron

*To build the electron Qri.app:*

`qri_build electron --frontend ${GOPATH}/src/github.com/qri-io/frontend --qri ${GOPATH}/src/github.com/qri-io/qri`

This will build a qri binary, place the binary in the correct location, and build the qri electron app.

TODO(dlong): Should we support building just the electron app without the backend? Even if not, error when --qri is not provided needs to be improved.

*To build and publish a signed Mac OSX installer:*

`qri_build electron --frontend ${GOPATH}/src/github.com/qri-io/frontend --qri ${GOPATH}/src/github.com/qri-io/qri --publish`

This builds a dmg installer including the app, signs it with developer credentials, and pushes it to github as a draft release.

Be aware that the process will need access to your keychain, you may need to input your password for each time you have to sign a different part of the application.

## Webapp

The webapp has two varieties. The "standard fallback webapp" and the "publically accessible app.qri.io".

### Standard fallback webapp

`qri_build webapp --frontend ${GOPATH}/src/github.com/qri-io/frontend`

This complies the webapp as a js blob. The command prints the api-url, usually something like http://localhost:2503

*To push this build to IPFS*

`qri_build webapp --frontend ${GOPATH}/src/github.com/qri-io/frontend --ipfs`

__How the fallback app works__

When a user runs `qri connect`, the application looks for a new version of the fallback app on ipfs. It does this by resolving the address "/ipns/webapp" (by default) to get an ipfs hash, then downloads the js blob at that hash. Note that, even though this address looks like it is using ipns, we do a normal dns lookup.

The address to resolve is in the config.yaml as `webapp.entrypointupdateaddress` and the resolved ipfs hash is saved as 'webapp.entrypointhash`.

If the fallback app is not working, make sure the config.yaml has `webapp.enabled` set to true.

### publicly accessible app.qri.io

`qri_build webapp --frontend ~/frontend --ipfs --read-only --api-url https://api.qri.io`

The webapp at app.qri.io runs with `--read-only` enabled, so that it only serves datasets, not creates them. The build flag `--api-url` sets the url used for api requests. It sets the env var `QRI_FRONTEND_BUILD_API_URL` for qri_build.

### building a non-minified webapp

Meant only for debugging purposes

Building is slow due to minification -> comment out MinifyPlugin

TODO(dlong): Elaborate on this documentation

## Qri backend command-line

```
cd ${GOPATH}/src/github.com/qri-io/qri_install
qri_build qri --qri ${GOPATH}/src/github.com/qri-io/qri \
 --templates qri_build/templates \
 --platforms darwin,linux,windows \
 --arches 386,amd64,arm
```

outputs to current directory as qri_darwin_amd64.zip, etc

TODO(dlong): Document cross-compilation
