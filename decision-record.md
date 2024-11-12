# Decision Record 

## D-BC-001: conventinal commits for `commit message formatting`

### Status

proposed by @peterbitfly

### Context

At the moment commit messages are not formatted in a standard way.
This makes it harder to work together, e.g. when resolving `merge conflicts`, using `line blaming`, to `revert` a change or
understand the `history of the project`.

Especially to create `release notes`.

### Decision

We are using [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages.

### Consequences

1. Commit messsages `MUST` be formatted in `conventional commits` style. 
1. Every change that belongs together `SHOULD` be part of `one commit`.
1. There `MUST NOT` be `commits` like `implement review feedback` or `make linter happy`.
1. Teams (e.g. frontend, backend, mobile) `SHOULD` come up with `guidelines` regarding:
    1. `types` (teams `SHOULD` stick to the [default set](https://github.com/angular/angular/blob/22b96b9/CONTRIBUTING.md#type)) and
    1. `scopes`
1. One pull requesst `MAY` contain multiple commits.
1. Every `Pull Request` (opened by a `bitflyer`) `SHOULD` have one `commit` that has a `Footer` with an `issue number` (like BEDS-XXX).
1. Before `merging` into `staging` all commits `MUST` be `rebased` (will be enforced by project settings).

### References

1. [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
1. [Conventinal Commits Cheat Sheet](https://gist.github.com/qoomon/5dfcdf8eec66a051ecd85625518cfd13)

👇 Template: copy from here 👇

## D-BC-XXX: Short_title_of_solved_problem_and_solution

### Status

proposed | rejected | accepted | deprecated | superseded by D-XXX

### Context

What is the issue that we're seeing that is motivating this decision or change?

### Decision

What is the change that we're proposing and/or doing?

### Consequences

What becomes easier or more difficult to do because of this change?

### References

1. [Documenting architecture decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) - Michael Nygard