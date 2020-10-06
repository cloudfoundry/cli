A JSONry release is a tagged git sha and a GitHub release.  To cut a release:

1. Create a tag with a version number, e.g. `git tag 'v1.3.0'`
2. Push the tag, e.g. `git push --tags`
3. Use `git log --pretty=format:'- %s [%h]' HEAD...vX.X.X` to list all the commits since the last release
  - Categorize the changes into
    - Breaking Changes (requires a major version)
    - New Features (minor version)
    - Fixes (fix version)
    - Maintenance (which in general can be omitted)
4. Create a new [GitHub release](https://help.github.com/articles/creating-releases/) with the version number as the tag  (e.g. `v1.3.0`).
5. Use the changes generated in a previous step (3) as a basis for the release notes.
