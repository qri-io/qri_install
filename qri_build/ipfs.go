package main

// IPFSAdd adds the given file to IPFS & returns the root CID
func IPFSAdd(path string) (hash string, err error) {
	return command{
		String: "ipfs add -rQ %s",
		Tmpl: []interface{}{
			path,
		},
	}.RunStdout()
}
