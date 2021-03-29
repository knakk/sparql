# This fork, and what it does differently (and why)

The source of this fork, https://github.com/knakk/sparql, has a number of characterstics which make it unattractive to us (and perhaps others).

1. The licence is non-OCI, is untested in any way. Aleister Crowley is not a great source for legal policy
   1. This repo, and other repos we've forked from knakk, have been granted the MIT Licence. Relicencing seems to be within the spirit of the original licence but frankly: who even knows?
1. There is no `go.mod` file which makes versioning of dependencies flakey
1. The docs aren't great
1. The scope of the project isn't well defined
   1. Is this repo 'officially'^1 a sparql client? If so, why provide operations to allow people to bring their own `http.Request` types?
   1. Is this repo 'officially'^1 a sparql parser? If so, why bother with a client at all? And, for that matter, a query bank?
1. There are a number of code smells, such as:
   1. Under defined interfaces (or, even, useless interfaces)
   1. Naked `interface{}` types with type switches to affect code paths
   1. Lots of copy/paste and duplicated code in modules
   1. Testing is pretty much solely happy path, and not all there (especially in the repo client)
   1. Lots of stuttering and overly verbose exported function names which could be solved by splitting solutions to subpackages

Thus, this fork.


## Notes

^1: I use the word 'officially' for want of a better word. I use it to signify 'What is the main problem this repo is designed to solve?'
