# qri install

## qri_build
qri_build allows the folks at Qri to build the production versions of the Qri webapp, the qri readonly webapp, the Qri electron app, and the Qri binary!

You can use this tool to build Qri, but only members of the Qri project can access the permissions to release an official version.

### Publishing a release of the electron app
Things you need to have before you attempt to publish a release of the electron app to Qri:
- the conventional-changelog cli (`npm add -g conventional-changelog-cli`)
- Developer ID Application certificate & private key (contact a Qri member who has permission to publish a release)
- A github access token. Get this from your github account. Add the token to your environment as GH_TOKEN (export GH_TOKEN="tokenstringhere")

Steps to take before publishing a release
- update the version of Qri on the frontend in:
  - version.js
  - app/package.json
  - package.json
- navigate to your frontend directory
- run `conventional-changelog -p angular -i CHANGELOG.md -s` (this will auto generate a changelog against the previous version. CHANGELOG.md is the input file, and the `-s` flag indicates we should append to the beginning of the changelog, not save over the file)
- draft a set of release notes, add these to the beginning of the CHANGELOG.md file, following the format that has already been established
- create a pr using the title `chore(release): release vX.X.X`
- get feedback and merge the pr

To build electron and publish a release
- run `qri_build electron --frontend path/to/frontend --qri path/to/qri --publish`
  Be aware that the process will need access to your keychain, you may need to input your password for each time you have to sign a different part of the application.
  This should build a qri binary, place the binary in the correct location, build the qri electron app, sign the app, and push it to github as a draft release
- add the release notes to the draft release
- publish the release!

### Publishing a release of the Qri binary
Things you need to have before you attempt to publish a release of the electron app to Qri:
- the conventional-changelog cli (`npm add -g conventional-changelog-cli`)
- A github access token. Get this from your github account. Add the token to your environment as GH_TOKEN (export GH_TOKEN="tokenstringhere")

Steps to take before publishing a release
- update the version of Qri on the backend in:
  - p2p/p2p.go
  - lib/lib.go
  - run api tests with -u
  
