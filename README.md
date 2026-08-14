# tundra

## Commits

Commit messages are linted by [.githooks/commit-msg](.githooks/commit-msg). Enable it once per
clone:

```sh
git config core.hooksPath .githooks
```

### Convention

```
<type>[(scope)][!]: <description>
```

- `type` is one of `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`,
  `chore`, `revert`.
- `scope` is optional and lower-case, e.g. `(tun)`, `(net)`, `(deps)`.
- `!` marks a breaking change.
- `description` starts lower-case, has no trailing period, and is at least 10 characters.
- The whole subject line is at most 72 characters.

```
feat(tun): allocate the device via TUNSETIFF
fix: drop packets shorter than the ip header
chore(deps): bump golang.org/x/sys to v0.47.0
refactor(net)!: return a reader instead of a raw fd
```
