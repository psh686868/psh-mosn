
## Contributing to MOSN
MOSN is released under the Apache 2.0 license, and follows a very
standard Github development process, using Github tracker for issues and
merging pull requests into master . If you would like to contribute something, 
or simply want to hack on the code this document should help you get started.

### [Sign the Contributor License Agreement](https://www.clahub.com/agreements/psh686868/psh-mosn)
Before we accept a non-trivial patch or pull request we will need you to 
sign the Contributor License Agreement. Signing the contributor’s agreement 
does not grant anyone commit rights to the main repository, but it does mean 
that we can accept your contributions, and you will get an author credit if 
we do. Active contributors might be asked to join the core team, and given 
the ability to merge pull requests.

### Code Conventions
None of these is essential for a pull request, but they will all help. 

1. Code format
  + With cli, run `goimports -w yourfile.go` and `golint yourfile.go` to format the style
  + With ide like goland, select 'Group stdlib imports', 'Move all stdlib imports in a single group', 'Move all imports in a single declaration' in Go->imports page   
  + We would check code format when run ci test, so please ensure that you have built project before you push branch.

2. Make sure all new `.go` files to have a simple doc class comment 
with at least an `author` tag identifying you, and preferably at least a 
paragraph on what the class is for.

3. Add the ASF license header comment to all new `.go` files (copy from existing files in the project)

4. Add yourself as an `author` to the `.go` files that you modify substantially (more than cosmetic changes).

5. Add some docs.

6. A few unit tests would help a lot as well — someone has to do it.

7. When writing a commit message please follow [these conventions](https://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html), if 
you are fixing an existing issue please add Fixes gh-XXXX at the end 
of the commit message (where XXXX is the issue number).

8. Ensure that code coverage does not decrease。

9. Contribute a PR as the rule of Gitflow Workflow; MOSN's version contains three digit, the first one is for compatibility; the second one is for new features and enhancement; the last one is for bug fix.