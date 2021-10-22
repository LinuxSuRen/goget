All parameters can be in the HTTP request query or header.

| Name | Default Value | Description |
|---|---|---|
| `branch` | `master` | The branch of your desired git repository. |
| `os` | Empty | The OS for go building. Taking value from the HTTP header of `user-agent` as well. |
| `arch` | Empty | The OS for go building. |
| `upx` | `true` | Compress the binary with upx or not. |
| `useCache` | `false` | Use the local repository when it exists if this is `true` |
