### Version Changes

We follow the [semver](https://semver.org/spec/v2.0.0.html) guidelines on version changes, although reviewers may exercise their discretion on individual PRs.

* _Major_ version revs when a backwards-incompatible change is made that would break clients. (Example: new required manifest fields or new required API call, or client code changes required)
* _Minor_ version revs when backwards-compatible functionality is added that would not break clients. (Example: new optional API parameter.)
* _Patch_ version revs when a change does not affect functionality but could affect how readers interpret the spec. (Example: Substantive new diagram illustrating a previously poorly-documented protocol interaction.)
* Cosmetic changes should _not_ affect the version number. (Example: Fixing typos, reformatting docs.)

#### Tracking Versions
Rather than use a CHANGELOG or VERSION file, we ask that you use annotated `git tags` when bumping the spec Semver, and use the annotation message to describe the change.
> Example: `git tag -s 4.1.0 -m "version 4.1.0 - twiddled a doohickey"`)

Please use the raw semver when tagging - no `v4.1.0`, just `4.1.0`

A list of `git tag` versions and their annotations can be generated at will via `git tag -n`

To create a CHANGELOG file, run the following command

`git tag -n --sort=-v:refname > CHANGELOG`

## Asking Questions & Submitting Feeback

Please use GitHub issues to ask questions, submit suggestions, or otherwise provide feedback. Thank you!
